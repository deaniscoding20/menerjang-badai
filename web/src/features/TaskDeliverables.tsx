import { useCallback, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { useWorkspace } from '../app/WorkspaceContext';
import { useQuery } from '../lib/useQuery';
import { ApiError, api, structuredDeliverableSchema } from '../lib/api';
import type { Task } from '../lib/api';
import { Loading } from '../components/Feedback';

function DeliverableError({ error, retry }: { error: unknown; retry?: () => void }) {
 return <div className="task-error" role="alert"><strong>Permintaan deliverable gagal.</strong><p>{error instanceof Error ? error.message : 'Data belum tersedia.'} Tidak ada data pengganti yang ditampilkan.</p>{retry && <button type="button" className="button secondary" onClick={retry}>Coba lagi</button>}</div>;
}

const writableRoles = new Set(['owner', 'administrator', 'marketing_manager', 'editor']);
const editableStatuses = new Set(['draft', 'ready', 'assigned']);
function matchesStructuredSchema(raw: string) {
 try { return structuredDeliverableSchema.safeParse(JSON.parse(raw)).success; } catch { return false; }
}

export function TaskDeliverables({ task, canWrite, canRefresh = true, onChanged, beginChange, endChange }: { task: Task; canWrite: boolean; canRefresh?: boolean; onChanged: () => void; beginChange?: () => boolean; endChange?: () => void }) {
 const { session, onError } = useWorkspace();
 const load = useCallback(() => api.taskDeliverables(task.id), [task.id]);
 const query = useQuery(load);
 const [selected, setSelected] = useState('');
 const [raw, setRaw] = useState('');
 const [pending, setPending] = useState(false);
 const inFlight = useRef(false);
 const [error, setError] = useState<unknown>(null);
 const [validation, setValidation] = useState('');
 const [notice, setNotice] = useState('');
 const writable = canWrite && writableRoles.has(session.user.role) && editableStatuses.has(task.status);
 const conflict = error instanceof ApiError && error.code === 'version_conflict';
 const blocked = pending || query.loading || !!query.error || !query.data || conflict;
 function reload() {
  if (!canRefresh || pending || query.loading || inFlight.current) return;
  if (beginChange && !beginChange()) return;
  try { setError(null); setNotice(''); query.retry(); onChanged(); } finally { endChange?.(); }
 }
 async function submit(event: FormEvent) {
  event.preventDefault();
  if (!writable || blocked || inFlight.current || !query.data) return;
  const bytes = new TextEncoder().encode(raw).length;
  if (!raw.trim() || raw.includes('\0') || bytes < 1 || bytes > 131072) {
   setValidation('Isi output 1–131072 byte UTF-8, tidak boleh hanya spasi atau memuat NUL.'); return;
  }
  if (beginChange && !beginChange()) return;
  inFlight.current = true; setPending(true); setError(null); setValidation(''); setNotice('');
  try {
   const saved = await api.importTaskDeliverable(task.id, { raw_output: raw, version: query.data.version }, session.csrf_token);
   setRaw(''); setSelected(saved.id); setNotice('Revisi impor tersimpan. Ini bukan persetujuan atau bukti eksekusi.');
   query.retry(); onChanged();
  } catch (reason) { setError(reason); onError(reason); }
  finally { inFlight.current = false; setPending(false); endChange?.(); }
 }
 const items = query.data ? [...query.data.items].sort((a, b) => b.revision - a.revision) : [];
 const selectedId = items.some(item => item.id === selected) ? selected : items[0]?.id;
 return <section className="panel task-deliverables" aria-labelledby="task-deliverables-title">
  <div className="section-heading"><h2 id="task-deliverables-title">Deliverable impor manusia</h2><button type="button" className="task-text-button" disabled={!canRefresh || query.loading || pending} onClick={reload}>Segarkan deliverable</button></div>
  <p className="task-history-note">Impor manusia, bukan bukti keluaran agen. Validasi schema bukan verifikasi fakta atau persetujuan. Impor tidak menjalankan agen, menerbitkan konten, atau mengubah status task.</p>
  <p className="task-history-note">Maksimal 100 revisi terbaru, urutan terbaru dahulu; pagination belum tersedia. Revisi tersimpan tidak dapat diedit atau dihapus.</p>
  {query.loading ? <Loading label="Memuat deliverable…" /> : query.error || !query.data ? <DeliverableError error={query.error} retry={query.retry} /> : items.length === 0 ? <p className="task-history-empty">Belum ada deliverable.</p> : <>
   <ol className="task-deliverable-list">{items.map(item => <li key={item.id}>
    <button type="button" className="task-deliverable-select" aria-pressed={item.id === selectedId} onClick={() => setSelected(item.id)}><strong>Revisi {item.revision} · {item.title}</strong><span>{item.validation_status === 'validated' ? 'Schema tervalidasi' : 'Schema tidak valid'}</span></button>
    <p>{item.kind || 'Tanpa struktur valid'} · Impor manusia · Oleh {item.author_id} · <time dateTime={item.created_at}>{new Date(item.created_at).toLocaleString('id-ID')}</time></p>
   </li>)}</ol>
   {selectedId && <DeliverableDetail key={task.id + '/' + selectedId} taskId={task.id} deliverableId={selectedId} />}
  </>}
  {notice && <p className="task-success" role="status">{notice}</p>}
  {conflict ? <div className="task-error" role="alert"><strong>Versi task berubah.</strong><p>Muat ulang task dan deliverable, periksa perubahan, lalu kirim kembali secara manual. Tidak ada pengiriman ulang otomatis.</p><button type="button" className="button secondary" disabled={!canRefresh || query.loading || pending} onClick={reload}>Muat ulang task dan deliverable</button></div> : error != null && <DeliverableError error={error} />}
  {writable ? <form className="task-deliverable-form" aria-label="Impor deliverable" onSubmit={event => void submit(event)} aria-busy={pending}><fieldset disabled={blocked}>
   <label htmlFor="task-deliverable-raw-input">Output mentah untuk impor</label><textarea id="task-deliverable-raw-input" rows={8} value={raw} aria-invalid={!!validation} aria-describedby="task-deliverable-import-help task-deliverable-validation" onChange={event => setRaw(event.target.value)} />
   <p className="field-help" id="task-deliverable-import-help">Teks biasa, 1–131072 byte UTF-8; tidak boleh kosong atau memuat NUL. Teks yang tidak sesuai schema v1 tetap disimpan dengan status schema tidak valid. Server menentukan hasil validasi.</p>
   {raw && <p className="field-help">{matchesStructuredSchema(raw) ? 'Pratinjau lokal: sesuai schema v1; bukan persetujuan.' : 'Peringatan: output ini akan disimpan sebagai schema tidak valid jika tidak sesuai schema server.'}</p>}
   <p className="field-error" id="task-deliverable-validation" role={validation ? 'alert' : undefined}>{validation}</p>
   <button type="submit" className="button primary" disabled={blocked}>{pending ? 'Mengimpor…' : 'Impor revisi baru'}</button>
  </fieldset></form> : <p className="task-history-note">Akses baca saja. Impor hanya untuk peran penulis saat task draf, siap, atau ditugaskan; tidak saat brief diedit atau perubahan task diproses.</p>}
 </section>;
}

function DeliverableDetail({ taskId, deliverableId }: { taskId: string; deliverableId: string }) {
 const load = useCallback(() => api.taskDeliverable(taskId, deliverableId), [taskId, deliverableId]);
 const query = useQuery(load);
 if (query.loading) return <Loading label="Memuat revisi deliverable…" />;
 if (query.error || !query.data) return <DeliverableError error={query.error} retry={query.retry} />;
 const detail = query.data;
 return <article className="task-deliverable-detail" aria-label={'Detail revisi ' + detail.revision}>
  <h3>{detail.title}</h3>
  {detail.structured ? <div className="task-deliverable-structured">{detail.structured.sections.map((section, index) => <section key={index}><h4>{section.heading}</h4><p className="preserve-lines">{section.body}</p></section>)}{detail.structured.sources && detail.structured.sources.length > 0 && <><h4>Sumber (teks biasa)</h4><ul>{detail.structured.sources.map((source, index) => <li key={index}>{source.label} · {source.url}</li>)}</ul></>}</div> : <p>Revisi ini disimpan sebagai schema tidak valid. Hanya output mentah tersedia.</p>}
  <h4>Output mentah · revisi {detail.revision}</h4><pre className="task-deliverable-raw">{detail.raw_output}</pre>
 </article>;
}
