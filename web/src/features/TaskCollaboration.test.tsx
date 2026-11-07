import { fireEvent,render,screen,waitFor,within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { expect,it,vi } from 'vitest';
import { App } from '../app/App';
import { brandFixture,fixtureResponse,sessionFixture } from '../test/fixtures';

// UI transport fixtures only: no live authentication or database acceptance.
const taskFixture={id:'dbf244dd-5907-4308-bdb4-6fe2d3d58cd8',brand_id:'d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860',title:'Audience brief',objective:'Learn',instructions:'',priority:'normal',status:'draft',agent_id:'',version:3,deadline_at:'2027-01-02T02:30:45.123Z',created_at:'2026-11-06T00:00:00Z',updated_at:'2026-11-06T00:00:00Z'};
const existing={id:'comment-original',task_id:taskFixture.id,author_id:'human-owner',kind:'comment',body:'<script>untrusted note</script>',created_at:'2026-11-06T00:00:00Z'};
function transport(role='editor') {
 let task={...taskFixture};
 const comments=[existing];
 const fetcher=vi.fn().mockImplementation((path:string,options:RequestInit)=>{
  let data:unknown;
  if(path==='/api/session')data={...sessionFixture,user:{...sessionFixture.user,role}};
  else if(path==='/api/tasks/'+task.id+'/comments'){
   if(options.method==='POST'){
    const input=JSON.parse(String(options.body));
    const saved={...existing,...input,id:'comment-created',author_id:'server-author'};
    comments.push(saved);data=saved;
   }else data={items:[...comments]};
  } else if(path==='/api/tasks/'+task.id+'/events')data={items:[]};
  else if(path==='/api/tasks/'+task.id){
   if(options.method==='PATCH')task={...task,...JSON.parse(String(options.body)),version:task.version+1};
   data=task;
  } else if(path==='/api/brands')data={items:[{...brandFixture,id:task.brand_id}]};
  else data=fixtureResponse(path);
  return Promise.resolve(new Response(JSON.stringify(data),{status:options.method==='POST'?201:200}));
 });
 vi.stubGlobal('fetch',fetcher);
 return fetcher;
}
function show(){render(<MemoryRouter initialEntries={['/tasks/'+taskFixture.id]}><App/></MemoryRouter>);}
it('submits 4000 emoji without a restrictive UTF16 maxlength or truncation',async()=>{
 const fetcher=transport();show();
 await screen.findByText(existing.body);
 const body='😀'.repeat(4000);
 const textarea=screen.getByLabelText('Pesan');
 expect(textarea).not.toHaveAttribute('maxlength');
 fireEvent.change(textarea,{target:{value:body}});
 expect(textarea).toHaveValue(body);
 fireEvent.click(screen.getByRole('button',{name:'Kirim catatan'}));
 await waitFor(()=>expect(fetcher).toHaveBeenCalledWith('/api/tasks/'+taskFixture.id+'/comments',expect.objectContaining({method:'POST',body:JSON.stringify({body,kind:'comment'})})));
 expect(await screen.findByText(body)).toBeInTheDocument();
});

it('rejects a 4001-codepoint comment before sending it',async()=>{
 const fetcher=transport();show();
 await screen.findByText(existing.body);
 fireEvent.change(screen.getByLabelText('Pesan'),{target:{value:'😀'.repeat(4001)}});
 fireEvent.click(screen.getByRole('button',{name:'Kirim catatan'}));
 expect(await screen.findByText('Isi pesan antara 1 dan 4000 karakter.')).toBeInTheDocument();
 expect(fetcher.mock.calls.some(([,options])=>(options as RequestInit).method==='POST')).toBe(false);
});

it('renders human notes as text and submits clarification with CSRF without running or approving',async()=>{
 const fetcher=transport();show();
 expect(await screen.findByText(existing.body)).toBeInTheDocument();
 expect(document.querySelector('script')).toBeNull();
 const section=screen.getByRole('region',{name:'Komentar & klarifikasi'});
 await userEvent.type(within(section).getByLabelText('Pesan'),'Who is the audience?');
 await userEvent.selectOptions(within(section).getByLabelText('Jenis catatan'),'clarification');
 await userEvent.click(within(section).getByRole('button',{name:'Kirim catatan'}));
 expect(await screen.findByText('Who is the audience?')).toBeInTheDocument();
 expect(fetcher).toHaveBeenCalledWith('/api/tasks/'+taskFixture.id+'/comments',expect.objectContaining({method:'POST',headers:expect.objectContaining({'X-CSRF-Token':'frontend-fixture-csrf'}),body:JSON.stringify({body:'Who is the audience?',kind:'clarification'})}));
 expect(screen.getByRole('button',{name:/Jalankan task/})).toBeDisabled();
 expect(fetcher.mock.calls.some(([path])=>String(path).includes('/transition'))).toBe(false);
});
it('keeps comments visible but hides their write form for a viewer',async()=>{
 transport('viewer');show();
 expect(await screen.findByText(existing.body)).toBeInTheDocument();
 expect(screen.queryByRole('button',{name:'Kirim catatan'})).not.toBeInTheDocument();
});
it('retains an untouched deadline instant and supports explicitly clearing it with the loaded version',async()=>{
 const fetcher=transport();show();
 await userEvent.click(await screen.findByRole('button',{name:'Edit brief'}));
 expect(screen.getByLabelText('Tenggat (waktu lokal)')).not.toHaveValue('');
 await userEvent.click(screen.getByRole('button',{name:'Simpan perubahan'}));
 await waitFor(()=>expect(fetcher).toHaveBeenCalledWith('/api/tasks/'+taskFixture.id,expect.objectContaining({method:'PATCH',body:expect.stringContaining('"deadline_at":"'+taskFixture.deadline_at+'"')})));
 await userEvent.click(await screen.findByRole('button',{name:'Edit brief'}));
 await userEvent.clear(screen.getByLabelText('Tenggat (waktu lokal)'));
 await userEvent.click(screen.getByRole('button',{name:'Simpan perubahan'}));
 await waitFor(()=>expect(fetcher).toHaveBeenCalledWith('/api/tasks/'+taskFixture.id,expect.objectContaining({method:'PATCH',body:expect.stringContaining('"deadline_at":null')})));
 const mutations=fetcher.mock.calls.filter(([,options])=>(options as RequestInit).method==='PATCH');
 expect(mutations).toHaveLength(2);
 expect(mutations[1]?.[1]).toEqual(expect.objectContaining({body:expect.stringContaining('"version":4')}));
});
