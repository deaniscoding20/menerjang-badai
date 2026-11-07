import { useCallback, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { useWorkspace } from '../app/WorkspaceContext';
import { Loading } from '../components/Feedback';
import { ApiError, api } from '../lib/api';
import type { Task, TaskStatus } from '../lib/api';
import { useQuery } from '../lib/useQuery';

const writableRoles = new Set(['owner', 'administrator', 'marketing_manager', 'editor']);
const editableStatuses = new Set<TaskStatus>(['draft', 'ready', 'assigned']);
const statusLabels: Record<TaskStatus, string> = {
 draft: 'Draf', ready: 'Siap', assigned: 'Ditugaskan', queued: 'Antrean', running: 'Berjalan',
 waiting_input: 'Menunggu input', waiting_approval: 'Menunggu persetujuan',
 completed: 'Selesai', failed: 'Gagal', cancelled: 'Dibatalkan',
};

function DependencyError({ error, retry }: { error: unknown; retry?: () => void }) {
 return <div className="task-error" role="alert"><strong>Permintaan dependensi gagal.</strong><p>{error instanceof Error ? error.message : 'Data belum tersedia.'}</p>{retry && <button type="button" className="button secondary" onClick={retry}>Coba lagi</button>}</div>;
}

export function TaskDependencies({ task, canWrite, canRefresh = true, onChanged, beginChange, endChange }: { task: Task; canWrite: boolean; canRefresh?: boolean; onChanged: () => void; beginChange?: () => boolean; endChange?: () => void }) {
 const { session, onError } = useWorkspace();
 const load = useCallback(() => api.taskDependencies(task.id), [task.id]);
 const query = useQuery(load);
 const [pending, setPending] = useState(false);
 const inFlight = useRef(false);
 const [error, setError] = useState<unknown>(null);
 const [notice, setNotice] = useState('');
 const writable = canWrite && writableRoles.has(session.user.role) && editableStatuses.has(task.status);
 const conflict = error instanceof ApiError && error.code === 'version_conflict';
 const disabled = pending || query.loading || !!query.error || !query.data || conflict;
 function reload() {
  if (!canRefresh || pending || query.loading || inFlight.current) return;
  if (beginChange && !beginChange()) return;
  try { setError(null); setNotice(''); query.retry(); onChanged(); } finally { endChange?.(); }
 }
 async function mutate(dependencyId: string, method: 'add' | 'remove') {
  if (!writable || disabled || inFlight.current || !query.data) return;
  if (beginChange && !beginChange()) return;
  inFlight.current = true;
  setPending(true); setError(null); setNotice('');
  try {
   const input = { dependency_id: dependencyId, version: query.data.version };
   if (method === 'add') await api.addTaskDependency(task.id, input, session.csrf_token);
   else await api.removeTaskDependency(task.id, input, session.csrf_token);
   setNotice('Dependensi tersimpan. Tidak ada perintah eksekusi yang dikirim.');
   query.retry(); onChanged();
  } catch (reason) { setError(reason); onError(reason); }
  finally { inFlight.current = false; setPending(false); endChange?.(); }
 }
 return <section className="panel task-dependencies" aria-labelledby="task-dependencies-title" aria-busy={pending}>
  <div className="section-heading"><h2 id="task-dependencies-title">Dependensi task</h2><button className="task-text-button" disabled={!canRefresh || query.loading || pending} onClick={reload}>Segarkan dependensi</button></div>
  <p className="task-history-note">Hanya prasyarat berstatus selesai yang terpenuhi. Dependensi tidak menjalankan agen atau mengubah status task secara otomatis.</p>
  {query.loading ? <Loading label="Memuat dependensi…" /> : query.error || !query.data ? <DependencyError error={query.error} retry={query.retry} /> : query.data.items.length === 0 ? <p className="task-history-empty">Belum ada dependensi.</p> : <ul className="task-dependency-list">{query.data.items.map(dependency => <li key={dependency.task_id}>
   <div><Link to={'/tasks/' + encodeURIComponent(dependency.task_id)}>{dependency.title}</Link><p>{statusLabels[dependency.status]} · <span>{dependency.status === 'completed' ? 'Prasyarat terpenuhi' : 'Prasyarat belum terpenuhi'}</span></p></div>
   {writable && <button type="button" className="button secondary" disabled={disabled} aria-label={'Hapus dependensi ' + dependency.title} onClick={() => void mutate(dependency.task_id, 'remove')}>Hapus</button>}
  </li>)}</ul>}
  {notice && <p className="task-success" role="status">{notice}</p>}
  {conflict ? <div className="task-error" role="alert"><strong>Versi task berubah.</strong><p>Muat ulang task dan dependensi, periksa perubahan, lalu coba lagi secara manual. Tidak ada pengiriman ulang otomatis.</p><button type="button" className="button secondary" disabled={!canRefresh || query.loading || pending} onClick={reload}>Muat ulang task dan dependensi</button></div> : error != null && <DependencyError error={error} />}
  {writable ? <DependencyPicker task={task} existingIds={query.data?.items.map(item => item.task_id) ?? []} disabled={disabled} onAdd={id => mutate(id, 'add')} /> : <p className="task-history-note">Dependensi hanya dapat diubah oleh peran penulis saat task draf, siap, atau ditugaskan; tidak saat brief sedang diedit.</p>}
 </section>;
}

function DependencyPicker({ task, existingIds, disabled, onAdd }: { task: Task; existingIds: string[]; disabled: boolean; onAdd: (id: string) => Promise<void> }) {
 const load = useCallback(() => api.tasks(task.brand_id), [task.brand_id]);
 const candidates = useQuery(load);
 const [selected, setSelected] = useState('');
 const items = candidates.data?.items.filter(item => item.brand_id === task.brand_id && item.id !== task.id && !existingIds.includes(item.id)) ?? [];
 const value = items.some(item => item.id === selected) ? selected : '';
 const blocked = disabled || candidates.loading || !!candidates.error || !candidates.data;
 function submit(event: FormEvent) {
  event.preventDefault();
  if (blocked || !value) return;
  void onAdd(value);
 }
 return <form className="task-dependency-form" aria-label="Tambah dependensi" onSubmit={submit}>
  <fieldset disabled={blocked}><label htmlFor="task-dependency-candidate">Tambah prasyarat</label><select id="task-dependency-candidate" value={value} aria-describedby="task-dependency-candidate-help" onChange={event => setSelected(event.target.value)}><option value="">Pilih task satu brand</option>{items.map(item => <option key={item.id} value={item.id}>{item.title} · {statusLabels[item.status]}</option>)}</select><button type="submit" className="button primary" disabled={blocked || !value}>Tambah dependensi</button></fieldset>
  <p className="field-help" id="task-dependency-candidate-help">Pilihan terbatas pada maksimal 500 task terbaru dari brand ini; bukan daftar lengkap. Server tetap memeriksa izin dan siklus dependensi.</p>
  {candidates.loading ? <p className="field-help">Memuat pilihan task…</p> : candidates.error ? <DependencyError error={candidates.error} retry={candidates.retry} /> : items.length === 0 && <p className="field-help">Tidak ada kandidat lain dalam daftar terbatas ini.</p>}
 </form>;
}
