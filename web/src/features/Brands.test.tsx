import { render,screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { expect,it,vi } from 'vitest';
import { App } from '../app/App';
import { brandFixture, fixtureResponse } from '../test/fixtures';
// Explicit frontend API fixtures. Backend ACL is separately verified by live tests.
it('lists authorized brands and opens their server-backed detail', async () => {
 vi.stubGlobal('fetch',vi.fn().mockImplementation((path:string)=>Promise.resolve(new Response(JSON.stringify(fixtureResponse(path))))));
 render(<MemoryRouter initialEntries={['/brands']}><App/></MemoryRouter>);
 await userEvent.click(await screen.findByRole('link',{name:/Laut Biru/}));
 expect(await screen.findByRole('heading',{name:'Laut Biru'})).toBeInTheDocument();
 expect(screen.getByText('Produk perawatan berkelanjutan.')).toBeInTheDocument();
 expect(screen.getByRole('link',{name:'https://example.test'})).toHaveAttribute('rel','noopener noreferrer');
});
it('validates brand input then creates through CSRF-protected API', async () => {
 const transport=vi.fn().mockImplementation((path:string,options:RequestInit)=>Promise.resolve(new Response(JSON.stringify(options.method==='POST'?brandFixture:fixtureResponse(path)),{status:options.method==='POST'?201:200})));
 vi.stubGlobal('fetch',transport);
 render(<MemoryRouter initialEntries={['/brands/new']}><App/></MemoryRouter>);
 await userEvent.click(await screen.findByRole('button',{name:'Simpan brand'}));
 expect(await screen.findByText('Nama brand wajib diisi.')).toBeInTheDocument();
 await userEvent.type(screen.getByLabelText(/Nama brand/),'Laut Biru');
 await userEvent.type(screen.getByLabelText(/Website/),'javascript:alert(1)');
 await userEvent.click(screen.getByRole('button',{name:'Simpan brand'}));
 expect(await screen.findByText('Gunakan URL lengkap http:// atau https://.')).toBeInTheDocument();
 await userEvent.clear(screen.getByLabelText(/Website/));
 await userEvent.type(screen.getByLabelText(/Website/),'https://example.test');
 await userEvent.click(screen.getByRole('button',{name:'Simpan brand'}));
 expect(await screen.findByRole('heading',{name:'Laut Biru'})).toBeInTheDocument();
 expect(transport).toHaveBeenCalledWith('/api/brands',expect.objectContaining({method:'POST',headers:expect.objectContaining({'X-CSRF-Token':'frontend-fixture-csrf'}),body:JSON.stringify({name:'Laut Biru',description:'',website:'https://example.test'})}));
});
it('edits a brand and requires explicit confirmation before archiving', async () => {
 let current={...brandFixture};
 const transport=vi.fn().mockImplementation((path:string,options:RequestInit)=>{
  if(options.method==='PATCH'){current={...current,...JSON.parse(String(options.body))};return Promise.resolve(new Response(JSON.stringify(current)));}
  if(options.method==='DELETE'){current={...current,status:'archived'};return Promise.resolve(new Response(null,{status:204}));}
  return Promise.resolve(new Response(JSON.stringify(path==='/api/brands/brand-1'?current:fixtureResponse(path))));
 });
 vi.stubGlobal('fetch',transport);
 render(<MemoryRouter initialEntries={['/brands/brand-1/edit']}><App/></MemoryRouter>);
 const name=await screen.findByLabelText(/Nama brand/);
 await userEvent.clear(name);await userEvent.type(name,'Laut Biru Baru');
 await userEvent.click(screen.getByRole('button',{name:'Simpan brand'}));
 expect(await screen.findByRole('heading',{name:'Laut Biru Baru'})).toBeInTheDocument();
 await userEvent.click(screen.getByRole('button',{name:'Arsipkan brand'}));
 expect(await screen.findByRole('dialog',{name:'Arsipkan brand ini?'})).toBeInTheDocument();
 expect(transport.mock.calls.filter(([,options])=>(options as RequestInit).method==='DELETE')).toHaveLength(0);
 await userEvent.click(screen.getByRole('button',{name:'Ya, arsipkan brand'}));
 expect(await screen.findByRole('heading',{name:'Setiap brand, punya arah.'})).toBeInTheDocument();
 expect(transport).toHaveBeenCalledWith('/api/brands/brand-1',expect.objectContaining({method:'DELETE',headers:expect.objectContaining({'X-CSRF-Token':'frontend-fixture-csrf'})}));
});
