import { useCallback, useState } from 'react';
import type { FormEvent } from 'react';
import { useWorkspace } from '../app/WorkspaceContext';
import { Loading, ErrorState } from '../components/Feedback';
import { api, taskCommentInputSchema } from '../lib/api';
import type { TaskCommentInput } from '../lib/api';
import { useQuery } from '../lib/useQuery';

export function TaskComments({ taskId, canWrite }: { taskId: string; canWrite: boolean }) {
 const { session, onError } = useWorkspace();
 const load = useCallback(() => api.taskComments(taskId), [taskId]);
 const query = useQuery(load);
 const [body, setBody] = useState('');
 const [kind, setKind] = useState<TaskCommentInput['kind']>('comment');
 const [pending, setPending] = useState(false);
 const [error, setError] = useState<unknown>(null);
 const [validation, setValidation] = useState('');
 const [notice, setNotice] = useState('');
 async function submit(event: FormEvent) {
  event.preventDefault();
  if (pending || !canWrite) return;
  const parsed = taskCommentInputSchema.safeParse({ body, kind });
  if (!parsed.success) { setValidation('Isi pesan antara 1 dan 4000 karakter.'); return; }
  setPending(true); setError(null); setValidation(''); setNotice('');
  try {
   await api.addTaskComment(taskId, parsed.data, session.csrf_token);
   setBody(''); setNotice('Catatan tersimpan. Tidak ada perintah eksekusi atau persetujuan yang dikirim.'); query.retry();
  } catch (reason) { setError(reason); onError(reason); }
  finally { setPending(false); }
 }
 return <section className="panel task-comments" aria-labelledby="task-comments-title">
  <div className="section-heading"><h2 id="task-comments-title">Komentar & klarifikasi</h2><button className="task-text-button" disabled={query.loading || pending} onClick={query.retry}>Segarkan catatan</button></div>
  <p className="task-history-note">Catatan manusia, bukan log agen atau approval. Klarifikasi tidak otomatis melanjutkan pekerjaan. Maksimal 500 catatan pertama; pagination belum tersedia.</p>
  {query.loading ? <Loading label="Memuat catatan…" /> : query.error || !query.data ? <ErrorState error={query.error} retry={query.retry} /> : query.data.items.length === 0 ? <p className="task-history-empty">Belum ada catatan.</p> : <ol className="task-comment-list">{query.data.items.map(comment => <li key={comment.id}>
   <header><strong>{comment.kind === 'clarification' ? 'Klarifikasi' : 'Komentar'}</strong><time dateTime={comment.created_at}>{new Date(comment.created_at).toLocaleString('id-ID')}</time></header>
   <p className="preserve-lines">{comment.body}</p><small>Oleh {comment.author_id}</small>
  </li>)}</ol>}
  {notice && <p className="task-success" role="status">{notice}</p>}
  {canWrite ? <form className="task-comment-form" aria-label="Tulis catatan" onSubmit={event => void submit(event)} aria-busy={pending}><fieldset disabled={pending}>
   <label htmlFor="task-comment-kind">Jenis catatan</label><select id="task-comment-kind" value={kind} onChange={event => setKind(event.target.value as TaskCommentInput['kind'])}><option value="comment">Komentar</option><option value="clarification">Klarifikasi</option></select>
   <label htmlFor="task-comment-body">Pesan</label><textarea id="task-comment-body" rows={4} value={body} aria-invalid={!!validation} aria-describedby={validation ? 'task-comment-error' : 'task-comment-help'} onChange={event => setBody(event.target.value)} />
   {validation ? <p id="task-comment-error" className="field-error">{validation}</p> : <p className="field-help" id="task-comment-help">Teks biasa. Catatan tersimpan tidak dapat diedit atau dihapus pada subset ini.</p>}
   {error != null && <ErrorState error={error} />}
   <div className="form-actions"><button className="button primary" type="submit" disabled={pending}>{pending ? 'Menyimpan…' : 'Kirim catatan'}</button></div>
  </fieldset></form> : <p className="task-history-note">Akses baca saja. Peran Anda tidak dapat mengirim catatan.</p>}
 </section>;
}
