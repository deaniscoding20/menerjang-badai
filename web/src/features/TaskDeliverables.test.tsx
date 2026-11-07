import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { expect, it, vi } from 'vitest';
import { WorkspaceContext } from '../app/WorkspaceContext';
import { App } from '../app/App';
import { TaskDeliverables } from './TaskDeliverables';
import { brandFixture, fixtureResponse, sessionFixture } from '../test/fixtures';
import type { Task } from '../lib/api';

// UI transport fixtures only: no live backend, authentication, database or runtime.
const task: Task = { id: 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8', brand_id: 'd0e0ddbb-23f8-4fe7-9be7-e53dd92e9860', title: 'Audience brief', objective: '', instructions: '', priority: 'normal', status: 'draft', version: 3, created_at: '2026-11-06T00:00:00Z', updated_at: '2026-11-06T00:00:00Z' };
const path = '/api/tasks/' + task.id + '/deliverables';
const authorId = 'c40a91e8-8f35-43d7-857b-d756c5179b60';
const validId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const invalidId = 'd5be6a58-cd9e-4d8d-9136-e7b8925cdca3';
const session = { ...sessionFixture, user: { ...sessionFixture.user, id: authorId } };
const validSummary = { id: validId, task_id: task.id, revision: 2, author_id: authorId, origin: 'human_import', validation_status: 'validated', title: '<img src=x onerror=alert(1)>', kind: 'research_report', created_at: task.created_at };
const invalidSummary = { ...validSummary, id: invalidId, revision: 1, validation_status: 'invalid', title: 'Unstructured import', kind: '' };
const structured = { schema_version: 1, kind: 'research_report', title: validSummary.title, sections: [{ heading: '<script>Finding</script>', body: '<iframe src="bad">literal body</iframe> **not markdown**' }], sources: [{ label: '<b>Reference</b>', url: 'https://example.test/report' }] };
const validDetail = { ...validSummary, raw_output: '<script>raw newest</script>', structured };
const invalidDetail = { ...invalidSummary, raw_output: '<img src=x onerror=alert(2)> old raw', structured: null };
function showDetail() { render(<MemoryRouter initialEntries={['/tasks/' + task.id]}><App /></MemoryRouter>); }
function readTransport(role = 'viewer') {
 const fetcher = vi.fn().mockImplementation((requestPath: string) => {
  let data: unknown;
  if (requestPath === '/api/session') data = { ...session, user: { ...session.user, role } };
  else if (requestPath === path) data = { version: 7, items: [validSummary, invalidSummary] };
  else if (requestPath === path + '/' + validId) data = validDetail;
  else if (requestPath === path + '/' + invalidId) data = invalidDetail;
  else if (requestPath === '/api/tasks/' + task.id) data = task;
  else if (requestPath === '/api/tasks/' + task.id + '/dependencies') data = { version: 7, items: [] };
  else if (requestPath === '/api/tasks/' + task.id + '/events' || requestPath === '/api/tasks/' + task.id + '/comments') data = { items: [] };
  else if (requestPath === '/api/brands') data = { items: [{ ...brandFixture, id: task.brand_id }] };
  else if (requestPath.startsWith('/api/tasks?')) data = { items: [] };
  else data = fixtureResponse(requestPath);
  return Promise.resolve(new Response(JSON.stringify(data)));
 });
 vi.stubGlobal('fetch', fetcher);
 return fetcher;
}

it('shows human-imported revisions between dependencies and comments with escaped structured/raw text and no approval claim', async () => {
 const fetcher = readTransport(); showDetail();
 const section = await screen.findByRole('region', { name: 'Deliverable impor manusia' });
 await within(section).findByText(validDetail.raw_output);
 expect(within(section).getByRole('heading', { name: validSummary.title })).toBeInTheDocument();
 expect(within(section).getByText(structured.sections[0]!.body)).toBeInTheDocument();
 expect(within(section).getByText('Schema tervalidasi')).toBeInTheDocument();
 expect(within(section).getByText('Schema tidak valid')).toBeInTheDocument();
 expect(within(section).getByText(/bukan verifikasi fakta atau persetujuan/)).toBeInTheDocument();
 expect(within(section).getByText(/100 revisi terbaru/)).toBeInTheDocument();
 expect(within(section).getByText(/tidak menjalankan agen/)).toBeInTheDocument();
 expect(within(section).queryByRole('form')).not.toBeInTheDocument();
 expect(section.querySelector('script,img,iframe,b')).toBeNull();
 expect(section.querySelector('a')).toBeNull(); // Sources are plain text, not navigable URLs.
 const dependencies = screen.getByRole('region', { name: 'Dependensi task' });
 const comments = screen.getByRole('region', { name: 'Komentar & klarifikasi' });
 expect(dependencies.compareDocumentPosition(section) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
 expect(section.compareDocumentPosition(comments) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
 expect(fetcher.mock.calls.filter(([requestPath]) => String(requestPath).startsWith(path + '/'))).toHaveLength(1);
 expect(screen.getByRole('button', { name: /Jalankan task/ })).toBeDisabled();
});

it('keeps old revisions selectable without prefilling imports and refreshes collection/task/history after importing a new revision', async () => {
 const fetcher = readTransport('owner');
 const base = fetcher.getMockImplementation()!;
 const raw = '  arbitrary human output 😀  ';
 const newId = 'aedcf272-80d1-40ee-93b1-3b38da438755';
 const saved = { ...invalidDetail, id: newId, revision: 3, title: 'New human import', raw_output: raw };
 let imported = false;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') { imported = true; return Promise.resolve(new Response(JSON.stringify(saved), { status: 201 })); }
  if (requestPath === path && imported) return Promise.resolve(new Response(JSON.stringify({ version: 8, items: [saved, validSummary, invalidSummary] })));
  if (requestPath === path + '/' + newId) return Promise.resolve(new Response(JSON.stringify(saved)));
  return base(requestPath, options);
 });
 showDetail();
 await screen.findByText(validDetail.raw_output);
 await userEvent.click(screen.getByRole('button', { name: /Revisi 1 · Unstructured import/ }));
 await screen.findByText(invalidDetail.raw_output);
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 expect(input).toHaveValue('');
 expect(input).not.toHaveAttribute('maxlength');
 expect(screen.getByText(/tetap disimpan dengan status schema tidak valid/)).toBeInTheDocument();
 fireEvent.change(input, { target: { value: raw } });
 await userEvent.click(screen.getByRole('button', { name: 'Impor revisi baru' }));
 await screen.findByRole('article', { name: 'Detail revisi 3' });
 expect(screen.getByText(raw.trim())).toBeInTheDocument();
 expect(screen.getByRole('textbox', { name: 'Output mentah untuk impor' })).toHaveValue('');
 expect(fetcher).toHaveBeenCalledWith(path, expect.objectContaining({ method: 'POST', credentials: 'same-origin', headers: expect.objectContaining({ 'X-CSRF-Token': session.csrf_token }), body: JSON.stringify({ raw_output: raw, version: 7 }) }));
 await waitFor(() => expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === '/api/tasks/' + task.id)).toHaveLength(2));
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === '/api/tasks/' + task.id + '/events')).toHaveLength(2);
 expect(fetcher.mock.calls.filter(([requestPath, options]) => requestPath === path && options.method === 'GET')).toHaveLength(2);
 expect(fetcher.mock.calls.some(([requestPath]) => String(requestPath).includes('/transition') || String(requestPath).includes('/runtime'))).toBe(false);
 await userEvent.click(screen.getByRole('button', { name: /Revisi 2 ·/ }));
 expect(await screen.findByText(validDetail.raw_output)).toBeInTheDocument();
 expect(screen.getByRole('textbox', { name: 'Output mentah untuk impor' })).toHaveValue('');
});

function showSlice(options: { role?: string; status?: Task['status']; canWrite?: boolean; onChanged?: () => void; onError?: (reason: unknown) => void } = {}) {
 return render(<WorkspaceContext.Provider value={{ session: { ...session, user: { ...session.user, role: options.role ?? 'owner' } }, onError: options.onError ?? vi.fn() }}><TaskDeliverables task={{ ...task, status: options.status ?? 'draft' }} canWrite={options.canWrite ?? true} onChanged={options.onChanged ?? vi.fn()} /></WorkspaceContext.Provider>);
}

it('requires manual snapshot reload after a conflict before resubmission and never retries POST automatically', async () => {
 const fetcher = readTransport('owner');
 const base = fetcher.getMockImplementation()!;
 let version = 7;
 let writes = 0;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') { writes += 1; version = 8; return Promise.resolve(new Response(JSON.stringify({ error: { code: 'version_conflict', message: 'Task changed.' } }), { status: 409 })); }
  if (requestPath === path) return Promise.resolve(new Response(JSON.stringify({ version, items: [validSummary, invalidSummary] })));
  return base(requestPath, options);
 });
 const onChanged = vi.fn(); const onError = vi.fn();
 showSlice({ onChanged, onError });
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 await waitFor(() => expect(input).toBeEnabled());
 fireEvent.change(input, { target: { value: 'manual import' } });
 await userEvent.click(screen.getByRole('button', { name: 'Impor revisi baru' }));
 expect(await screen.findByText('Versi task berubah.')).toBeInTheDocument();
 expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeDisabled();
 expect(input).toHaveValue('manual import');
 expect(writes).toBe(1);
 expect(onChanged).not.toHaveBeenCalled();
 expect(onError).toHaveBeenCalledWith(expect.objectContaining({ code: 'version_conflict' }));
 await userEvent.click(screen.getByRole('button', { name: 'Muat ulang task dan deliverable' }));
 await waitFor(() => expect(input).toBeEnabled());
 expect(writes).toBe(1);
 expect(onChanged).toHaveBeenCalledTimes(1);
 await userEvent.click(screen.getByRole('button', { name: 'Impor revisi baru' }));
 expect(writes).toBe(2);
 expect(fetcher).toHaveBeenLastCalledWith(path, expect.objectContaining({ method: 'POST', body: JSON.stringify({ raw_output: 'manual import', version: 8 }) }));
});

it.each(['owner', 'administrator', 'marketing_manager', 'editor'])('permits the canonical write role %s to import', async role => {
 readTransport(); showSlice({ role });
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 await waitFor(() => expect(input).toBeEnabled());
});
it.each(['draft', 'ready', 'assigned'] as const)('permits imports in %s', async status => {
 readTransport(); showSlice({ status });
 await waitFor(() => expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeEnabled());
});
it.each(['viewer', 'analyst', 'approver'])('allows readonly %s to view revisions but not import', async role => {
 const fetcher = readTransport(); showSlice({ role });
 await screen.findByText(validDetail.raw_output);
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
 await userEvent.click(screen.getByRole('button', { name: /Revisi 1 ·/ }));
 expect(await screen.findByText(invalidDetail.raw_output)).toBeInTheDocument();
 expect(fetcher.mock.calls.every(([, options]) => options.method === 'GET')).toBe(true);
});
it.each(['queued', 'running', 'waiting_input', 'waiting_approval', 'completed', 'failed', 'cancelled'] as const)('keeps deliverables readable without import in %s', async status => {
 readTransport(); showSlice({ status });
 await screen.findByText(validDetail.raw_output);
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
});
it('hides imports while the parent brief is edited', async () => {
 readTransport('owner'); showDetail();
 await waitFor(() => expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeEnabled());
 await userEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
 expect(screen.queryByRole('form', { name: 'Impor deliverable' })).not.toBeInTheDocument();
 await userEvent.click(screen.getByRole('button', { name: 'Batal edit' }));
 expect(await screen.findByRole('button', { name: 'Impor revisi baru' })).toBeEnabled();
});
it('never enables imports when the parent blocks writes', async () => {
 readTransport(); showSlice({ canWrite: false });
 await screen.findByText(validDetail.raw_output);
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
});

it.each([
 ['blank', '   \n\t'], ['NUL', 'human\0output'], ['ASCII overflow', 'x'.repeat(131073)], ['UTF-8 overflow', '😀'.repeat(32769)],
])('rejects %s raw output without sending it', async (_name, raw) => {
 const fetcher = readTransport(); showSlice();
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 await waitFor(() => expect(input).toBeEnabled());
 fireEvent.change(input, { target: { value: raw } });
 fireEvent.submit(screen.getByRole('form', { name: 'Impor deliverable' }));
 expect(await screen.findByText('Isi output 1–131072 byte UTF-8, tidak boleh hanya spasi atau memuat NUL.')).toBeInTheDocument();
 expect(fetcher.mock.calls.some(([, options]) => options.method === 'POST')).toBe(false);
});
it.each(['x', '😀'.repeat(32768)])('accepts UTF-8 byte boundary inputs without truncation (length %i)', async raw => {
 const fetcher = readTransport(); const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && options.method === 'POST' ? Promise.resolve(new Response(JSON.stringify({ ...invalidDetail, raw_output: raw }), { status: 201 })) : base(requestPath, options));
 showSlice();
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 await waitFor(() => expect(input).toBeEnabled());
 fireEvent.change(input, { target: { value: raw } });
 fireEvent.submit(screen.getByRole('form', { name: 'Impor deliverable' }));
 await waitFor(() => expect(fetcher).toHaveBeenCalledWith(path, expect.objectContaining({ method: 'POST', body: JSON.stringify({ raw_output: raw, version: 7 }) })));
 await screen.findByText(/Revisi impor tersimpan/);
});

it('blocks writes while the snapshot loads and prevents simultaneous submissions with an in-flight guard', async () => {
 let resolveRead!: (response: Response) => void;
 let resolveWrite!: (response: Response) => void;
 const fetcher = vi.fn().mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'GET') return new Promise<Response>(resolve => { resolveRead = resolve; });
  if (requestPath === path && options.method === 'POST') return new Promise<Response>(resolve => { resolveWrite = resolve; });
  return Promise.resolve(new Response(JSON.stringify(validDetail)));
 });
 vi.stubGlobal('fetch', fetcher);
 const onChanged = vi.fn(); showSlice({ onChanged });
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 const form = screen.getByRole('form', { name: 'Impor deliverable' });
 expect(input).toBeDisabled();
 fireEvent.submit(form);
 expect(fetcher.mock.calls).toHaveLength(1);
 await act(async () => resolveRead(new Response(JSON.stringify({ version: 7, items: [validSummary] }))));
 fireEvent.change(input, { target: { value: 'double submit' } });
 act(() => { fireEvent.submit(form); fireEvent.submit(form); });
 expect(input).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Segarkan deliverable' })).toBeDisabled();
 expect(fetcher.mock.calls.filter(([, options]) => options.method === 'POST')).toHaveLength(1);
 await act(async () => resolveWrite(new Response(JSON.stringify(validDetail), { status: 201 })));
 expect(onChanged).toHaveBeenCalledTimes(1);
 expect(input).toBeDisabled();
 await act(async () => resolveRead(new Response(JSON.stringify({ version: 8, items: [validSummary] }))));
 expect(input).toBeEnabled();
});

it('recovers a missing collection endpoint without pretending it is empty or breaking task/comments', async () => {
 const fetcher = readTransport('owner'); const base = fetcher.getMockImplementation()!;
 let fail = true;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && fail ? Promise.resolve(new Response(JSON.stringify({ error: { code: 'not_found', message: 'Endpoint unavailable.' } }), { status: 404 })) : base(requestPath, options));
 showDetail();
 const section = await screen.findByRole('region', { name: 'Deliverable impor manusia' });
 expect(await within(section).findByRole('alert')).toHaveTextContent('Endpoint unavailable.');
 expect(within(section).queryByText('Belum ada deliverable.')).not.toBeInTheDocument();
 expect(within(section).getByRole('textbox')).toBeDisabled();
 expect(screen.getByRole('heading', { name: task.title })).toBeInTheDocument();
 expect(screen.getByRole('region', { name: 'Komentar & klarifikasi' })).toBeInTheDocument();
 fail = false;
 await userEvent.click(within(section).getByRole('button', { name: 'Coba lagi' }));
 expect(await within(section).findByText(validDetail.raw_output)).toBeInTheDocument();
});
it('keeps failed detail loads separate from summaries and permits explicit retry', async () => {
 const fetcher = readTransport(); const base = fetcher.getMockImplementation()!;
 let fail = true;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path + '/' + invalidId && fail ? Promise.resolve(new Response(JSON.stringify({ error: { code: 'not_found', message: 'Revision unavailable.' } }), { status: 404 })) : base(requestPath, options));
 showSlice({ role: 'viewer' });
 await screen.findByText(validDetail.raw_output);
 await userEvent.click(screen.getByRole('button', { name: /Revisi 1 ·/ }));
 expect(await screen.findByRole('alert')).toHaveTextContent('Revision unavailable.');
 expect(screen.queryByText(validDetail.raw_output)).not.toBeInTheDocument();
 expect(screen.queryByText(invalidDetail.raw_output)).not.toBeInTheDocument();
 fail = false;
 await userEvent.click(screen.getByRole('button', { name: 'Coba lagi' }));
 expect(await screen.findByText(invalidDetail.raw_output)).toBeInTheDocument();
});
it('shows server mutation errors without success, retry, or losing the draft', async () => {
 const fetcher = readTransport(); const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => requestPath === path && options.method === 'POST' ? Promise.resolve(new Response(JSON.stringify({ error: { code: 'invalid_input', message: 'Import rejected.' } }), { status: 422 })) : base(requestPath, options));
 const onChanged = vi.fn(); showSlice({ onChanged });
 const input = screen.getByRole('textbox', { name: 'Output mentah untuk impor' });
 await waitFor(() => expect(input).toBeEnabled());
 fireEvent.change(input, { target: { value: 'preserve this draft' } });
 await userEvent.click(screen.getByRole('button', { name: 'Impor revisi baru' }));
 expect(await screen.findByRole('alert')).toHaveTextContent('Import rejected.');
 expect(input).toHaveValue('preserve this draft');
 expect(screen.queryByText(/Revisi impor tersimpan/)).not.toBeInTheDocument();
 expect(onChanged).not.toHaveBeenCalled();
 expect(fetcher.mock.calls.filter(([, options]) => options.method === 'POST')).toHaveLength(1);
});
