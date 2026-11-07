import { expect, it, vi } from 'vitest';
import { ApiClient, taskRevisionRequestInputSchema } from './api';

// Actual HTTP-status transport fixtures, not live backend or runtime acceptance.
const taskId = 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8';
const deliverableId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const item = { id: '3426dd91-369c-4c51-8e4c-559d82997b22', task_id: taskId, deliverable_id: deliverableId, deliverable_revision: 2, author_id: 'c40a91e8-8f35-43d7-857b-d756c5179b60', body: '<script>Human feedback</script> https://example.test', created_at: '2026-11-06T00:00:00Z' };
const path = '/api/tasks/' + taskId + '/revision-requests';

it('reads the versioned latest requests with a real 200 schema and appends an immutable record with a 201 schema', async () => {
 const snapshot = { version: 7, items: [item] };
 const transport = vi.fn().mockResolvedValueOnce(new Response(JSON.stringify(snapshot), { status: 200 })).mockResolvedValueOnce(new Response(JSON.stringify(item), { status: 201 }));
 const api = new ApiClient(transport);
 expect(await api.taskRevisionRequests(taskId)).toEqual(snapshot);
 expect(transport).toHaveBeenLastCalledWith(path, expect.objectContaining({ method: 'GET', credentials: 'same-origin' }));
 const input = { deliverable_id: deliverableId, body: 'human request', version: 7 };
 expect(await api.requestTaskRevision(taskId, input, 'fixture-csrf')).toEqual(item);
 expect(transport).toHaveBeenLastCalledWith(path, expect.objectContaining({ method: 'POST', credentials: 'same-origin', headers: expect.objectContaining({ 'X-CSRF-Token': 'fixture-csrf', 'Content-Type': 'application/json' }), body: JSON.stringify(input) }));
 expect(transport).toHaveBeenCalledTimes(2);
});

const allWhitespace = ' \t\n\r\v\f\u0085\u00A0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u2028\u2029\u202F\u205F\u3000';
it.each<[string, string, boolean]>([
 ['empty', '', false], ['Go whitespace', allWhitespace, false], ['NEL', '\u0085', false], ['FEFF', '\uFEFF', true],
 ['NUL', 'request\0', false], ['one codepoint', '😀', true], ['4000 codepoints', '😀'.repeat(4000), true], ['4001 codepoints', '😀'.repeat(4001), false],
])('validates %s using Go whitespace and Unicode codepoints in input and actual 200/201 responses', async (_name, body, accepted) => {
 const input = { deliverable_id: deliverableId, body, version: 7 };
 expect(taskRevisionRequestInputSchema.safeParse(input).success).toBe(accepted);
 for (const status of [200, 201]) {
  const data = { ...item, body };
  const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(status === 200 ? { version: 7, items: [data] } : data), { status })));
  const response = status === 200 ? api.taskRevisionRequests(taskId) : api.requestTaskRevision(taskId, input, 'fixture-csrf');
  if (accepted) await expect(response).resolves.toEqual(status === 200 ? { version: 7, items: [data] } : data);
  else await expect(response).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
 }
});
it('trims Go White_Space before checking the input codepoint limit but preserves response text', () => {
 const input = { deliverable_id: deliverableId, body: '\u0085' + '😀'.repeat(4000) + '\u0085', version: 7 };
 expect(taskRevisionRequestInputSchema.parse(input)).toEqual({ ...input, body: '😀'.repeat(4000) });
});
it.each<[string, unknown]>([
 ['missing version', { items: [item] }], ['zero version', { version: 0, items: [] }], ['fractional version', { version: 1.5, items: [] }],
 ['more than latest 100', { version: 7, items: Array.from({ length: 101 }, () => item) }],
 ...['id', 'task_id', 'deliverable_id', 'author_id'].map(field => ['invalid ' + field, { version: 7, items: [{ ...item, [field]: 'bad-id' }] }] as [string, unknown]),
 ['zero revision', { version: 7, items: [{ ...item, deliverable_revision: 0 }] }], ['fractional revision', { version: 7, items: [{ ...item, deliverable_revision: 1.5 }] }],
 ['bad time', { version: 7, items: [{ ...item, created_at: 'yesterday' }] }], ['unknown field', { version: 7, items: [{ ...item, approved: true }] }],
])('rejects malformed collection %s without an empty fallback', async (_name, data) => {
 await expect(new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data), { status: 200 }))).taskRevisionRequests(taskId)).rejects.toMatchObject({ code: 'invalid_response', status: 502 });
});
it('requires CSRF before sending a request', async () => {
 const transport = vi.fn();
 await expect(new ApiClient(transport).requestTaskRevision(taskId, { deliverable_id: deliverableId, body: 'request', version: 7 }, '')).rejects.toMatchObject({ status: 403, code: 'csrf_missing' });
 expect(transport).not.toHaveBeenCalled();
});
it.each([[409, 'version_conflict'], [422, 'invalid_input'], [403, 'forbidden'], [404, 'not_found']])('propagates %s/%s without retry', async (status, code) => {
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code, message: 'Rejected.' } }), { status: Number(status) }));
 await expect(new ApiClient(transport).requestTaskRevision(taskId, { deliverable_id: deliverableId, body: 'request', version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ status, code });
 expect(transport).toHaveBeenCalledTimes(1);
});
it('does not accept malformed 201 success', async () => {
 await expect(new ApiClient(vi.fn().mockResolvedValue(new Response('{"ok":true}', { status: 201 }))).requestTaskRevision(taskId, { deliverable_id: deliverableId, body: 'request', version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ code: 'invalid_response' });
});
