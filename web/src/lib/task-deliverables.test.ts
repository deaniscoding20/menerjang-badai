import { expect, it, vi } from 'vitest';
import { ApiClient } from './api';

// Transport fixtures only; no live backend, runtime or database acceptance.
const taskId = 'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8';
const deliverableId = '751c94c2-a256-4f84-bd69-c37621a65af4';
const summary = { id: deliverableId, task_id: taskId, revision: 2, author_id: 'c40a91e8-8f35-43d7-857b-d756c5179b60', origin: 'human_import', validation_status: 'validated', title: 'Audience report', kind: 'research_report', created_at: '2026-11-06T00:00:00Z' };
const structured = { schema_version: 1, kind: 'research_report', title: 'Audience report', sections: [{ heading: 'Finding', body: '<script>Human review required</script>' }], sources: [{ label: 'Reference', url: 'https://example.test/report' }] };
const detail = { ...summary, raw_output: JSON.stringify(structured), structured };
const path = '/api/tasks/' + taskId + '/deliverables';

it('loads a versioned summary collection and the selected immutable detail separately', async () => {
 const snapshot = { version: 7, items: [summary] };
 const transport = vi.fn()
  .mockResolvedValueOnce(new Response(JSON.stringify(snapshot)))
  .mockResolvedValueOnce(new Response(JSON.stringify(detail)));
 const api = new ApiClient(transport);
 expect(await api.taskDeliverables(taskId)).toEqual(snapshot);
 expect(transport).toHaveBeenLastCalledWith(path, expect.objectContaining({ method: 'GET', credentials: 'same-origin' }));
 expect(await api.taskDeliverable(taskId, deliverableId)).toEqual(detail);
 expect(transport).toHaveBeenLastCalledWith(path + '/' + deliverableId, expect.objectContaining({ method: 'GET', credentials: 'same-origin' }));
});

it('imports raw text with the collection snapshot version and same-origin CSRF, returning the immutable detail', async () => {
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify(detail), { status: 201 }));
 const api = new ApiClient(transport);
 const input = { raw_output: '  arbitrary text 😀  ', version: 7 };
 expect(await api.importTaskDeliverable(taskId, input, 'fixture-csrf')).toEqual(detail);
 expect(transport).toHaveBeenCalledExactlyOnceWith(path, expect.objectContaining({ method: 'POST', credentials: 'same-origin', headers: expect.objectContaining({ 'X-CSRF-Token': 'fixture-csrf', 'Content-Type': 'application/json' }), body: JSON.stringify(input) }));
});

it.each<[string, unknown]>([
 ['missing version', { items: [summary] }], ['zero version', { version: 0, items: [summary] }],
 ['fractional version', { version: 1.5, items: [summary] }],
 ...['id', 'task_id', 'author_id'].map(field => ['invalid UUID ' + field, { version: 7, items: [{ ...summary, [field]: 'not-a-uuid' }] }] as [string, unknown]),
 ['zero revision', { version: 7, items: [{ ...summary, revision: 0 }] }],
 ['bad timestamp', { version: 7, items: [{ ...summary, created_at: 'yesterday' }] }],
 ['agent provenance', { version: 7, items: [{ ...summary, origin: 'agent' }] }],
 ['approval status', { version: 7, items: [{ ...summary, validation_status: 'approved' }] }],
 ['invalid with kind', { version: 7, items: [{ ...summary, validation_status: 'invalid' }] }],
 ['more than latest 100', { version: 7, items: Array.from({ length: 101 }, () => summary) }],
])('rejects malformed collection metadata: %s', async (_name, data) => {
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 await expect(api.taskDeliverables(taskId)).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});

it('accepts invalid imports only with null structured output and preserves their raw text', async () => {
 const invalid = { ...summary, validation_status: 'invalid', kind: '', raw_output: '<script>plain text</script>', structured: null };
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(invalid))));
 expect(await api.taskDeliverable(taskId, deliverableId)).toEqual(invalid);
});
it.each<[string, unknown]>([
 ['invalid with structure', { ...detail, validation_status: 'invalid', kind: '' }],
 ['validated without structure', { ...detail, structured: null }],
 ['missing raw output', { ...detail, raw_output: undefined }],
 ['unsupported version', { ...detail, structured: { ...structured, schema_version: 2 } }],
 ['unknown kind', { ...detail, structured: { ...structured, kind: 'agent_log' } }],
 ['empty sections', { ...detail, structured: { ...structured, sections: [] } }],
 ['blank heading', { ...detail, structured: { ...structured, sections: [{ heading: ' ', body: 'text' }] } }],
 ['NUL body', { ...detail, structured: { ...structured, sections: [{ heading: 'Finding', body: 'text\0' }] } }],
 ['unknown schema field', { ...detail, structured: { ...structured, auto_approved: true } }],
 ['unknown section field', { ...detail, structured: { ...structured, sections: [{ heading: 'Finding', body: 'Evidence', auto_approved: true }] } }],
 ['unknown source field', { ...detail, structured: { ...structured, sources: [{ label: 'Reference', url: 'https://example.com/', token: 'secret' }] } }],
 ['mismatched title', { ...detail, title: 'different' }],
 ...['javascript:alert(1)', 'http://example.test', 'https://user:password@example.test', 'https://@example.test', 'https://example.test?token=secret'].map(url => ['unsafe source ' + url, { ...detail, structured: { ...structured, sources: [{ label: 'Source', url }] } }] as [string, unknown]),
])('rejects malformed or inconsistent detail without substituting output: %s', async (_name, data) => {
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 await expect(api.taskDeliverable(taskId, deliverableId)).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});
it.each(['research_report', 'marketing_strategy', 'creative_brief', 'caption_package', 'design_specification', 'campaign_plan', 'performance_report'])('accepts structured v1 kind %s', async kind => {
 const data = { ...detail, kind, structured: { ...structured, kind } };
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 expect(await api.taskDeliverable(taskId, deliverableId)).toEqual(data);
});
it('accepts an at-sign in an HTTPS source path, which is not embedded credentials', async () => {
 const data = { ...detail, structured: { ...structured, sources: [{ label: 'Public reference', url: 'https://example.test/people/@human' }] } };
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 expect(await api.taskDeliverable(taskId, deliverableId)).toEqual(data);
});
it.each([200, 201])('accepts and preserves a Go-valid empty source query separator in a %s API response', async status => {
 const output = { ...structured, sources: [{ label: 'Public reference', url: 'https://example.com/?' }] };
 const data = { ...detail, raw_output: JSON.stringify(output), structured: output };
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify(data), { status }));
 const api = new ApiClient(transport);
 const response = status === 201
  ? api.importTaskDeliverable(taskId, { raw_output: data.raw_output, version: 7 }, 'fixture-csrf')
  : api.taskDeliverable(taskId, deliverableId);
 await expect(response).resolves.toEqual(data);
 expect(transport).toHaveBeenCalledTimes(1); // Source URLs are plain text, never fetched.
});
it.each<[string, string, boolean]>([
 ['empty query before fragment', 'https://example.com/?#reference', true],
 ['question mark inside fragment', 'https://example.com/#reference?token=public-text', true],
 ['UTF-8 byte limit', 'https://example.com/' + 'é'.repeat(1014), true],
 ['over UTF-8 byte limit', 'https://example.com/' + 'é'.repeat(1014) + 'x', false],
 ['query key', 'https://example.com/?q', false],
 ['query token', 'https://example.com/?token=secret', false],
 ['query separator as content', 'https://example.com/??', false],
 ['query whitespace normalized away by URL', 'https://example.com/?\t', false],
 ['path control normalized away by URL', 'https://example.com/\n?', false],
 ['authority control normalized away by URL', 'https://exam\tple.com/?', false],
 ['credentials', 'https://user:password@example.com/?', false],
 ['empty credentials', 'https://@example.com/?', false],
 ['NUL', 'https://example.com/\0?', false],
])('checks the raw Go source contract without normalizing stored text: %s', async (_name, url, accepted) => {
 const output = { ...structured, sources: [{ label: 'Public reference', url }] };
 const data = { ...detail, raw_output: JSON.stringify(output), structured: output };
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify(data)));
 const response = new ApiClient(transport).taskDeliverable(taskId, deliverableId);
 if (accepted) await expect(response).resolves.toEqual(data);
 else await expect(response).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
 expect(transport).toHaveBeenCalledTimes(1);
});
it.each([200, 201])('accepts and preserves a Go-nonblank U+FEFF-only section body in a %s API response', async status => {
 const output = { ...structured, sections: [{ heading: 'Finding', body: '\uFEFF' }] };
 const data = { ...detail, raw_output: JSON.stringify(output), structured: output };
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify(data), { status }));
 const api = new ApiClient(transport);
 const response = status === 201
  ? api.importTaskDeliverable(taskId, { raw_output: data.raw_output, version: 7 }, 'fixture-csrf')
  : api.taskDeliverable(taskId, deliverableId);
 await expect(response).resolves.toEqual(data);
 expect(transport).toHaveBeenCalledTimes(1);
});
it.each((['title', 'heading', 'body', 'label'] as const).flatMap(field => ([
 ['U+FEFF', '\uFEFF', true],
 ['U+0085', '\u0085', false],
 ['all Go whitespace', ' \t\n\r\v\f\u0085\u00A0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u2028\u2029\u202F\u205F\u3000', false],
] as const).map(([name, value, accepted]) => ({ field, name, value, accepted }))))('uses Go nonblank semantics for $name in $field without trimming response text', async ({ field, value, accepted }) => {
 const output = {
  ...structured,
  title: field === 'title' ? value : structured.title,
  sections: [{ heading: field === 'heading' ? value : 'Finding', body: field === 'body' ? value : 'Evidence' }],
  sources: [{ label: field === 'label' ? value : 'Reference', url: 'https://example.com/' }],
 };
 const data = { ...detail, title: output.title, raw_output: JSON.stringify(output), structured: output };
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 if (accepted) await expect(api.taskDeliverable(taskId, deliverableId)).resolves.toEqual(data);
 else await expect(api.taskDeliverable(taskId, deliverableId)).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});
it.each(([
 ['title', 200], ['heading', 160], ['body', 12000], ['label', 200],
] as const).flatMap(([field, max]) => ([
 ['at codepoint limit', '😀'.repeat(max), true],
 ['over codepoint limit', '😀'.repeat(max + 1), false],
 ['NUL', 'text\0', false],
] as const).map(([name, value, accepted]) => ({ field, name, value, accepted }))))('keeps $field $name validation strict', async ({ field, value, accepted }) => {
 const output = {
  ...structured,
  title: field === 'title' ? value : structured.title,
  sections: [{ heading: field === 'heading' ? value : 'Finding', body: field === 'body' ? value : 'Evidence' }],
  sources: [{ label: field === 'label' ? value : 'Reference', url: 'https://example.com/' }],
 };
 const data = { ...detail, title: output.title, raw_output: JSON.stringify(output), structured: output };
 const api = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(data))));
 if (accepted) await expect(api.taskDeliverable(taskId, deliverableId)).resolves.toEqual(data);
 else await expect(api.taskDeliverable(taskId, deliverableId)).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
});
it('requires CSRF before transporting an import', async () => {
 const transport = vi.fn();
 const api = new ApiClient(transport);
 await expect(api.importTaskDeliverable(taskId, { raw_output: 'human text', version: 7 }, '')).rejects.toMatchObject({ status: 403, code: 'csrf_missing' });
 expect(transport).not.toHaveBeenCalled();
});
it.each([[409, 'version_conflict'], [422, 'invalid_input'], [403, 'forbidden'], [404, 'not_found']])('propagates %s/%s without retrying an import', async (status, code) => {
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code, message: 'Rejected by server' } }), { status: Number(status) }));
 const api = new ApiClient(transport);
 await expect(api.importTaskDeliverable(taskId, { raw_output: 'human text', version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ status, code });
 expect(transport).toHaveBeenCalledTimes(1);
});
it('does not claim success for a malformed import response', async () => {
 const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 201 }));
 const api = new ApiClient(transport);
 await expect(api.importTaskDeliverable(taskId, { raw_output: 'human text', version: 7 }, 'fixture-csrf')).rejects.toMatchObject({ status: 502, code: 'invalid_response' });
 expect(transport).toHaveBeenCalledTimes(1);
});
