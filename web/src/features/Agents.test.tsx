import { render,screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { expect,it,vi } from 'vitest';
import { App } from '../app/App';
import { fixtureResponse } from '../test/fixtures';
// Frontend directory fixtures; no Hermes execution is invoked by these tests.
it('renders all six server profiles without claiming reachability and shows profile details', async () => {
 vi.stubGlobal('fetch',vi.fn().mockImplementation((path:string)=>Promise.resolve(new Response(JSON.stringify(fixtureResponse(path))))));
 render(<MemoryRouter initialEntries={['/agents']}><App/></MemoryRouter>);
 for(const name of ['Claudia','Jordan','Kobe','Curry','Sabrina','Kyrie'])expect(await screen.findByRole('link',{name:new RegExp('Lihat profil '+name)})).toBeInTheDocument();
 expect(screen.queryByText('Online')).not.toBeInTheDocument();
 expect(screen.getByText('Kolaborasi berdasarkan peran, bukan rantai komando wajib.')).toBeInTheDocument();
 await userEvent.click(screen.getByRole('link',{name:'Lihat profil Claudia'}));
 expect(await screen.findByRole('heading',{name:'Claudia'})).toBeInTheDocument();
 expect(screen.getByText('@MB_claudia_bot')).toBeInTheDocument();
 expect(screen.getByText('marketingdirector')).toBeInTheDocument();
 expect(screen.getByRole('button',{name:'Eksekusi belum diaktifkan'})).toBeDisabled();
});
