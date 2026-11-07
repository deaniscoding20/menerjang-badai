import { useCallback, useEffect, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { ArrowLeft, ArrowUpRight, Check, Clock3, Columns3, List, ListTodo, LockKeyhole, Plus, RotateCw, Search, X } from 'lucide-react';
import { useWorkspace } from '../app/WorkspaceContext';
import { Loading, PageHeading, StatusBadge } from '../components/Feedback';
import { ApiError, api, taskInputSchema } from '../lib/api';
import type { Brand, Task, TaskInput, TaskPriority, TaskStatus, TaskUpdate } from '../lib/api';
import { useQuery } from '../lib/useQuery';
import './Tasks.css';
import { TaskComments } from './TaskComments';
import { TaskDependencies } from './TaskDependencies';
import { TaskDeliverables } from './TaskDeliverables';
import { TaskRevisionRequests } from './TaskRevisionRequests';

const loadBrands = () => api.brands();
const loadAgents = () => api.agents();
const writableRoles = new Set(['owner', 'administrator', 'marketing_manager', 'editor']);
const terminalStatuses = new Set<TaskStatus>(['completed', 'failed', 'cancelled']);
const editableStatuses = new Set<TaskStatus>(['draft', 'ready', 'assigned']);
const statusLabels: Record<TaskStatus, string> = {
 draft: 'Draf', ready: 'Siap', assigned: 'Ditugaskan', queued: 'Antrean', running: 'Berjalan',
 waiting_input: 'Menunggu input', waiting_approval: 'Menunggu persetujuan',
 completed: 'Selesai', failed: 'Gagal', cancelled: 'Dibatalkan',
};
const priorityLabels: Record<TaskPriority, string> = { low: 'Rendah', normal: 'Normal', high: 'Tinggi', urgent: 'Mendesak' };
const boardColumns: { label: string; statuses: TaskStatus[] }[] = [
 { label: 'Draf & persiapan', statuses: ['draft', 'ready', 'assigned'] },
 { label: 'Dalam proses', statuses: ['queued', 'running'] },
 { label: 'Menunggu', statuses: ['waiting_input', 'waiting_approval'] },
 { label: 'Ditutup', statuses: ['completed', 'failed', 'cancelled'] },
];

function dateLabel(value: string) {
 const date = new Date(value);
 return Number.isNaN(date.getTime()) ? value : date.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
}
function taskUrl(id: string) { return '/tasks/' + encodeURIComponent(id); }
function TaskStatusBadge({ status }: { status: TaskStatus }) {
 return <StatusBadge tone={status === 'completed' || status === 'ready' ? 'teal' : status === 'failed' || status === 'waiting_approval' ? 'coral' : 'muted'}>{statusLabels[status]}</StatusBadge>;
}
function Priority({ priority }: { priority: TaskPriority }) {
 return <span className={`task-priority task-priority-${priority}`}><span aria-hidden="true" />{priorityLabels[priority]}</span>;
}
function TaskError({ error, retry }: { error: unknown; retry?: () => void }) {
 const unavailable = error instanceof ApiError && error.status === 404;
 return <div className="task-error" role="alert"><strong>{unavailable ? 'Endpoint atau task belum tersedia.' : 'Permintaan task gagal.'}</strong><p>{error instanceof Error ? error.message : 'Data tidak dapat dimuat.'} Tidak ada data contoh yang ditampilkan.</p>{retry && <button type="button" className="button secondary" onClick={retry}><RotateCw size={15} aria-hidden="true" />Coba lagi</button>}</div>;
}
function RuntimeNotice() {
 return <aside className="task-runtime" aria-label="Status eksekusi"><LockKeyhole size={20} aria-hidden="true" /><div><strong>Kelola pekerjaan. Eksekusi tetap terkunci.</strong><p>Runtime dispatch belum diverifikasi; bridge eksekusi belum diaktifkan. Status siap atau penugasan manual tidak menjalankan agen. Draf tetap dapat dibuat dan dikelola.</p></div><button className="button secondary" disabled title="Memerlukan bridge runtime yang terverifikasi">Jalankan task <LockKeyhole size={14} aria-hidden="true" /></button></aside>;
}

export function TasksPage() {
 const { session } = useWorkspace();
 const [params, setParams] = useSearchParams();
 const brandId = params.get('brand_id') || '';
 const brands = useQuery(loadBrands);
 const [search, setSearch] = useState('');
 const [status, setStatus] = useState('all');
 const [view, setView] = useState<'board' | 'list'>('board');
 const canWrite = writableRoles.has(session.user.role);
 return <div className="tasks-page"><PageHeading eyebrow="RUANG KERJA / TASKS" title="Pekerjaan yang punya arah." description="Susun brief, tentukan prioritas, dan telusuri perubahan task dari data server." action={canWrite ? <Link className="button primary" to={'/tasks/new' + (brandId ? '?brand_id=' + encodeURIComponent(brandId) : '')}>Buat task <Plus size={17} aria-hidden="true" /></Link> : undefined} /><RuntimeNotice />
  <section className="panel tasks-workspace" aria-label="Daftar pekerjaan"><div className="task-toolbar"><div className="search-field"><Search size={17} aria-hidden="true" /><input aria-label="Cari task" placeholder="Cari judul atau objektif…" value={search} onChange={event => setSearch(event.target.value)} /></div>
   <label className="task-filter">Brand<select value={brandId} disabled={brands.loading || !!brands.error} onChange={event => { const next = new URLSearchParams(params); if (event.target.value) next.set('brand_id', event.target.value); else next.delete('brand_id'); setParams(next); }}><option value="">Semua brand</option>{brandId && !brands.data?.items.some(brand => brand.id === brandId) && <option value={brandId}>Brand terpilih ({brandId})</option>}{brands.data?.items.map(brand => <option key={brand.id} value={brand.id}>{brand.name}{brand.status === 'archived' ? ' · Diarsipkan' : ''}</option>)}</select></label>
   <label className="task-filter">Status<select value={status} onChange={event => setStatus(event.target.value)}><option value="all">Semua status</option>{Object.entries(statusLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>
   <div className="task-view-toggle" role="group" aria-label="Tampilan task"><button aria-label="Tampilan papan" aria-pressed={view === 'board'} onClick={() => setView('board')}><Columns3 size={17} aria-hidden="true" /><span>Papan</span></button><button aria-label="Tampilan daftar" aria-pressed={view === 'list'} onClick={() => setView('list')}><List size={17} aria-hidden="true" /><span>Daftar</span></button></div>
  </div>{brands.error != null && <TaskError error={brands.error} retry={brands.retry} />}<TaskCollection key={brandId} brandId={brandId} brands={brands.data?.items ?? []} search={search} status={status} view={view} canWrite={canWrite} />
  <footer className="list-foot">{canWrite ? 'Perubahan mengikuti izin brand dan aturan lifecycle di server.' : 'Akses baca saja untuk peran Anda.'} Otorisasi tetap wajib ditegakkan backend.</footer></section></div>;
}

function TaskCollection({ brandId, brands, search, status, view, canWrite }: { brandId: string; brands: Brand[]; search: string; status: string; view: 'board' | 'list'; canWrite: boolean }) {
 const load = useCallback(() => api.tasks(brandId || undefined), [brandId]);
 const query = useQuery(load);
 if (query.loading) return <Loading label="Memuat task dari server…" />;
 if (query.error || !query.data) return <TaskError error={query.error} retry={query.retry} />;
 const needle = search.trim().toLocaleLowerCase('id-ID');
 const items = query.data.items.filter(task => (status === 'all' || task.status === status) && (task.title + ' ' + task.objective).toLocaleLowerCase('id-ID').includes(needle)).sort((a, b) => b.updated_at.localeCompare(a.updated_at));
 const brandName = (id: string) => brands.find(brand => brand.id === id)?.name ?? id;
 return <><div className="task-collection-summary"><span aria-live="polite">{items.length} dari {query.data.items.length} task</span><button className="task-text-button" onClick={query.retry}><RotateCw size={13} aria-hidden="true" />Segarkan</button></div>
  {items.length === 0 ? <div className="state"><ListTodo size={30} aria-hidden="true" /><h2>{query.data.items.length === 0 ? 'Belum ada task.' : 'Tidak ada task yang cocok.'}</h2><p>{query.data.items.length === 0 ? 'Mulai dari brief yang jelas. Task baru disimpan sebagai draf, bukan eksekusi agen.' : 'Ubah pencarian, brand, atau filter status.'}</p>{canWrite && query.data.items.length === 0 && <Link className="button primary" to={'/tasks/new' + (brandId ? '?brand_id=' + encodeURIComponent(brandId) : '')}>Buat task pertama</Link>}</div> : view === 'board' ? <div className="task-board">{boardColumns.map(column => {
   const columnItems = items.filter(task => column.statuses.includes(task.status));
   return <section className="task-column" key={column.label} aria-label={column.label}><header><h2>{column.label}</h2><span>{columnItems.length}</span></header><div className="task-column-cards">{columnItems.map(task => <article className="task-card" key={task.id}><div className="task-card-meta"><Priority priority={task.priority} /><TaskStatusBadge status={task.status} /></div><h3><Link to={taskUrl(task.id)}>{task.title}</Link></h3><p className="task-objective-preview">{task.objective || 'Objektif belum ditambahkan.'}</p><div className="task-card-brand">{brandName(task.brand_id)}</div><footer><span>{task.agent_id ? 'Agen: ' + task.agent_id : 'Belum ditugaskan'}</span><ArrowUpRight size={15} aria-hidden="true" /></footer></article>)}{columnItems.length === 0 && <p className="task-column-empty">Tidak ada task di tahap ini.</p>}</div></section>;
  })}</div> : <div className="task-table-scroll" tabIndex={0} role="region" aria-label="Tabel task; geser untuk kolom lainnya"><table className="task-table"><thead><tr><th scope="col">Task / brand</th><th scope="col">Status</th><th scope="col">Prioritas</th><th scope="col">Agen</th><th scope="col">Diperbarui</th></tr></thead><tbody>{items.map(task => <tr key={task.id}><th scope="row"><Link to={taskUrl(task.id)}>{task.title}<ArrowUpRight size={14} aria-hidden="true" /></Link><small>{brandName(task.brand_id)}</small></th><td><TaskStatusBadge status={task.status} /></td><td><Priority priority={task.priority} /></td><td>{task.agent_id || 'Belum ditugaskan'}</td><td><time dateTime={task.updated_at}>{dateLabel(task.updated_at)}</time></td></tr>)}</tbody></table></div>}</>;
}

export function TaskCreate() {
 const { session, onError } = useWorkspace();
 const navigate = useNavigate();
 const [params] = useSearchParams();
 const brands = useQuery(loadBrands);
 const canWrite = writableRoles.has(session.user.role);
 if (!canWrite) return <div className="tasks-page"><TaskBack /><TaskError error={new Error('Peran Anda hanya dapat membaca task. Backend tetap memeriksa otorisasi.')} /></div>;
 return <div className="tasks-page"><TaskBack /><PageHeading eyebrow="TASKS / DRAF BARU" title="Mulai dengan brief yang jelas." description="Simpan task sebagai draf. Penugasan agen hanya metadata, bukan perintah menjalankan pekerjaan." /><RuntimeNotice />
  {brands.loading ? <Loading label="Memuat brand yang diizinkan…" /> : brands.error || !brands.data ? <TaskError error={brands.error} retry={brands.retry} /> : brands.data.items.filter(brand => brand.status === 'active').length === 0 ? <section className="panel state"><h2>Belum ada brand aktif.</h2><p>Task perlu terhubung ke brand aktif yang dapat Anda akses.</p><Link className="button secondary" to="/brands">Lihat brand</Link></section> : <TaskForm brands={brands.data.items.filter(brand => brand.status === 'active')} initialBrandId={params.get('brand_id') || ''} onSave={async input => { try { const task = await api.createTask(input, session.csrf_token); navigate(taskUrl(task.id)); } catch (reason) { onError(reason); throw reason; } }} />}</div>;
}

function TaskBack() {
 return <Link className="back-link" to="/tasks"><ArrowLeft size={16} aria-hidden="true" />Semua task</Link>;
}

function localDeadline(value?: string | null) {
 if (!value) return '';
 const date = new Date(value);
 if (Number.isNaN(date.getTime())) return '';
 const pad = (part: number) => String(part).padStart(2, '0');
 return `${String(date.getFullYear()).padStart(4, '0')}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
function TaskForm({ brands, initial, initialBrandId = '',  onSave, onCancel }: { brands: Brand[]; initial?: Task; initialBrandId?: string; onSave: (input: TaskInput) => Promise<void>; onCancel?: () => void }) {
 const agents = useQuery(loadAgents);
 const [values, setValues] = useState({
  brand_id: initial?.brand_id ?? (brands.some(brand => brand.id === initialBrandId) ? initialBrandId : ''),
  title: initial?.title ?? '', objective: initial?.objective ?? '', instructions: initial?.instructions ?? '',
  priority: initial?.priority ?? 'normal' as TaskPriority, agent_id: initial?.agent_id ?? '',
  deadline_at: localDeadline(initial?.deadline_at),
 });
 const [pending, setPending] = useState(false);
 const [error, setError] = useState<unknown>(null);
 const [errors, setErrors] = useState<Partial<Record<keyof TaskInput, string>>>({});
 const form = useRef<HTMLFormElement>(null);
 const [deadlineChanged, setDeadlineChanged] = useState(false);
 const activeAgentMissing = values.agent_id && !agents.data?.items.some(agent => agent.id === values.agent_id);
 async function submit(event: FormEvent) {
  event.preventDefault();
  if (pending) return;
  let deadline = initial?.deadline_at ?? null;
  if (deadlineChanged) {
   const date = values.deadline_at ? new Date(values.deadline_at) : null;
   if (date && (Number.isNaN(date.getTime()) || date.getFullYear() < 2000 || date.getFullYear() > 9999)) {
    setErrors({ deadline_at: 'Gunakan tanggal valid antara tahun 2000 dan 9999.' }); return;
   }
   deadline = date ? date.toISOString() : null;
  }
  const input = { ...values, deadline_at: deadline, title: values.title.trim(), objective: values.objective.trim(), instructions: values.instructions.trim(), agent_id: values.agent_id || null };
  const parsed = taskInputSchema.safeParse(input);
  if (!parsed.success) {
   const next: typeof errors = {};
   for (const issue of parsed.error.issues) {
    const field = issue.path[0] as keyof TaskInput;
    next[field] = field === 'brand_id' ? 'Pilih brand yang valid.' : field === 'title' ? 'Judul task wajib diisi.' : 'Periksa nilai field ini.';
   }
   setErrors(next);
   const first = Object.keys(next)[0];
   if (first) form.current?.querySelector<HTMLElement>(`[name="${first}"]`)?.focus();
   return;
  }
  setErrors({}); setError(null); setPending(true);
  try { await onSave(parsed.data); } catch (reason) { setError(reason); } finally { setPending(false); }
 }
 return <section className="panel form-panel task-form-panel"><div className="section-heading"><h2>{initial ? 'Edit brief task' : 'Informasi task'}</h2><span className="mini-label">* WAJIB DIISI</span></div><form ref={form} noValidate onSubmit={event => void submit(event)} aria-busy={pending}><fieldset disabled={pending}>
  <label htmlFor="task-brand">Brand *</label><select id="task-brand" name="brand_id" value={values.brand_id} disabled={!!initial} required aria-invalid={!!errors.brand_id} aria-describedby={errors.brand_id ? 'task-brand-error' : 'task-brand-help'} onChange={event => setValues({ ...values, brand_id: event.target.value })}><option value="">Pilih brand</option>{initial && !brands.some(brand => brand.id === initial.brand_id) && <option value={initial.brand_id}>{initial.brand_id}</option>}{brands.map(brand => <option key={brand.id} value={brand.id}>{brand.name}</option>)}</select>{errors.brand_id ? <p className="field-error" id="task-brand-error">{errors.brand_id}</p> : <p className="field-help" id="task-brand-help">{initial ? 'Brand tidak dapat dipindahkan melalui formulir ini.' : 'Hanya brand aktif dari API yang dapat dipilih.'}</p>}
  <label htmlFor="task-title">Judul task *</label><input id="task-title" name="title" value={values.title} required aria-invalid={!!errors.title} aria-describedby={errors.title ? 'task-title-error' : undefined} onChange={event => setValues({ ...values, title: event.target.value })} />{errors.title && <p className="field-error" id="task-title-error">{errors.title}</p>}
  <label htmlFor="task-deadline">Tenggat (waktu lokal)</label><input id="task-deadline" name="deadline_at" type="datetime-local" min="2000-01-01T00:00" max="9999-12-31T23:59" value={values.deadline_at} aria-invalid={!!errors.deadline_at} aria-describedby={errors.deadline_at ? 'task-deadline-error' : 'task-deadline-help'} onChange={event => { setValues({ ...values, deadline_at: event.target.value }); setDeadlineChanged(true); }} />{errors.deadline_at ? <p className="field-error" id="task-deadline-error">{errors.deadline_at}</p> : <p className="field-help" id="task-deadline-help">Opsional. Zona waktu perangkat: {Intl.DateTimeFormat().resolvedOptions().timeZone}. Kosongkan untuk menghapus. Tenggat hanya metadata, bukan penjadwalan agen.</p>}
  <label htmlFor="task-objective">Objektif</label><textarea id="task-objective" name="objective" rows={3} value={values.objective} placeholder="Hasil apa yang ingin dicapai?" onChange={event => setValues({ ...values, objective: event.target.value })} />
  <label htmlFor="task-instructions">Instruksi</label><textarea id="task-instructions" name="instructions" rows={5} value={values.instructions} placeholder="Konteks, batasan, dan kriteria hasil…" onChange={event => setValues({ ...values, instructions: event.target.value })} />
  <div className="task-form-pair"><div><label htmlFor="task-priority">Prioritas</label><select id="task-priority" name="priority" value={values.priority} onChange={event => setValues({ ...values, priority: event.target.value as TaskPriority })}>{Object.entries(priorityLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></div><div><label htmlFor="task-agent">Agen (opsional)</label><select id="task-agent" name="agent_id" value={values.agent_id} aria-describedby="task-agent-help" onChange={event => setValues({ ...values, agent_id: event.target.value })}><option value="">Tanpa agen</option>{activeAgentMissing && <option value={values.agent_id}>{values.agent_id} · Penugasan saat ini</option>}{agents.data?.items.map(agent => <option key={agent.id} value={agent.id}>{agent.name} · {agent.role}</option>)}</select></div></div>
  <p className="field-help" id="task-agent-help">{agents.loading ? 'Memuat direktori agen… ' : ''}Memilih agen mencatat penugasan, tidak mengaktifkan runtime. {initial ? 'Task berstatus siap dapat ditugaskan manual dari detail.' : 'Task baru tetap berstatus draf.'}</p>{agents.error != null && <TaskError error={agents.error} retry={agents.retry} />}
  {error != null && <TaskError error={error} />}<div className="form-actions">{onCancel ? <button type="button" className="button secondary" onClick={onCancel}>Batal edit</button> : <Link className="button secondary" to="/tasks">Kembali</Link>}<button type="submit" className="button primary" disabled={pending}>{pending ? 'Menyimpan…' : initial ? 'Simpan perubahan' : 'Simpan draf'}<Check size={16} aria-hidden="true" /></button></div>
 </fieldset></form></section>;
}

export function TaskDetail() {
 const { id = '' } = useParams();
 return <TaskDetailContent key={id} id={id} />;
}
function TaskDetailContent({ id }: { id: string }) {
 const { session, onError } = useWorkspace();
 const loadTask = useCallback(() => api.task(id), [id]);
 const loadEvents = useCallback(() => api.taskEvents(id), [id]);
 const query = useQuery(loadTask);
 const events = useQuery(loadEvents);
 const brands = useQuery(loadBrands);
 // Keep the draft and its optimistic version fixed until save or explicit cancel.
 const [editing, setEditing] = useState<Task | null>(null);
 const editingRef = useRef(false);
 // Refs close same-turn races before the disabled controls rerender.
 const panelInFlight = useRef(false);
 const taskRefreshing = useRef(false);
 useEffect(() => { if (!query.loading) taskRefreshing.current = false; }, [query.loading]);
 const [panelPending, setPanelPending] = useState(false);
 const [revisionRequestsOpened, setRevisionRequestsOpened] = useState(false);
 const [confirmCancel, setConfirmCancel] = useState(false);
 const [pending, setPending] = useState(false);
 const [mutationError, setMutationError] = useState<unknown>(null);
 const [notice, setNotice] = useState('');
 const canWrite = writableRoles.has(session.user.role);
 function beginPanelChange() {
  if (editingRef.current || panelInFlight.current || taskRefreshing.current || pending || query.loading || query.error) return false;
  panelInFlight.current = true; setPanelPending(true); return true;
 }
 function endPanelChange() { panelInFlight.current = false; setPanelPending(false); }
 function refreshTask() {
  taskRefreshing.current = true; query.retry(); events.retry();
 }
 async function transition(status: 'ready' | 'assigned' | 'cancelled') {
  if (pending || panelInFlight.current || taskRefreshing.current || editingRef.current || query.loading || query.error || !canWrite) return;
  setPending(true); setMutationError(null); setNotice('');
  try {
   if (!query.data) return;
   await api.transitionTask(id, status, session.csrf_token, query.data.version);
   setConfirmCancel(false);
   setNotice('Status task diperbarui. Tidak ada perintah eksekusi yang dikirim.');
   refreshTask();
  } catch (reason) { onError(reason); setMutationError(reason); } finally { setPending(false); }
 }
 if (query.loading && !query.data) return <div className="tasks-page"><TaskBack /><Loading label="Memuat detail task…" /></div>;
 if (!query.data) return <div className="tasks-page"><TaskBack /><TaskError error={query.error} retry={query.retry} /></div>;
 const task = query.data;
 const editable = canWrite && editableStatuses.has(task.status);
 const taskBusy = pending || panelPending || query.loading || !!query.error;
 const panelsAvailable = !editing && !taskBusy;
 const brand = brands.data?.items.find(item => item.id === task.brand_id);
 return <div className="tasks-page"><TaskBack /><PageHeading eyebrow="TASKS / DETAIL" title={task.title} description={brand ? 'Task untuk ' + brand.name : 'Task untuk brand ' + task.brand_id} action={<TaskStatusBadge status={task.status} />} /><RuntimeNotice />
  {query.error != null && <TaskError error={query.error} retry={editing ? undefined : refreshTask} />}
  {notice && <p className="task-success" role="status">{notice}</p>}{!confirmCancel && mutationError != null && <TaskError error={mutationError} />}
  <div className="task-detail-grid"><div>{editing ? <TaskForm initial={editing} brands={brands.data?.items ?? []} onCancel={() => { editingRef.current = false; setEditing(null); }} onSave={async input => {
   const update: TaskUpdate = { title: input.title, objective: input.objective, instructions: input.instructions, priority: input.priority, agent_id: input.agent_id, deadline_at:input.deadline_at, version:editing.version };
   try { await api.updateTask(id, update, session.csrf_token); editingRef.current = false; setEditing(null); setNotice('Brief task tersimpan.'); refreshTask(); } catch (reason) { onError(reason); throw reason; }
  }} /> : <section className="panel task-brief"><div className="section-heading"><h2>Brief pekerjaan</h2>{editable && <button className="button secondary" disabled={taskBusy} onClick={() => { if (panelInFlight.current || taskRefreshing.current || pending || query.loading || query.error) return; editingRef.current = true; setEditing(task); setMutationError(null); }}>Edit brief</button>}</div><dl><dt>Objektif</dt><dd className="preserve-lines">{task.objective || 'Belum ditambahkan.'}</dd><dt>Instruksi</dt><dd className="preserve-lines">{task.instructions || 'Belum ditambahkan.'}</dd></dl></section>}
   <section className="panel task-history" aria-labelledby="task-history-title"><div className="section-heading"><h2 id="task-history-title"><Clock3 size={17} aria-hidden="true" />Riwayat status</h2><button className="task-text-button" disabled={events.loading || pending} onClick={events.retry}><RotateCw size={14} aria-hidden="true" />Segarkan</button></div>{events.loading ? <Loading label="Memuat riwayat status…" /> : events.error || !events.data ? <TaskError error={events.error} retry={events.retry} /> : events.data.items.length === 0 ? <p className="task-history-empty">Belum ada peristiwa yang dicatat server.</p> : <ol className="task-event-list">{[...events.data.items].sort((a, b) => a.created_at.localeCompare(b.created_at)).map(event => <li key={event.id}><div><strong>{event.from_status ? statusLabels[event.from_status] + ' → ' : ''}{statusLabels[event.to_status]}</strong><time dateTime={event.created_at}>{dateLabel(event.created_at)}</time></div><p>Oleh <span>{event.actor_id || 'Sistem'}</span></p></li>)}</ol>}<p className="task-history-note">Riwayat ini mencatat perubahan status dari API, bukan log eksekusi. Catatan manusia ditampilkan terpisah.</p></section>
   <TaskDependencies task={task} canWrite={canWrite && panelsAvailable} canRefresh={panelsAvailable} beginChange={beginPanelChange} endChange={endPanelChange} onChanged={refreshTask} />
   <TaskDeliverables task={task} canWrite={canWrite && panelsAvailable} canRefresh={panelsAvailable} beginChange={beginPanelChange} endChange={endPanelChange} onChanged={refreshTask} />
   <details className="task-revision-disclosure" onToggle={event => { if (event.currentTarget.open) setRevisionRequestsOpened(true); }}><summary>Buka permintaan revisi manusia</summary>{revisionRequestsOpened && <TaskRevisionRequests task={task} canWrite={canWrite && panelsAvailable} canRefresh={panelsAvailable} beginChange={beginPanelChange} endChange={endPanelChange} onChanged={refreshTask} />}</details>
   <TaskComments taskId={task.id} canWrite={canWrite} />
  </div><aside className="panel task-detail-aside"><p className="eyebrow">PENGELOLAAN TASK</p><dl><dt>Prioritas</dt><dd><Priority priority={task.priority} /></dd><dt>Agen</dt><dd>{task.agent_id || 'Belum ditugaskan'}</dd><dt>Tenggat</dt><dd>{task.deadline_at ? <time dateTime={task.deadline_at}>{dateLabel(task.deadline_at)}</time> : 'Tidak ditentukan'}</dd><dt>Dibuat</dt><dd><time dateTime={task.created_at}>{dateLabel(task.created_at)}</time></dd><dt>Diperbarui</dt><dd><time dateTime={task.updated_at}>{dateLabel(task.updated_at)}</time></dd><dt>ID task</dt><dd><code>{task.id}</code></dd></dl>
   {canWrite && !editing && <div className="task-detail-actions">{task.status === 'draft' && <button className="button primary" disabled={taskBusy} onClick={() => void transition('ready')}>{pending ? 'Menyimpan…' : 'Tandai siap'}<Check size={15} aria-hidden="true" /></button>}{task.status === 'ready' && task.agent_id && <button className="button secondary" disabled={taskBusy} onClick={() => void transition('assigned')}>{pending ? 'Menyimpan…' : 'Tugaskan manual'}</button>}{task.status === 'ready' && !task.agent_id && <p>Pilih agen melalui Edit brief sebelum penugasan manual.</p>}{!terminalStatuses.has(task.status) && <button className="button danger-outline" disabled={taskBusy} onClick={() => { setConfirmCancel(true); setMutationError(null); }}>Batalkan task</button>}</div>}
   <p className="task-permission-note">{canWrite ? 'Transisi diperiksa server. Siap dan ditugaskan bukan bukti agen berjalan.' : 'Akses baca saja. Peran Anda tidak memiliki tindakan pengelolaan.'} Otorisasi backend tetap diperlukan.</p>{brands.error != null && <TaskError error={brands.error} retry={brands.retry} />}
  </aside></div>{confirmCancel && canWrite && <CancelTaskDialog title={task.title} pending={pending} error={mutationError} onClose={() => { setConfirmCancel(false); setMutationError(null); }} onConfirm={() => void transition('cancelled')} />}</div>;
}
function CancelTaskDialog({ title, pending, error, onClose, onConfirm }: { title: string; pending: boolean; error: unknown; onClose: () => void; onConfirm: () => void }) {
 const dialog = useRef<HTMLDialogElement>(null);
 useEffect(() => { const element = dialog.current; element?.showModal(); return () => element?.close(); }, []);
 return <dialog ref={dialog} className="confirm-dialog task-cancel-dialog" aria-labelledby="task-cancel-title" aria-describedby="task-cancel-description" onCancel={event => { event.preventDefault(); if (!pending) onClose(); }}><button className="icon-button task-dialog-close" aria-label="Tutup konfirmasi" disabled={pending} onClick={onClose}><X size={19} aria-hidden="true" /></button><p className="eyebrow">KONFIRMASI PERUBAHAN</p><h2 id="task-cancel-title">Batalkan task?</h2><p id="task-cancel-description">Status <strong>{title}</strong> akan diubah menjadi dibatalkan sesuai aturan server. Ini bukan jaminan penghentian proses eksternal; bridge runtime belum diverifikasi.</p>{error != null && <TaskError error={error} />}<div className="form-actions"><button autoFocus className="button secondary" disabled={pending} onClick={onClose}>Kembali</button><button className="button danger" disabled={pending} onClick={onConfirm}>{pending ? 'Membatalkan…' : 'Ya, batalkan task'}</button></div></dialog>;
}
