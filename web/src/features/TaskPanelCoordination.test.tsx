import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { expect, it, vi } from 'vitest';
import { App } from '../app/App';
import type { Task } from '../lib/api';
import { brandFixture, fixtureResponse } from '../test/fixtures';

// App transport fixtures only: no live backend, authentication, database or runtime.
const task: Task = { id: 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8', brand_id: 'd0e0ddbb-23f8-4fe7-9be7-e53dd92e9860', title: 'Audience brief', objective: '', instructions: '', priority: 'normal', status: 'draft', version: 3, created_at: '2026-11-06T00:00:00Z', updated_at: '2026-11-06T00:00:00Z' };
const taskPath = '/api/tasks/' + task.id;
const dependenciesPath = taskPath + '/dependencies';
const deliverablesPath = taskPath + '/deliverables';
const candidate = { ...task, id: 'c40a91e8-8f35-43d7-857b-d756c5179b60', title: 'Prerequisite', status: 'completed' };
function response(data: unknown) { return new Response(JSON.stringify(data)); }
function transport() {
 const fetcher = vi.fn().mockImplementation((path: string, _options: RequestInit) => {
  void _options;
  if (path === taskPath) return Promise.resolve(response(task));
  if (path === dependenciesPath || path === deliverablesPath) return Promise.resolve(response({ version: 7, items: [] }));
  if (path === taskPath + '/events' || path === taskPath + '/comments') return Promise.resolve(response({ items: [] }));
  if (path === '/api/brands') return Promise.resolve(response({ items: [{ ...brandFixture, id: task.brand_id }] }));
  if (path === '/api/tasks?brand_id=' + task.brand_id) return Promise.resolve(response({ items: [task, candidate] }));
  return Promise.resolve(response(fixtureResponse(path)));
 });
 vi.stubGlobal('fetch', fetcher);
 return fetcher;
}
function showDetail() { render(<MemoryRouter initialEntries={[ '/tasks/' + task.id ]}><App /></MemoryRouter>); }
async function ready() {
 await waitFor(() => expect(screen.getByRole('button', { name: 'Tambah dependensi' })).toBeDisabled());
 await waitFor(() => expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeEnabled());
 await waitFor(() => expect(screen.getByRole('button', { name: 'Segarkan dependensi' })).toBeEnabled());
}

it.each(['Segarkan dependensi', 'Segarkan deliverable'])('keeps a dirty brief intact when %s is attempted during editing', async refresh => {
 const fetcher = transport(); showDetail(); await ready();
 await userEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
 fireEvent.change(screen.getByLabelText('Judul task *'), { target: { value: 'Unsaved human brief' } });
 await userEvent.click(screen.getByRole('button', { name: refresh }));
 expect(screen.getByLabelText('Judul task *')).toHaveValue('Unsaved human brief');
 expect(screen.getByRole('button', { name: refresh })).toBeDisabled();
 expect(fetcher.mock.calls.filter(([path]) => path === taskPath)).toHaveLength(1);
 expect(fetcher.mock.calls.filter(([path]) => path === taskPath + '/events')).toHaveLength(1);
 await userEvent.click(screen.getByRole('button', { name: 'Simpan perubahan' }));
 await waitFor(() => expect(fetcher).toHaveBeenCalledWith(taskPath, expect.objectContaining({ method: 'PATCH', body: expect.stringContaining('"version":3') })));
});

it.each(['dependency', 'deliverable'] as const)('does not start a brief edit during a pending %s write or its follow-up refresh', async panel => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 const path = panel === 'dependency' ? dependenciesPath : deliverablesPath;
 let resolveWrite!: (value: Response) => void;
 let resolveTask!: (value: Response) => void;
 let resolvePanel!: (value: Response) => void;
 let saved = false;
 fetcher.mockImplementation((requestPath: string, options: RequestInit) => {
  if (requestPath === path && options.method === 'POST') return new Promise<Response>(resolve => { resolveWrite = resolve; });
  if (saved && requestPath === taskPath) return new Promise<Response>(resolve => { resolveTask = resolve; });
  if (saved && requestPath === path) return new Promise<Response>(resolve => { resolvePanel = resolve; });
  return base(requestPath, options);
 });
 showDetail(); await ready();
 const edit = screen.getByRole('button', { name: 'Edit brief' });
 if (panel === 'dependency') {
  await userEvent.selectOptions(screen.getByLabelText('Tambah prasyarat'), candidate.id);
  await userEvent.click(screen.getByRole('button', { name: 'Tambah dependensi' }));
 } else {
  fireEvent.change(screen.getByLabelText('Output mentah untuk impor'), { target: { value: 'Human import' } });
  await userEvent.click(screen.getByRole('button', { name: 'Impor revisi baru' }));
 }
 await userEvent.click(edit);
 expect(screen.queryByLabelText('Judul task *')).not.toBeInTheDocument();
 expect(edit).toBeDisabled();
 expect(screen.getByRole('button', { name: 'Tandai siap' })).toBeDisabled();
 expect(fetcher.mock.calls.filter(([requestPath, options]) => requestPath === path && options.method === 'POST')).toHaveLength(1);
 saved = true;
 const imported = { id: candidate.id, task_id: task.id, revision: 1, author_id: '751c94c2-a256-4f84-bd69-c37621a65af4', origin: 'human_import', validation_status: 'invalid', title: 'Human import', kind: '', raw_output: 'Human import', structured: null, created_at: task.created_at };
 await act(async () => resolveWrite(response(panel === 'dependency' ? { ...task, version: 8 } : imported)));
 expect(screen.queryByLabelText('Judul task *')).not.toBeInTheDocument();
 await act(async () => resolveTask(response({ ...task, version: 8, title: 'Refreshed brief' })));
 await waitFor(() => expect(resolvePanel).toBeTypeOf('function'));
 await act(async () => resolvePanel(response({ version: 8, items: [] })));
 await waitFor(() => expect(screen.getByRole('button', { name: 'Edit brief' })).toBeEnabled());
 await userEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
 expect(screen.getByLabelText('Judul task *')).toHaveValue('Refreshed brief');
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath)).toHaveLength(2);
 expect(fetcher.mock.calls.filter(([requestPath]) => requestPath === taskPath + '/events')).toHaveLength(2);
});

it.each(['Segarkan dependensi', 'Segarkan deliverable'])('blocks edit-start in the same turn as %s', async refresh => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 let refreshing = false;
 let resolveTask!: (value: Response) => void;
 fetcher.mockImplementation((path: string, options: RequestInit) => {
  if (refreshing && path === taskPath) return new Promise<Response>(resolve => { resolveTask = resolve; });
  return base(path, options);
 });
 showDetail(); await ready();
 refreshing = true;
 const edit = screen.getByRole('button', { name: 'Edit brief' });
 act(() => {
  fireEvent.click(screen.getByRole('button', { name: refresh }));
  fireEvent.click(edit);
 });
 expect(screen.queryByLabelText('Judul task *')).not.toBeInTheDocument();
 expect(edit).toBeDisabled();
 await act(async () => resolveTask(response({ ...task, version: 8 })));
 await waitFor(() => expect(screen.getByRole('button', { name: 'Edit brief' })).toBeEnabled());
});

it.each(['dependency refresh', 'brief save'] as const)('preserves an unsent raw import across an unrelated %s', async change => {
 const fetcher = transport(); const base = fetcher.getMockImplementation()!;
 let refreshing = false;
 let resolveTask!: (value: Response) => void;
 fetcher.mockImplementation((path: string, options: RequestInit) => {
  if (refreshing && path === taskPath && options.method === 'GET') return new Promise<Response>(resolve => { resolveTask = resolve; });
  if (path === taskPath && options.method === 'PATCH') return Promise.resolve(response({ ...task, version: 8 }));
  return base(path, options);
 });
 showDetail(); await ready();
 const raw = '  Unsent human output 😀\nDo not discard.  ';
 fireEvent.change(screen.getByLabelText('Output mentah untuk impor'), { target: { value: raw } });
 refreshing = true;
 if (change === 'dependency refresh') await userEvent.click(screen.getByRole('button', { name: 'Segarkan dependensi' }));
 else {
  await userEvent.click(screen.getByRole('button', { name: 'Edit brief' }));
  fireEvent.change(screen.getByLabelText('Judul task *'), { target: { value: 'Saved brief' } });
  await userEvent.click(screen.getByRole('button', { name: 'Simpan perubahan' }));
 }
 await waitFor(() => expect(resolveTask).toBeTypeOf('function'));
 await act(async () => resolveTask(response({ ...task, version: 8 })));
 await waitFor(() => expect(screen.getByRole('button', { name: 'Impor revisi baru' })).toBeEnabled());
 expect(screen.getByLabelText('Output mentah untuk impor')).toHaveValue(raw);
 expect(fetcher.mock.calls.filter(([path, options]) => path === taskPath && options.method === 'GET')).toHaveLength(2);
 expect(fetcher.mock.calls.filter(([path]) => path === deliverablesPath)).toHaveLength(1);
});
