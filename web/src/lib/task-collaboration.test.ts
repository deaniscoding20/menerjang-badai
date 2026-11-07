import { expect,it,vi } from 'vitest';
import { ApiClient,taskCommentInputSchema } from './api';

it.each([
 ['4000 emoji','😀'.repeat(4000),true],
 ['4001 emoji','😀'.repeat(4001),false],
 ['mixed 4000 codepoints','a😀'.repeat(2000),true],
 ['mixed 4001 codepoints','a😀'.repeat(2000)+'b',false],
 ['trimmed boundary','  '+ '😀'.repeat(4000)+'\n',true],
 ['empty after trim',' \n ',false],
])('validates comment Unicode codepoints: %s',(_name,body,valid)=>{
 const result=taskCommentInputSchema.safeParse({body,kind:'comment'});
 expect(result.success).toBe(valid);
 if(result.success) expect(result.data.body).toBe(String(body).trim());
});

const task={id:'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8',brand_id:'d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860',title:'Brief',objective:'',instructions:'',priority:'normal',status:'draft',agent_id:'',version:3,deadline_at:'2027-01-02T09:30:00+07:00',created_at:'2026-11-06T00:00:00Z',updated_at:'2026-11-06T00:00:00Z'};
const comment={id:'comment-fixture',task_id:task.id,author_id:'human-fixture',kind:'clarification',body:'Who is the audience?',created_at:'2026-11-06T00:00:00Z'};
it('preserves a server deadline and can explicitly clear it using the current revision',async()=>{
 const transport=vi.fn().mockImplementation(()=>Promise.resolve(new Response(JSON.stringify(task))));
 const api=new ApiClient(transport);
 const loaded=await api.task(task.id);
 expect(loaded).toHaveProperty('deadline_at',task.deadline_at);
 await api.updateTask(task.id,{deadline_at:null,version:loaded.version},'frontend-fixture-csrf');
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/'+task.id,expect.objectContaining({body:JSON.stringify({deadline_at:null,version:3})}));
});
it('loads human notes and creates clarification without assigning an author or running the task',async()=>{
 const transport=vi.fn().mockResolvedValueOnce(new Response(JSON.stringify({items:[comment]}))).mockResolvedValueOnce(new Response(JSON.stringify(comment)));
 const api=new ApiClient(transport);
 expect((await api.taskComments(task.id)).items).toEqual([comment]);
 expect(await api.addTaskComment(task.id,{body:comment.body,kind:'clarification'},'frontend-fixture-csrf')).toEqual(comment);
 expect(transport).toHaveBeenLastCalledWith('/api/tasks/'+task.id+'/comments',expect.objectContaining({method:'POST',body:JSON.stringify({body:comment.body,kind:'clarification'})}));
});
it('rejects an invalid deadline instead of silently discarding it',async()=>{
 const api=new ApiClient(vi.fn().mockResolvedValue(new Response(JSON.stringify({...task,deadline_at:'tomorrow'}))));
 await expect(api.task(task.id)).rejects.toMatchObject({code:'invalid_response'});
});
