import { render,screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { expect,it,vi } from 'vitest';
import { App } from '../app/App';
import { fixtureResponse,sessionFixture } from '../test/fixtures';

// Transport fixtures only. These do not prove PostgreSQL, ACL or live agents.
function transport(role:string){
 vi.stubGlobal('fetch',vi.fn().mockImplementation((path:string)=>{
  const data=path==='/api/session'?{...sessionFixture,user:{...sessionFixture.user,role}}:path==='/api/tasks'?{items:[]}:fixtureResponse(path);
  return Promise.resolve(new Response(JSON.stringify(data)));
 }));
}
it.each(['owner','administrator','marketing_manager','editor'])('allows the documented %s role to prepare a draft without activating execution',async role=>{
 transport(role);
 render(<MemoryRouter initialEntries={['/tasks']}><App/></MemoryRouter>);
 expect(await screen.findByRole('heading',{name:'Pekerjaan yang punya arah.'})).toBeInTheDocument();
 expect(screen.getByRole('link',{name:'Buat task'})).toHaveAttribute('href','/tasks/new');
 expect(screen.getByRole('button',{name:/Jalankan task/})).toBeDisabled();
});
it.each(['viewer','analyst','approver'])('keeps the %s role read-only in the task UI',async role=>{
 transport(role);
 render(<MemoryRouter initialEntries={['/tasks']}><App/></MemoryRouter>);
 expect(await screen.findByRole('heading',{name:'Pekerjaan yang punya arah.'})).toBeInTheDocument();
 expect(screen.queryByRole('link',{name:'Buat task'})).not.toBeInTheDocument();
});
