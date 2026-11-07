import { useCallback, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { useWorkspace } from '../app/WorkspaceContext';
import { ApiError, api, taskRevisionRequestInputSchema } from '../lib/api';
import type { Task } from '../lib/api';
import { useQuery } from '../lib/useQuery';
import { Loading } from '../components/Feedback';

type Props = { task: Task; canWrite: boolean; canRefresh?: boolean; onChanged: () => void; beginChange?: () => boolean; endChange?: () => void };
const writableRoles = new Set(['owner', 'administrator', 'marketing_manager', 'editor']);
const editableStatuses = new Set(['draft', 'ready', 'assigned']);
function RevisionRequestError({ error }: { error: unknown }) {
 return <div className="task-error" role="alert"><strong>Permintaan revisi gagal.</strong><p>{error instanceof Error ? error.message : 'Data belum tersedia.'} Tidak ada data pengganti yang ditampilkan.</p></div>;
}
export function TaskRevisionRequests({ task, canWrite, canRefresh = true, onChanged, beginChange, endChange }: Props) {
 const { session, onError } = useWorkspace();
 const load = useCallback(() => api.taskRevisionRequests(task.id), [task.id]);
 const query = useQuery(load);
 const [composing, setComposing] = useState(false);
 const loadTargets = useCallback(() => composing ? api.taskDeliverables(task.id) : Promise.resolve(null), [task.id, composing]);
 const targets = useQuery(loadTargets);
 const [selected, setSelected] = useState('');
 const [body, setBody] = useState('');
 const [pending, setPending] = useState(false);
 const [notice, setNotice] = useState('');
 const [error, setError] = useState<unknown>(null);
 const [validation, setValidation] = useState('');
 const inFlight = useRef(false);
 const conflict = error instanceof ApiError && (error.status === 409 || error.code === 'version_conflict');
 const mismatch = !!query.data && !!targets.data && query.data.version !== targets.data.version;
 const selectedAvailable = !!targets.data?.items.some(item => item.id === selected);
 const writable = canWrite && writableRoles.has(session.user.role) && editableStatuses.has(task.status);
 const blocked = pending || query.loading || !!query.error || !query.data || targets.loading || !!targets.error || !targets.data || mismatch || conflict;
 const refreshBlocked = !canRefresh || pending || query.loading || (composing && targets.loading);
 function reload() {
  if (refreshBlocked || inFlight.current || (beginChange && !beginChange())) return;
  try { setError(null); setNotice(''); query.retry(); if (composing) targets.retry(); onChanged(); } finally { endChange?.(); }
 }
 async function submit(event: FormEvent) {
  event.preventDefault();
  if (!writable || blocked || inFlight.current || !query.data || !selectedAvailable) return;
  const parsed = taskRevisionRequestInputSchema.safeParse({ deliverable_id: selected, body, version: query.data.version });
  if (!parsed.success) { setValidation('Isi catatan 1–4000 karakter Unicode setelah pemangkasan spasi, tanpa NUL.'); return; }
  if (beginChange && !beginChange()) return;
  inFlight.current = true; setPending(true); setError(null); setValidation(''); setNotice('');
  try {
   await api.requestTaskRevision(task.id, parsed.data, session.csrf_token);
   setBody(''); setNotice('Permintaan revisi tersimpan. Ini bukan persetujuan atau eksekusi.');
   query.retry(); targets.retry(); onChanged();
  } catch (reason) { setError(reason); onError(reason); }
  finally { inFlight.current = false; setPending(false); endChange?.(); }
 }
 return <section className="panel task-revision-requests" aria-labelledby="task-revision-requests-title">
  <div className="section-heading"><h2 id="task-revision-requests-title">Permintaan revisi manusia</h2><button type="button" className="task-text-button" disabled={refreshBlocked} onClick={reload}>Segarkan permintaan revisi</button></div>
  <p className="task-history-note">Catatan manusia ini tidak menjalankan agen, mengulang eksekusi, menyetujui output, atau mengubah status task.</p>
  <p className="task-history-note">Maksimal 100 catatan terbaru, urutan terbaru dahulu; pagination belum tersedia. Catatan tersimpan tidak dapat diedit atau dihapus.</p>
  {query.loading ? <Loading label="Memuat permintaan revisi…" /> : query.error || !query.data ? <RevisionRequestError error={query.error} /> : query.data.items.length === 0 ? <p className="task-history-empty">Belum ada permintaan revisi.</p> : <ol className="task-revision-request-list">{[...query.data.items].sort((a, b) => b.created_at.localeCompare(a.created_at)).map(item => <li key={item.id}>
   <header><strong>Deliverable {item.deliverable_id} · Revisi {item.deliverable_revision}</strong><time dateTime={item.created_at}>{new Date(item.created_at).toLocaleString('id-ID')}</time></header>
   <p>Oleh {item.author_id}</p><p className="preserve-lines">{item.body}</p>
  </li>)}</ol>}
  {notice && <p className="task-success" role="status">{notice}</p>}
  {conflict ? <div className="task-error" role="alert"><strong>Versi task berubah.</strong><p>Muat ulang task, deliverable, dan catatan; periksa perubahan lalu kirim kembali secara manual. Tidak ada pengiriman ulang otomatis.</p><button type="button" className="button secondary" disabled={refreshBlocked} onClick={reload}>Muat ulang task dan permintaan revisi</button></div> : error != null && <RevisionRequestError error={error} />}
  {composing && targets.error != null && <RevisionRequestError error={targets.error} />}
  {composing && targets.loading && <Loading label="Memuat pilihan revisi deliverable…" />}
  {composing && !targets.loading && !targets.error && targets.data?.items.length === 0 && <p className="task-history-note">Tidak ada revisi deliverable dalam snapshot ini.</p>}
  {composing && !query.loading && !targets.loading && mismatch && <p className="task-history-note" role="alert">Snapshot permintaan dan deliverable berbeda. Segarkan kedua snapshot sebelum mengirim; tidak memakai versi task sebagai pengganti.</p>}
  {writable && !composing && <button type="button" className="button secondary task-revision-compose" onClick={() => { targets.retry(); setComposing(true); }}>Tulis permintaan revisi</button>}
  {writable && composing && <form className="task-revision-request-form" aria-label="Minta revisi deliverable" onSubmit={event => void submit(event)} aria-busy={pending}><fieldset disabled={blocked}>
   <label htmlFor="task-revision-target">Revisi deliverable tujuan</label><select id="task-revision-target" value={selected} onChange={event => setSelected(event.target.value)}><option value="">Pilih revisi secara eksplisit</option>{selected && !selectedAvailable && <option value={selected} disabled>Pilihan sebelumnya {selected} · tidak tersedia dalam snapshot</option>}{targets.data?.items.map(item => <option key={item.id} value={item.id}>Revisi {item.revision} · {item.title} · {item.validation_status === 'invalid' ? 'Schema tidak valid' : 'Schema tervalidasi'}</option>)}</select>
   <label htmlFor="task-revision-body">Catatan permintaan revisi</label><textarea id="task-revision-body" rows={5} value={body} aria-invalid={!!validation} aria-describedby="task-revision-help task-revision-validation" onChange={event => setBody(event.target.value)} />
   <p className="field-help" id="task-revision-help">Teks biasa, 1–4000 karakter Unicode setelah pemangkasan spasi, tanpa NUL. Pilihan terbatas pada 100 revisi terbaru; tidak otomatis memilih revisi terbaru.</p>
   <p className="field-error" id="task-revision-validation" role={validation ? 'alert' : undefined}>{validation}</p>
   <button type="submit" className="button primary" disabled={blocked || !selectedAvailable}>{pending ? 'Menyimpan…' : 'Kirim permintaan revisi'}</button>
  </fieldset></form>}
 </section>;
}
