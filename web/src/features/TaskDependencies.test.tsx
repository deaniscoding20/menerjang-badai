import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { WorkspaceContext } from '../app/WorkspaceContext';
import { TaskDependencies } from './TaskDependencies';
import { expect, it, vi } from 'vitest';
import { App } from '../app/App';
import { brandFixture, fixtureResponse, sessionFixture } from '../test/fixtures';
import type { Task, TaskStatus } from '../lib/api';

// UI transport fixtures only: no live authentication, runtime or database acceptance.
const task: Task = { id: 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8', brand_id: 'd0e0ddbb-23f8-4fe7-9be7-e53dd92e9860', title: 'Audience brief', objective: '', instructions: '', priority: 'normal', status: 'draft', agent_id: '', version: 3, deadline_at: null, created_at: '2026-11-06T00:00:00Z', updated_at: '2026-11-06T00:00:00Z' };
const dependencyId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const dependencyPath = '/api/tasks/' + task.id + '/dependencies';
type Dependency = { task_id: string; title: string; status: TaskStatus };
const existing: Dependency = { task_id: dependencyId, title: '<script>Prerequisite</script>', status: 'ready' };
function showDetail() { render(<MemoryRouter initialEntries={['/tasks/' + task.id]}><App /></MemoryRouter>); }
function readTransport(role = 'viewer', items: Dependency[] = [existing]) {
 const fetcher = vi.fn().mockImplementation((path: string) => {
  let data: unknown;
  if (path === '/api/session') data = { ...sessionFixture, user: { ...sessionFixture.user, role } };
  else if (path === dependencyPath) data = { version: 7, items };
  else if (path === '/api/tasks/' + task.id) data = task;
  else if (path === '/api/tasks/' + task.id + '/events' || path === '/api/tasks/' + task.id + '/comments') data = { items: [] };
  else if (path === '/api/brands') data = { items: [{ ...brandFixture, id: task.brand_id }] };
  else data = fixtureResponse(path);
  return Promise.resolve(new Response(JSON.stringify(data)));
 });
 vi.stubGlobal('fetch', fetcher);
 return fetcher;
}

it('shows escaped dependency titles above comments with only completed prerequisites satisfied for a viewer', async () => {
 const statuses: TaskStatus[] = ['draft', 'ready', 'assigned', 'queued', 'running', 'waiting_input', 'waiting_approval', 'completed', 'failed', 'cancelled'];
 const fetcher = readTransport('viewer', statuses.map((status, index) => ({ task_id: `751c94c2-a256-4f84-bd69-c37621a65a${String(index).padStart(2, '0')}`, title: index === 0 ? existing.title : 'Prerequisite ' + status, status })));
 showDetail();
 const section = await screen.findByRole('region', { name: 'Dependensi task' });
 expect(await within(section).findByText(existing.title)).toBeInTheDocument();
 expect(document.querySelector('script')).toBeNull();
 expect(within(section).getAllByText('Prasyarat belum terpenuhi')).toHaveLength(9);
 expect(within(section).getAllByText('Prasyarat terpenuhi')).toHaveLength(1);
 expect(within(section).queryByRole('form')).not.toBeInTheDocument();
 expect(within(section).queryByRole('button', { name: /Hapus dependensi/ })).not.toBeInTheDocument();
 expect(fetcher.mock.calls.some(([path]) => String(path).startsWith('/api/tasks?'))).toBe(false);
 const comments = screen.getByRole('region', { name: 'Komentar & klarifikasi' });
 expect(section.compareDocumentPosition(comments) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
 expect(screen.getByRole('button', { name: /Jalankan task/ })).toBeDisabled();
});

const candidateId = 'c40a91e8-8f35-43d7-857b-d756c5179b60';
const candidate: Task = { ...task, id: candidateId, title: 'Research candidate', status: 'completed' };
function writableTransport() {
 let version = 7;
 let items = [existing];
 const fetcher = vi.fn().mockImplementation((path: string, options: RequestInit) => {
  let data: unknown;
  if (path === '/api/session') data = sessionFixture;
  else if (path === dependencyPath) {
   if (options.method !== 'GET') {
    const input = JSON.parse(String(options.body));
    items = options.method === 'POST' ? [...items, { task_id: input.dependency_id, title: candidate.title, status: candidate.status }] : items.filter(item => item.task_id !== input.dependency_id);
    version += 1;
    data = { ...task, version };
   } else data = { version, items };
  } else if (path === '/api/tasks?brand_id=' + task.brand_id) data = { items: [task, { ...task, id: dependencyId, title: existing.title }, candidate, { ...candidate, id: 'd5be6a58-cd9e-4d8d-9136-e7b8925cdca3', brand_id: 'other-brand', title: 'Wrong brand' }] };
  else if (path === '/api/tasks/' + task.id) data = task;
  else if (path === '/api/tasks/' + task.id + '/events' || path === '/api/tasks/' + task.id + '/comments') data = { items: [] };
  else if (path === '/api/brands') data = { items: [{ ...brandFixture, id: task.brand_id }] };
  else data = fixtureResponse(path);
  return Promise.resolve(new Response(JSON.stringify(data)));
 });
 vi.stubGlobal('fetch', fetcher);
 return fetcher;
}
it('adds and removes with the dependency snapshot version and refreshes parent/history without dispatch', async () => {
 const fetcher = writableTransport();
 showDetail();
 const select = await screen.findByLabelText('Tambah prasyarat');
 await waitFor(() => expect(select).toBeEnabled());
 expect(within(select).queryByRole('option', { name: task.title })).not.toBeInTheDocument();
 expect(within(select).queryByRole('option', { name: existing.title })).not.toBeInTheDocument();
 expect(within(select).queryByRole('option', { name: /Wrong brand/ })).not.toBeInTheDocument();
 expect(screen.getByText(/500 task terbaru/)).toBeInTheDocument();
 await userEvent.selectOptions(select, candidateId);
 await userEvent.click(screen.getByRole('button', { name: 'Tambah dependensi' }));
 await screen.findByRole('link', { name: candidate.title });
 expect(fetcher).toHaveBeenCalledWith(dependencyPath, expect.objectContaining({ method: 'POST', headers: expect.objectContaining({ 'X-CSRF-Token': sessionFixture.csrf_token }), body: JSON.stringify({ dependency_id: candidateId, version: 7 }) }));
 await userEvent.click(screen.getByRole('button', { name: 'Hapus dependensi ' + candidate.title }));
 await screen.findByRole('link', { name: existing.title });
 expect(screen.queryByRole('link', { name: candidate.title })).not.toBeInTheDocument();
 expect(fetcher).toHaveBeenCalledWith(dependencyPath, expect.objectContaining({ method: 'DELETE', headers: expect.objectContaining({ 'X-CSRF-Token': sessionFixture.csrf_token }), body: JSON.stringify({ dependency_id: candidateId, version: 8 }) }));
 expect(fetcher.mock.calls.filter(([path]) => path === '/api/tasks/' + task.id)).toHaveLength(3);
 expect(fetcher.mock.calls.filter(([path]) => path === '/api/tasks/' + task.id + '/events')).toHaveLength(3);
 expect(fetcher.mock.calls.filter(([path, options]) => path === dependencyPath && options.method === 'GET')).toHaveLength(3);
 expect(fetcher.mock.calls.some(([path]) => path === '/api/tasks' || String(path).includes('/transition'))).toBe(false);
});

it('requires an explicit reload after a version conflict and never automatically retries the mutation', async () => {
 let version = 7;
 let writes = 0;
 const fetcher = vi.fn().mockImplementation((path: string, options: RequestInit) => {
  if (path === dependencyPath && options.method === 'POST') {
   writes += 1;
   version = 8;
   return Promise.resolve(new Response(JSON.stringify({ error: { code: 'version_conflict', message: 'Task changed in another session.' } }), { status: 409 }));
  }
  const data = path === dependencyPath ? { version, items: [existing] } : { items: [task, candidate] };
  return Promise.resolve(new Response(JSON.stringify(data)));
 });
 vi.stubGlobal('fetch', fetcher);
 const onChanged = vi.fn();
 const onError = vi.fn();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: sessionFixture, onError }}><TaskDependencies task={task} canWrite onChanged={onChanged} /></WorkspaceContext.Provider></MemoryRouter>);
 const select = await screen.findByLabelText('Tambah prasyarat');
 await waitFor(() => expect(select).toBeEnabled());
 await userEvent.selectOptions(select, candidateId);
 await userEvent.click(screen.getByRole('button', { name: 'Tambah dependensi' }));
 expect(await screen.findByText(/Versi task berubah/)).toBeInTheDocument();
 expect(screen.getByRole('button', { name: 'Tambah dependensi' })).toBeDisabled();
 expect(writes).toBe(1);
 expect(onError).toHaveBeenCalledWith(expect.objectContaining({ code: 'version_conflict' }));
 await userEvent.click(screen.getByRole('button', { name: 'Muat ulang task dan dependensi' }));
 await waitFor(() => expect(select).toBeEnabled());
 expect(onChanged).toHaveBeenCalledTimes(1);
 expect(writes).toBe(1);
 await userEvent.click(screen.getByRole('button', { name: 'Tambah dependensi' }));
 await waitFor(() => expect(writes).toBe(2));
 expect(fetcher).toHaveBeenLastCalledWith(dependencyPath, expect.objectContaining({ method: 'POST', body: JSON.stringify({ dependency_id: candidateId, version: 8 }) }));
});

it.each(['owner', 'administrator', 'marketing_manager', 'editor'])('allows the canonical write role %s to manage draft dependencies', async role => {
 writableTransport();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: { ...sessionFixture, user: { ...sessionFixture.user, role } }, onError: vi.fn() }}><TaskDependencies task={task} canWrite onChanged={vi.fn()} /></WorkspaceContext.Provider></MemoryRouter>);
 expect(await screen.findByRole('button', { name: 'Hapus dependensi ' + existing.title })).toBeEnabled();
 expect(await screen.findByRole('form', { name: 'Tambah dependensi' })).toBeInTheDocument();
});

it.each(['viewer', 'analyst', 'approver'])('never renders write controls or requests candidates for readonly %s even if the caller enables writes', async role => {
 const fetcher = readTransport();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: { ...sessionFixture, user: { ...sessionFixture.user, role } }, onError: vi.fn() }}><TaskDependencies task={task} canWrite onChanged={vi.fn()} /></WorkspaceContext.Provider></MemoryRouter>);
 await screen.findByRole('link', { name: existing.title });
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', { name: /Hapus dependensi/ })).not.toBeInTheDocument();
 expect(fetcher.mock.calls.some(([path]) => String(path).startsWith('/api/tasks?'))).toBe(false);
});

it.each(['draft', 'ready', 'assigned'] as const)('permits dependency changes in %s', async status => {
 writableTransport();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: sessionFixture, onError: vi.fn() }}><TaskDependencies task={{ ...task, status }} canWrite onChanged={vi.fn()} /></WorkspaceContext.Provider></MemoryRouter>);
 expect(await screen.findByRole('button', { name: 'Hapus dependensi ' + existing.title })).toBeEnabled();
});

it.each(['queued', 'running', 'waiting_input', 'waiting_approval', 'completed', 'failed', 'cancelled'] as const)('renders dependencies readonly in %s', async status => {
 const fetcher = readTransport();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: sessionFixture, onError: vi.fn() }}><TaskDependencies task={{ ...task, status }} canWrite onChanged={vi.fn()} /></WorkspaceContext.Provider></MemoryRouter>);
 await screen.findByRole('link', { name: existing.title });
 expect(screen.queryByRole('form')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', { name: /Hapus dependensi/ })).not.toBeInTheDocument();
 expect(fetcher.mock.calls.some(([path]) => String(path).startsWith('/api/tasks?'))).toBe(false);
});

it('hides dependency write controls while the parent brief is being edited', async () => {
 writableTransport(); showDetail();
 await screen.findByRole('button', { name: 'Hapus dependensi ' + existing.title });
 await userEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
 expect(screen.queryByRole('button', { name: /Hapus dependensi/ })).not.toBeInTheDocument();
 expect(screen.queryByRole('form', { name: 'Tambah dependensi' })).not.toBeInTheDocument();
 await userEvent.click(screen.getByRole('button', { name: 'Batal edit' }));
 expect(await screen.findByRole('button', { name: 'Hapus dependensi ' + existing.title })).toBeEnabled();
});

it('shows a recoverable missing dependency endpoint without hiding the task or comments', async () => {
 const fetcher = readTransport();
 const base = fetcher.getMockImplementation()!;
 let fail = true;
 fetcher.mockImplementation((path: string) => path === dependencyPath && fail ? Promise.resolve(new Response(JSON.stringify({ error: { code: 'not_found', message: 'Endpoint not available' } }), { status: 404 })) : base(path));
 showDetail();
 const section = await screen.findByRole('region', { name: 'Dependensi task' });
 expect(await within(section).findByRole('alert')).toHaveTextContent('Endpoint not available');
 expect(screen.getByRole('heading', { name: task.title })).toBeInTheDocument();
 expect(screen.getByRole('region', { name: 'Komentar & klarifikasi' })).toBeInTheDocument();
 fail = false;
 await userEvent.click(within(section).getByRole('button', { name: 'Coba lagi' }));
 expect(await within(section).findByRole('link', { name: existing.title })).toBeInTheDocument();
});

it('keeps a server-rejected dependency cycle visible as an error with no success or automatic retry', async () => {
 const fetcher = writableTransport();
 const base = fetcher.getMockImplementation()!;
 fetcher.mockImplementation((path: string, options: RequestInit) => path === dependencyPath && options.method === 'POST' ? Promise.resolve(new Response(JSON.stringify({ error: { code: 'dependency_cycle', message: 'Dependensi membentuk siklus.' } }), { status: 409 })) : base(path, options));
 showDetail();
 const select = await screen.findByLabelText('Tambah prasyarat');
 await waitFor(() => expect(select).toBeEnabled());
 await userEvent.selectOptions(select, candidateId);
 await userEvent.click(screen.getByRole('button', { name: 'Tambah dependensi' }));
 expect(await screen.findByText('Dependensi membentuk siklus.')).toBeInTheDocument();
 expect(screen.queryByText(/Dependensi tersimpan/)).not.toBeInTheDocument();
 expect(fetcher.mock.calls.filter(([path, options]) => path === dependencyPath && options.method === 'POST')).toHaveLength(1);
 expect(screen.queryByRole('link', { name: candidate.title })).not.toBeInTheDocument();
});

it('blocks writes while the dependency snapshot loads and while one mutation is pending', async () => {
 let resolveRead!: (response: Response) => void;
 let resolveWrite!: (response: Response) => void;
 const fetcher = vi.fn().mockImplementation((path: string, options: RequestInit) => {
  if (path === dependencyPath && options.method === 'GET') return new Promise<Response>(resolve => { resolveRead = resolve; });
  if (path === dependencyPath && options.method === 'POST') return new Promise<Response>(resolve => { resolveWrite = resolve; });
  return Promise.resolve(new Response(JSON.stringify({ items: [task, candidate] })));
 });
 vi.stubGlobal('fetch', fetcher);
 const onChanged = vi.fn();
 render(<MemoryRouter><WorkspaceContext.Provider value={{ session: sessionFixture, onError: vi.fn() }}><TaskDependencies task={task} canWrite onChanged={onChanged} /></WorkspaceContext.Provider></MemoryRouter>);
 const select = screen.getByLabelText('Tambah prasyarat');
 expect(select).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Tambah dependensi' })).toBeDisabled();
 resolveRead(new Response(JSON.stringify({ version: 7, items: [existing] })));
 await waitFor(() => expect(select).toBeEnabled());
 await userEvent.selectOptions(select, candidateId);
 const add = screen.getByRole('button', { name: 'Tambah dependensi' });
 await userEvent.click(add);
 expect(add).toBeDisabled();
 expect(select).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Hapus dependensi ' + existing.title })).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Segarkan dependensi' })).toBeDisabled();
 await userEvent.click(add);
 expect(fetcher.mock.calls.filter(([path, options]) => path === dependencyPath && options.method === 'POST')).toHaveLength(1);
 resolveWrite(new Response(JSON.stringify({ ...task, version: 8 })));
 await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
 expect(select).toBeDisabled();
 resolveRead(new Response(JSON.stringify({ version: 8, items: [existing, { task_id: candidateId, title: candidate.title, status: candidate.status }] })));
 expect(await screen.findByRole('link', { name: candidate.title })).toBeInTheDocument();
});
