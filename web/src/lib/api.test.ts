import { describe, expect, it, vi } from 'vitest';
import { ApiClient, ApiError } from './api';
// Explicit frontend transport fixtures. They do not authenticate against the backend.
describe('same-origin API boundary', () => {
 it('reads a validated session without persisting bearer credentials', async () => {
  const transport = vi.fn().mockResolvedValue(new Response(JSON.stringify({user:{id:'u1',display_name:'Nadia',role:'owner'},csrf_token:'frontend-fixture-csrf'})));
  const client = new ApiClient(transport);
  expect((await client.session()).user.display_name).toBe('Nadia');
  expect(transport).toHaveBeenCalledWith('/api/session', expect.objectContaining({credentials:'same-origin',redirect:'error',cache:'no-store'}));
 });
 it('sends CSRF only on same-origin mutations and maps authorization failures', async () => {
  const transport = vi.fn().mockResolvedValueOnce(new Response(null,{status:204})).mockResolvedValueOnce(new Response(JSON.stringify({error:{code:'forbidden',message:'Tidak diizinkan.'}}),{status:403}));
  const client = new ApiClient(transport);
  await client.logout('frontend-fixture-csrf');
  expect(transport).toHaveBeenCalledWith('/auth/logout', expect.objectContaining({method:'POST',credentials:'same-origin',headers:expect.objectContaining({'X-CSRF-Token':'frontend-fixture-csrf'})}));
  await expect(client.session()).rejects.toMatchObject({status:403,code:'forbidden'});
 });
 it('rejects malformed responses and exposes typed errors rather than invented data', async () => {
  const client = new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify({user:'bad',csrf_token:''}))));
  await expect(client.session()).rejects.toBeInstanceOf(ApiError);
 });
});
