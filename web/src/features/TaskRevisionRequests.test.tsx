import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { WorkspaceContext } from '../app/WorkspaceContext';
import { TaskRevisionRequests } from './TaskRevisionRequests';
import { brandFixture, fixtureResponse, sessionFixture } from '../test/fixtures';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { TaskDetail } from './Tasks';
import type { Task } from '../lib/api';

// UI transport fixtures only; no live backend, database, auth or execution.
const task: Task = { id: 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8', brand_id: 'd0e0ddbb-23f8-4fe7-9be7-e53dd92e9860', title: 'Human brief', objective: '', instructions: '', priority: 'normal', status: 'draft', version: 3, created_at: '2026-11-06T00:00:00Z', updated_at: '2026-11-06T00:00:00Z' };
const taskPath = '/api/tasks/' + task.id;
const path = taskPath + '/revision-requests';
const targetsPath = taskPath + '/deliverables';
const authorId = 'c40a91e8-8f35-43d7-857b-d756c5179b60';
const targetId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const oldId = 'd5be6a58-cd9e-4d8d-9136-e7b8925cdca3';
const target = { id: targetId, task_id: task.id, revision: 2, author_id: authorId, origin: 'human_import', validation_status: 'validated', title: '<img src=x> New report', kind: 'research_report', created_at: task.created_at };
const old = { ...target, id: oldId, revision: 1, validation_status: 'invalid', title: 'Historical output', kind: '' };
const note = { id: 'aedcf272-80d1-40ee-93b1-3b38da438755', task_id: task.id, deliverable_id: oldId, deliverable_revision: 1, author_id: authorId, body: '<script>literal</script> https://example.test/ref **not markdown**', created_at: task.created_at };
function response(data: unknown, status = 200) { return new Response(JSON.stringify(data), { status }); }
function transport(items: unknown[] = [note]) {
 const fetcher = vi.fn().mockImplementation((requestPath: string) => {
  if (requestPath === path) return Promise.resolve(response({ version: 7, items }));
  if (requestPath === targetsPath) return Promise.resolve(response({ version: 7, items: [target, old] }));
  throw new Error('Unexpected endpoint: ' + requestPath);
 });
 vi.stubGlobal('fetch', fetcher);
 return fetcher;
}
function show(options: { role?: string; status?: Task['status']; canWrite?: boolean; canRefresh?: boolean; onChanged?: () => void; beginChange?: () => boolean; endChange?: () => void } = {}) {
 return render(<WorkspaceContext.Provider value={{ session: { ...sessionFixture, user: { ...sessionFixture.user, role: options.role ?? 'owner' } }, onError: vi.fn() }}><TaskRevisionRequests task={{ ...task, status: options.status ?? 'draft' }} canWrite={options.canWrite ?? true} canRefresh={options.canRefresh} onChanged={options.onChanged ?? vi.fn()} beginChange={options.beginChange} endChange={options.endChange} /></WorkspaceContext.Provider>);
}
const bodyInput = () => screen.getByRole('textbox', { name: 'Catatan permintaan revisi' });
const targetInput = () => screen.getByRole('combobox', { name: 'Revisi deliverable tujuan' });
const submitForm = () => fireEvent.submit(screen.getByRole('form', { name: 'Minta revisi deliverable' }));
async function compose() {
 fireEvent.click(screen.getByRole('button', { name: 'Tulis permintaan revisi' }));
 await waitFor(() => expect(bodyInput()).toBeEnabled());
}

it('lists immutable human notes with escaped body, exact target revision, author and timestamp without execution claims', async () => {
 transport(); show({ role: 'viewer' });
 const panel = screen.getByRole('region', { name: 'Permintaan revisi manusia' });
 expect(await within(panel).findByText(note.body)).toBeInTheDocument();
 expect(within(panel).getByText('Deliverable ' + oldId + ' · Revisi 1')).toBeInTheDocument();
 expect(within(panel).getByText('Oleh ' + authorId)).toBeInTheDocument();
 expect(panel.querySelector('time')).toHaveAttribute('datetime', note.created_at);
 expect(within(panel).getByText(/100 catatan terbaru/)).toHaveTextContent(/tidak dapat diedit atau dihapus/);
 expect(within(panel).getByText(/tidak menjalankan agen/)).toHaveTextContent(/mengulang eksekusi.*menyetujui.*mengubah status/);
 expect(panel.querySelector('script,img,a')).toBeNull();
 expect(within(panel).queryByRole('form')).not.toBeInTheDocument();
});

it('explicitly selects a historical invalid target and submits the coherent snapshot version with current CSRF then refreshes both snapshots', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 let accepted = false;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') { accepted = true; return Promise.resolve(response({ ...note, body: 'Please correct this' }, 201)); }
  if (accepted && requestPath === path) return Promise.resolve(response({ version: 8, items: [{ ...note, body: 'Please correct this' }] }));
  if (accepted && requestPath === targetsPath) return Promise.resolve(response({ version: 8, items: [target, old] }));
  return base(requestPath, options);
 });
 const onChanged = vi.fn(); show({ onChanged });
 await compose();
 expect(targetInput()).toHaveValue('');
 expect(screen.getByRole('option', { name: /Revisi 2.*<img src=x> New report.*Schema tervalidasi/ })).toBeInTheDocument();
 expect(screen.getByRole('option', { name: /Revisi 1.*Historical output.*Schema tidak valid/ })).toBeInTheDocument();
 expect(bodyInput()).not.toHaveAttribute('maxlength');
 fireEvent.change(targetInput(), { target: { value: oldId } });
 fireEvent.change(bodyInput(), { target: { value: '\u0085  Please correct this  \u0085' } });
 submitForm();
 await screen.findByText(/Permintaan revisi tersimpan/);
 expect(fetcher).toHaveBeenCalledWith(path, expect.objectContaining({ method: 'POST', credentials: 'same-origin', headers: expect.objectContaining({ 'X-CSRF-Token': sessionFixture.csrf_token }), body: JSON.stringify({ deliverable_id: oldId, body: 'Please correct this', version: 7 }) }));
 await waitFor(() => expect(bodyInput()).toBeEnabled());
 expect(bodyInput()).toHaveValue('');
 expect(targetInput()).toHaveValue(oldId);
 expect(onChanged).toHaveBeenCalledTimes(1);
 expect(fetcher.mock.calls.filter(([requestPath, options]) => requestPath === path && options.method === 'GET')).toHaveLength(2);
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === targetsPath)).toHaveLength(2);
 expect(fetcher.mock.calls.every(([requestPath]) => requestPath === path || requestPath === targetsPath)).toBe(true);
});

it.each(['owner', 'administrator', 'marketing_manager', 'editor'])('allows canonical writer %s', async role => {
 transport(); show({ role }); await compose();
});
it.each(['draft', 'ready', 'assigned'] as const)('allows preparation status %s', async status => {
 transport(); show({ status }); await compose();
});
it.each(['viewer', 'analyst', 'approver'])('is readonly for %s', async role => {
 transport(); show({ role }); await screen.findByText(note.body);
 expect(screen.queryByRole('button', { name: 'Tulis permintaan revisi' })).not.toBeInTheDocument();
});
it.each(['queued', 'running', 'waiting_input', 'waiting_approval', 'completed', 'failed', 'cancelled'] as const)('is readonly in %s', async status => {
 transport(); show({ status }); await screen.findByText(note.body);
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', { name: 'Tulis permintaan revisi' })).not.toBeInTheDocument();
});
it.each([['blank', '\u0085 \n\t'], ['NUL', 'fix\0this'], ['4001 codepoints', '😀'.repeat(4001)]])('rejects %s without POST', async (_name, body) => {
 const fetcher = transport(); show(); await compose();
 fireEvent.change(targetInput(), { target: { value: oldId } });
 fireEvent.change(bodyInput(), { target: { value: body } }); submitForm();
 expect(await screen.findByText('Isi catatan 1–4000 karakter Unicode setelah pemangkasan spasi, tanpa NUL.')).toBeInTheDocument();
 expect(fetcher.mock.calls.some(([, options]) => options.method === 'POST')).toBe(false);
 expect(bodyInput()).toHaveValue(body);
});
it('accepts 4000 emoji without UTF-16 truncation and prevents same-turn duplicate submissions until accepted', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 let resolveWrite!: (value: Response) => void;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && options.method === 'POST' ? new Promise<Response>(resolve => { resolveWrite = resolve; }) : base(requestPath, options));
 const beginChange = vi.fn(() => true); const endChange = vi.fn(); show({ beginChange, endChange }); await compose();
 const body = '😀'.repeat(4000);
 fireEvent.change(targetInput(), { target: { value: targetId } });
 fireEvent.change(bodyInput(), { target: { value: body } });
 act(() => { submitForm(); submitForm(); });
 expect(bodyInput()).toHaveValue(body); expect(bodyInput()).toBeDisabled();
 expect(fetcher.mock.calls.filter(([, options]) => options.method === 'POST')).toHaveLength(1);
 expect(beginChange).toHaveBeenCalledTimes(1);
 await act(async () => resolveWrite(response({ ...note, deliverable_id: targetId, deliverable_revision: 2, body }, 201)));
 await waitFor(() => expect(bodyInput()).toHaveValue(''));
 expect(endChange).toHaveBeenCalledTimes(1);
});
it('preserves body and target after 409 and requires manual coherent reload, never automatic POST retry', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!; let version = 7;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') { version = 8; return Promise.resolve(response({ error: { code: 'version_conflict', message: 'Task changed.' } }, 409)); }
  if (requestPath === path) return Promise.resolve(response({ version, items: [note] }));
  if (requestPath === targetsPath) return Promise.resolve(response({ version, items: [target, old] }));
  return base(requestPath, options);
 });
 show(); await compose();
 fireEvent.change(targetInput(), { target: { value: oldId } }); fireEvent.change(bodyInput(), { target: { value: 'Keep this draft' } }); submitForm();
 expect(await screen.findByText('Versi task berubah.')).toBeInTheDocument();
 expect(bodyInput()).toHaveValue('Keep this draft'); expect(targetInput()).toHaveValue(oldId);
 expect(screen.getByRole('button', { name: 'Kirim permintaan revisi' })).toBeDisabled();
 expect(fetcher.mock.calls.filter(([, options]) => options.method === 'POST')).toHaveLength(1);
 fireEvent.click(screen.getByRole('button', { name: 'Muat ulang task dan permintaan revisi' }));
 await waitFor(() => expect(bodyInput()).toBeEnabled());
 expect(bodyInput()).toHaveValue('Keep this draft'); expect(targetInput()).toHaveValue(oldId);
 expect(fetcher.mock.calls.filter(([, options]) => options.method === 'POST')).toHaveLength(1);
 submitForm();
 await waitFor(() => expect(fetcher).toHaveBeenLastCalledWith(path, expect.objectContaining({ method: 'POST', body: JSON.stringify({ deliverable_id: oldId, body: 'Keep this draft', version: 8 }) })));
});
it('blocks differing target/request snapshot versions until explicit refresh with no task-version fallback', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!; let version = 6;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === targetsPath ? Promise.resolve(response({ version, items: [target, old] })) : base(requestPath, options));
 show(); fireEvent.click(screen.getByRole('button', { name: 'Tulis permintaan revisi' }));
 expect(await screen.findByText(/Snapshot permintaan dan deliverable berbeda/)).toBeInTheDocument();
 expect(bodyInput()).toBeDisabled(); submitForm();
 expect(fetcher.mock.calls.some(([, options]) => options.method === 'POST')).toBe(false);
 version = 7; fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 await waitFor(() => expect(bodyInput()).toBeEnabled());
});
it('shows server rejection without clearing drafts, fabricating a note or refreshing', async () => {
 const fetcher = transport([]); const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && options.method === 'POST' ? Promise.resolve(response({ error: { code: 'invalid_input', message: 'Request rejected.' } }, 422)) : base(requestPath, options));
 const onChanged = vi.fn(); show({ onChanged }); await compose();
 fireEvent.change(targetInput(), { target: { value: oldId } }); fireEvent.change(bodyInput(), { target: { value: 'Do not discard' } }); submitForm();
 expect(await screen.findByRole('alert')).toHaveTextContent('Request rejected.');
 expect(bodyInput()).toHaveValue('Do not discard'); expect(targetInput()).toHaveValue(oldId);
 expect(screen.getByText('Belum ada permintaan revisi.')).toBeInTheDocument();
 expect(onChanged).not.toHaveBeenCalled();
});
it('distinguishes unavailable lists from empty data and retries explicitly', async () => {
 const fetcher = transport([]); const base = fetcher.getMockImplementation()!; let fail = true;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && fail ? Promise.resolve(response({ error: { code: 'not_found', message: 'Endpoint unavailable.' } }, 404)) : base(requestPath, options));
 show();
 expect(await screen.findByRole('alert')).toHaveTextContent('Endpoint unavailable.');
 expect(screen.queryByText('Belum ada permintaan revisi.')).not.toBeInTheDocument();
 fail = false; fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 expect(await screen.findByText('Belum ada permintaan revisi.')).toBeInTheDocument();
});

it('distinguishes unavailable target snapshots from empty candidates and preserves an explicitly selected missing revision', async () => {
 const fetcher = transport([]); const base = fetcher.getMockImplementation()!; let state = 'available';
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === targetsPath && state === 'failed') return Promise.resolve(response({ error: { code: 'not_found', message: 'Targets unavailable.' } }, 404));
  if (requestPath === targetsPath && state === 'empty') return Promise.resolve(response({ version: 7, items: [] }));
  return base(requestPath, options);
 });
 show(); await compose();
 fireEvent.change(targetInput(), { target: { value: oldId } }); fireEvent.change(bodyInput(), { target: { value: 'Keep target and body' } });
 state = 'failed'; fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 expect(await screen.findByRole('alert')).toHaveTextContent('Targets unavailable.');
 expect(bodyInput()).toBeDisabled(); expect(targetInput()).toHaveValue(oldId); expect(bodyInput()).toHaveValue('Keep target and body');
 state = 'empty'; fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 expect(await screen.findByText('Tidak ada revisi deliverable dalam snapshot ini.')).toBeInTheDocument();
 expect(targetInput()).toHaveValue(oldId);
 expect(screen.getByRole('option', { name: /Pilihan sebelumnya.*tidak tersedia/ })).toBeInTheDocument();
 expect(screen.getByRole('button', { name: 'Kirim permintaan revisi' })).toBeDisabled(); submitForm();
 expect(fetcher.mock.calls.some(([, options]) => options.method === 'POST')).toBe(false);
});
it('blocks submissions while either snapshot is loading and honors a rejected shared guard', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 let resolveTargets!: (value: Response) => void;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === targetsPath ? new Promise<Response>(resolve => { resolveTargets = resolve; }) : base(requestPath, options));
 const beginChange = vi.fn(() => false); const endChange = vi.fn(); show({ beginChange, endChange });
 await screen.findByText(note.body);
 fireEvent.click(screen.getByRole('button', { name: 'Tulis permintaan revisi' }));
 expect(bodyInput()).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Segarkan permintaan revisi' })).toBeDisabled();
 submitForm();
 expect(beginChange).not.toHaveBeenCalled();
 await act(async () => resolveTargets(response({ version: 7, items: [target, old] })));
 fireEvent.change(targetInput(), { target: { value: oldId } }); fireEvent.change(bodyInput(), { target: { value: 'Guarded draft' } }); submitForm();
 expect(beginChange).toHaveBeenCalledTimes(1); expect(endChange).not.toHaveBeenCalled();
 expect(fetcher.mock.calls.some(([, options]) => options.method === 'POST')).toBe(false);
 expect(bodyInput()).toHaveValue('Guarded draft');
});
it('uses current CSRF while retaining the original loaded version and explicit target across parent updates', async () => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && options.method === 'POST' ? Promise.resolve(response(note, 201)) : base(requestPath, options));
 const onError = vi.fn(); const onChanged = vi.fn();
 const view = render(<WorkspaceContext.Provider value={{ session: sessionFixture, onError }}><TaskRevisionRequests task={task} canWrite onChanged={onChanged} /></WorkspaceContext.Provider>);
 await compose();
 fireEvent.change(targetInput(), { target: { value: oldId } }); fireEvent.change(bodyInput(), { target: { value: 'Exact loaded target' } });
 view.rerender(<WorkspaceContext.Provider value={{ session: { ...sessionFixture, csrf_token: 'rotated-csrf' }, onError }}><TaskRevisionRequests task={{ ...task, version: 99 }} canWrite onChanged={onChanged} /></WorkspaceContext.Provider>);
 submitForm();
 await waitFor(() => expect(fetcher).toHaveBeenCalledWith(path, expect.objectContaining({ method: 'POST', headers: expect.objectContaining({ 'X-CSRF-Token': 'rotated-csrf' }), body: JSON.stringify({ deliverable_id: oldId, body: 'Exact loaded target', version: 7 }) })));
});
it('honors parent write and refresh restrictions', async () => {
 const fetcher = transport(); show({ canWrite: false, canRefresh: false }); await screen.findByText(note.body);
 expect(screen.queryByRole('button', { name: 'Tulis permintaan revisi' })).not.toBeInTheDocument();
 expect(screen.getByRole('button', { name: 'Segarkan permintaan revisi' })).toBeDisabled();
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === path)).toHaveLength(1);
});

function detailTransport() {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path || requestPath === targetsPath) return base(requestPath, options);
  if (requestPath === taskPath) return Promise.resolve(response(task));
  if (requestPath === targetsPath + '/' + targetId) return Promise.resolve(response({ ...target, raw_output: 'raw import', structured: { schema_version: 1, kind: target.kind, title: target.title, sections: [{ heading: 'Finding', body: 'Human finding' }] } }));
  if (requestPath === taskPath + '/dependencies') return Promise.resolve(response({ version: 7, items: [] }));
  if (requestPath === taskPath + '/events' || requestPath === taskPath + '/comments') return Promise.resolve(response({ items: [] }));
  if (requestPath === '/api/brands') return Promise.resolve(response({ items: [{ ...brandFixture, id: task.brand_id }] }));
  if (requestPath.startsWith('/api/tasks?')) return Promise.resolve(response({ items: [] }));
  return Promise.resolve(response(fixtureResponse(requestPath)));
 });
 return fetcher;
}
async function detail() {
 render(<WorkspaceContext.Provider value={{ session: sessionFixture, onError: vi.fn() }}><MemoryRouter initialEntries={['/tasks/' + task.id]}><Routes><Route path="/tasks/:id" element={<TaskDetail />} /></Routes></MemoryRouter></WorkspaceContext.Provider>);
 await waitFor(() => expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeEnabled());
 fireEvent.click(screen.getByText('Buka permintaan revisi manusia'));
 await screen.findByRole('region', { name: 'Permintaan revisi manusia' });
 await compose();
}
it('integrates a persistent revision panel while protecting dirty brief and raw drafts during edits and manual refresh', async () => {
 const fetcher = detailTransport(); await detail();
 fireEvent.change(bodyInput(), { target: { value: 'Unsent revision request' } });
 fireEvent.change(targetInput(), { target: { value: oldId } });
 fireEvent.change(screen.getByLabelText('Output mentah untuk impor'), { target: { value: 'Unsent raw import' } });
 fireEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
 fireEvent.change(screen.getByLabelText('Judul task *'), { target: { value: 'Unsaved brief' } });
 expect(screen.getByRole('button', { name: 'Segarkan permintaan revisi' })).toBeDisabled();
 fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 expect(screen.getByLabelText('Judul task *')).toHaveValue('Unsaved brief');
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath)).toHaveLength(1);
 fireEvent.click(screen.getByRole('button', { name: 'Batal edit' }));
 expect(bodyInput()).toHaveValue('Unsent revision request'); expect(targetInput()).toHaveValue(oldId);
 fireEvent.click(screen.getByRole('button', { name: 'Segarkan permintaan revisi' }));
 await waitFor(() => expect(bodyInput()).toBeEnabled());
 expect(screen.getByLabelText('Output mentah untuk impor')).toHaveValue('Unsent raw import');
 expect(bodyInput()).toHaveValue('Unsent revision request');
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath)).toHaveLength(2);
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath + '/events')).toHaveLength(2);
});
it('blocks sibling changes and same-turn brief edits while a revision request or follow-up task refresh is pending', async () => {
 const fetcher = detailTransport(); const base = fetcher.getMockImplementation()!;
 let resolveWrite!: (value: Response) => void; let resolveTask!: (value: Response) => void; let saved = false;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') return new Promise<Response>(resolve => { resolveWrite = resolve; });
  if (requestPath === taskPath && saved) return new Promise<Response>(resolve => { resolveTask = resolve; });
  return base(requestPath, options);
 });
 await detail();
 fireEvent.change(screen.getByLabelText('Output mentah untuk impor'), { target: { value: 'Keep raw draft' } });
 fireEvent.change(bodyInput(), { target: { value: 'Fix specific revision' } }); fireEvent.change(targetInput(), { target: { value: oldId } });
 const edit = screen.getByRole('button', { name: 'Edit brief' });
 act(() => { submitForm(); fireEvent.click(edit); });
 expect(screen.queryByLabelText('Judul task *')).not.toBeInTheDocument();
 expect(edit).toBeDisabled(); expect(screen.getByRole('button', { name: 'Tandai siap' })).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Segarkan deliverable' })).toBeDisabled();
 saved = true; await act(async () => resolveWrite(response({ ...note, body: 'Fix specific revision' }, 201)));
 expect(edit).toBeDisabled();
 await act(async () => resolveTask(response({ ...task, version: 8 })));
 await waitFor(() => expect(edit).toBeEnabled());
 expect(bodyInput()).toHaveValue('');
 expect(screen.getByLabelText('Output mentah untuk impor')).toHaveValue('Keep raw draft');
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath + '/events')).toHaveLength(2);
});
