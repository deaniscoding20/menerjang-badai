import { expect,it,vi } from 'vitest';
import { ApiClient } from './api';

const task={id:'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8',brand_id:'d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860',title:'Brief',objective:'',instructions:'',priority:'normal',status:'draft',agent_id:'',version:3,created_at:'2026-11-06T00:00:00Z',updated_at:'2026-11-06T00:00:00Z'};
it('preserves the task revision and sends it with transitions',async()=>{
 const transport=vi.fn().mockImplementation(()=>Promise.resolve(new Response(JSON.stringify(task))));
 const api=new ApiClient(transport);
 const loaded=await api.task(task.id);
 expect(loaded).toHaveProperty('version',3);
 // Version is from the read task, not an invented latest revision.
 await api.transitionTask(task.id,'ready','frontend-fixture-csrf',3);
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/'+task.id+'/transition',expect.objectContaining({body:JSON.stringify({status:'ready',version:3})}));
});
it('does not silently accept a task response missing optimistic revision',async()=>{
 const {version,...missing}=task;
 expect(version).toBe(3);
 const api=new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify(missing))));
 await expect(api.task(task.id)).rejects.toMatchObject({code:'invalid_response'});
});
