import { expect, it, vi } from 'vitest';
import { ApiClient } from './api';

// Transport fixtures only; these tests do not exercise a live backend.
const taskId = 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8';
const dependencyId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const snapshot = { version: 7, items: [{ task_id: dependencyId, title: 'Research prerequisite', status: 'ready' }] };
const task = { id: taskId, brand_id: 'd0e0ddbb-23f8-4fe7-9be7-e53dd92e9860', title: 'Brief', objective: '', instructions: '', priority: 'normal', status: 'assigned', agent_id: 'claudia', version: 8, deadline_at: null, created_at: '2026-11-06T00:00:00Z', updated_at: '2026-11-06T00:00:00Z' };

it('loads a versioned dependency snapshot and adds/removes with JSON version and CSRF', async () => {
 const transport = vi.fn()
  .mockResolvedValueOnce(new Response(JSON.stringify(snapshot)))
  .mockResolvedValueOnce(new Response(JSON.stringify(task)))
  .mockResolvedValueOnce(new Response(JSON.stringify({ ...task, version: 9 })));
 const api = new ApiClient(transport);
 const loaded = await api.taskDependencies(taskId);
 expect(loaded).toEqual(snapshot);
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/' + taskId + '/dependencies', expect.objectContaining({ method: 'GET', credentials: 'same-origin' }));
 const added = await api.addTaskDependency(taskId, { dependency_id: dependencyId, version: loaded.version }, 'fixture-csrf');
 expect(added).toEqual(task);
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/' + taskId + '/dependencies', expect.objectContaining({ method: 'POST', headers: expect.objectContaining({ 'X-CSRF-Token': 'fixture-csrf', 'Content-Type': 'application/json' }), body: JSON.stringify({ dependency_id: dependencyId, version: 7 }) }));
 const removed = await api.removeTaskDependency(taskId, { dependency_id: dependencyId, version: added.version }, 'fixture-csrf');
 expect(removed.version).toBe(9);
 expect(removed.status).toBe('assigned');
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/' + taskId + '/dependencies', expect.objectContaining({ method: 'DELETE', headers: expect.objectContaining({ 'X-CSRF-Token': 'fixture-csrf', 'Content-Type': 'application/json' }), body: JSON.stringify({ dependency_id: dependencyId, version: 8 }) }));
});

it.each([
 ['missing version', { items: snapshot.items }],
 ['zero version', { ...snapshot, version: 0 }],
 ['negative version', { ...snapshot, version: -1 }],
 ['fractional version', { ...snapshot, version: 1.5 }],
 ['noncanonical status', { ...snapshot, items: [{ ...snapshot.items[0], status: 'done' }] }],
 ['invalid task UUID', { ...snapshot, items: [{ ...snapshot.items[0], task_id: 'not-a-uuid' }] }],
 ['nontext title', { ...snapshot, items: [{ ...snapshot.items[0], title: null }] }],
])('rejects a malformed dependency snapshot: %s', async (_name, data) => {
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 await expect(api.taskDependencies(taskId)).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});

it.each(['addTaskDependency', 'removeTaskDependency'] as const)('requires CSRF before transporting %s', async method => {
 const transport = vi.fn();
 const api = new ApiClient(transport);
 await expect(api[method](taskId, { dependency_id: dependencyId, version: 7 }, '')).rejects.toMatchObject({ status: 403, code: 'csrf_missing' });
 expect(transport).not.toHaveBeenCalled();
});

it.each([
 [409, 'version_conflict'], [409, 'dependency_cycle'], [422, 'invalid_dependency'], [404, 'not_found'],
])('propagates dependency mutation rejection %s/%s without automatic retry', async (status, code) => {
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code, message: 'Rejected by server' } }), { status }));
 const api = new ApiClient(transport);
 await expect(api.addTaskDependency(taskId, { dependency_id: dependencyId, version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ status, code });
 expect(transport).toHaveBeenCalledTimes(1);
});

it('does not accept a malformed mutation response as a saved Task', async () => {
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }))));
 await expect(api.removeTaskDependency(taskId, { dependency_id: dependencyId, version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});
