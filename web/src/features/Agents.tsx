import { Link, useParams } from 'react-router-dom';
import { ArrowLeft,ArrowUpRight,LockKeyhole,ShieldCheck } from 'lucide-react';
import { api,ApiError } from '../lib/api';
import { useQuery } from '../lib/useQuery';
import { Loading,ErrorState,PageHeading,StatusBadge } from '../components/Feedback';
const load=()=>api.agents();
export function AgentDirectory() {
 const {data,error,loading,retry}=useQuery(load);
 if(loading)return <Loading label="Memuat direktori agen…"/>;
 if(error || !data)return <ErrorState error={error} retry={retry}/>;
 return <><PageHeading eyebrow="KANTOR PEMASARAN AI / TIM AGEN" title="Enam peran. Satu arah." description="Direktori profil Hermes yang sudah ada. Peran melengkapi, bukan menggantikan keputusan Anda."/><section className="team-intro"><div><p className="eyebrow">CARA TIM INI BEKERJA</p><h2>Kolaborasi berdasarkan peran, bukan rantai komando wajib.</h2><p>Koordinasi, strategi, riset, copy, dan kreativitas berada dalam satu ruang. Setiap pekerjaan nantinya mengikuti kebutuhan brand dan izin yang diberikan — tidak harus selalu melewati Claudia.</p></div><div className="team-security"><LockKeyhole size={22} aria-hidden="true"/><strong>Eksekusi nonaktif</strong><p>Gate keamanan native A2A belum selesai diverifikasi. Telegram bukan bukti agen online.</p></div></section><div className="section-heading directory-heading"><h2>Kenali tim Anda</h2><span className="mini-label">{data.items.length} PROFIL DARI SERVER</span></div><div className="agent-grid">{data.items.map((agent,index)=><article className={`agent-card agent-${index%3}`} key={agent.id}><div className="agent-card-top"><span className="agent-monogram" aria-hidden="true">{agent.name.slice(0,1)}</span><span className="agent-number">{String(index+1).padStart(2,'0')}</span></div><h3>{agent.name}</h3><p className="agent-role">{agent.role}</p><p className="telegram">{agent.telegram}</p><div className="agent-card-bottom"><StatusBadge>Eksekusi nonaktif</StatusBadge><Link className="icon-button" to={'/agents/'+encodeURIComponent(agent.id)} aria-label={'Lihat profil '+agent.name}><ArrowUpRight size={19} aria-hidden="true"/></Link></div></article>)}</div>{data.items.length===0&&<div className="panel state"><h2>Belum ada profil yang tersedia.</h2><p>Server belum mengembalikan direktori agen untuk akun Anda.</p></div>}<p className="directory-note">Profil, prompt, dan konfigurasi native tidak diubah dari dashboard ini.</p></>;
}
export function AgentDetail() {
 const {id}=useParams();
 const {data,error,loading,retry}=useQuery(load);
 if(loading)return <Loading label="Memuat profil agen…"/>;
 if(error || !data)return <ErrorState error={error} retry={retry}/>;
 const agent=data.items.find(item=>item.id===id);
 if(!agent)return <ErrorState error={new ApiError(404,'not_found','Profil tidak ditemukan.')}/>;
 return <><Link className="back-link" to="/agents"><ArrowLeft size={16} aria-hidden="true"/>Semua agen</Link><PageHeading eyebrow="DIREKTORI / PROFIL AGEN" title={agent.name} description={agent.role} action={<StatusBadge>Eksekusi nonaktif</StatusBadge>}/><div className="detail-grid"><section className="panel profile-content"><div className="section-heading"><h2>Identitas native</h2><span className="mini-label">DARI REGISTRY SERVER</span></div><dl><dt>Nama</dt><dd>{agent.name}</dd><dt>Peran</dt><dd>{agent.role}</dd><dt>Telegram</dt><dd>{agent.telegram}</dd><dt>Profil Hermes</dt><dd><code>{agent.profile}</code></dd><dt>ID aplikasi</dt><dd><code>{agent.id}</code></dd><dt>Status registry</dt><dd><code>{agent.status}</code><small className="block-note">Status registry bukan konfirmasi agen online atau siap menerima pekerjaan.</small></dd></dl></section><aside className="panel context-panel"><ShieldCheck size={28} aria-hidden="true"/><h2>Kemampuan belum diaktifkan.</h2><p>{agent.reason || 'Server belum mengonfirmasi gate keamanan native A2A.'}</p><button className="button secondary" disabled><LockKeyhole size={16} aria-hidden="true"/>Eksekusi belum diaktifkan</button><div className="quiet-note">Tidak ada heartbeat atau aktivitas terakhir yang terverifikasi pada rilis ini. Prompt, persona, dan konfigurasi asli tetap utuh.</div></aside></div></>;
}
