import { useEffect, useState, useCallback } from 'react';
import { ApiError, api } from '../lib/api';
import type { Session } from '../lib/api';
import { Shell } from './Shell';
import { WorkspaceContext } from './WorkspaceContext';
export function App() {
 const [session,setSession]=useState<Session|null>(null);
 const [pending,setPending]=useState(true);
 const [expired,setExpired]=useState(false);
 const handleError=useCallback((reason:unknown)=>{if(reason instanceof ApiError && reason.status===401){setSession(null);setExpired(true);}},[]);
 const [error,setError]=useState<string|null>(null);
 useEffect(()=>{ let active=true; api.session().then(value=>{if(active)setSession(value);}).catch((reason:unknown)=>{if(active && !(reason instanceof ApiError && reason.status===401))setError('Sesi tidak dapat diperiksa.');}).finally(()=>{if(active)setPending(false);}); return ()=>{active=false;}; },[]);
 if(pending) return <main role="status">Memeriksa sesi aman…</main>;
 if(error) return <main role="alert">{error}</main>;
 if(session) return <WorkspaceContext value={{session,onError:handleError}}><Shell onLogout={()=>{setSession(null);setExpired(false);}}/></WorkspaceContext>;
 return <main className="login"><div className="login-story"><div className="wordmark">Menerjang<span>Badai</span><i>®</i></div><p className="eyebrow">KANTOR PEMASARAN AI</p><h1>Brand yang kuat.<br/>Tim yang terarah.</h1><p>Ruang kerja untuk membangun fondasi brand dan mengenal tim agen Anda. Keputusan tetap di tangan manusia.</p><div className="login-foot">FOUNDATION RELEASE <span>01 / FONDASI</span></div></div><section className="login-form" aria-labelledby="login-title"><div className="stamp" aria-hidden="true">MB.</div><p className="eyebrow">AKSES RUANG KERJA</p><h2 id="login-title">{expired?'Sesi Anda telah berakhir.':'Selamat datang kembali.'}</h2><p>Masuk melalui identitas organisasi Anda. Akses brand mengikuti izin yang diberikan administrator.</p><a className="button primary" href="/auth/login">Masuk dengan Authentik <span aria-hidden="true">↗</span></a><small>Sesi aman dikelola server. Tidak ada token akses yang disimpan di browser.</small><p className="login-note">Agen & integrasi belum menerima pekerjaan sampai verifikasi keamanan selesai.</p></section></main>;
}
