import { render,screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { expect,it,vi } from 'vitest';
import { App } from './App';
import { fixtureResponse } from '../test/fixtures';
// Transport-only frontend fixtures; no fake auth path exists in production.
it('offers only the backend OIDC login when unauthenticated', async () => {
 vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response(JSON.stringify({error:{code:'unauthorized',message:'Login required'}}),{status:401})));
 render(<MemoryRouter><App /></MemoryRouter>);
 expect(await screen.findByRole('link',{name:/Masuk dengan Authentik/})).toHaveAttribute('href','/auth/login');
 expect(screen.queryByRole('navigation')).not.toBeInTheDocument();
});
it('loads a secure application shell with an honest Foundation overview', async () => {
 vi.stubGlobal('fetch',vi.fn().mockImplementation((path:string)=>Promise.resolve(new Response(JSON.stringify(fixtureResponse(path))))));
 render(<MemoryRouter><App /></MemoryRouter>);
 expect(await screen.findByRole('heading',{name:'Ruang kendali brand.'})).toBeInTheDocument();
 expect(screen.getByRole('navigation',{name:'Navigasi utama'})).toBeInTheDocument();
 expect(screen.getByText('Foundation release')).toBeInTheDocument();
 expect(await screen.findByText('Gate keamanan native A2A belum terselesaikan.')).toBeInTheDocument();
 expect(screen.getByRole('link',{name:'Tasks'})).toHaveAttribute('href','/tasks');
 expect(screen.getByRole('button',{name:/Knowledge.*Belum tersedia/})).toHaveAttribute('aria-disabled','true');
});
