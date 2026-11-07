import { useCallback, useState } from 'react';
import { Link,useParams,useNavigate } from 'react-router-dom';
import { ArrowUpRight, ArrowLeft, Search, Layers, ExternalLink } from 'lucide-react';
import { useWorkspace } from '../app/WorkspaceContext';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { api } from '../lib/api';
import { useQuery } from '../lib/useQuery';
import { Loading,ErrorState,PageHeading,StatusBadge } from '../components/Feedback';
const loadBrands=()=>api.brands();
function SafeWebsite({url}:{url:string}) {
 let safe=false;try{safe=['https:','http:'].includes(new URL(url).protocol);}catch{/* Invalid websites are plain text, never executable links. */}
 return safe?<a href={url} target="_blank" rel="noopener noreferrer">{url}<ExternalLink size={14} aria-hidden="true"/></a>:<span>{url || 'Belum ditambahkan'}</span>;
}
export function BrandList() {
 const {data,error,loading,retry}=useQuery(loadBrands);
 const [search,setSearch]=useState('');
 const [status,setStatus]=useState('all');
 if(loading)return <Loading label="Memuat brand yang diizinkan…"/>;
 if(error || !data)return <ErrorState error={error} retry={retry}/>;
 const items=data.items.filter(brand=>brand.name.toLocaleLowerCase('id-ID').includes(search.toLocaleLowerCase('id-ID')) && (status==='all'||brand.status===status));
 return <><PageHeading eyebrow="RUANG KERJA / BRAND" title="Setiap brand, punya arah." description="Identitas dan konteks brand yang dapat diakses oleh akun Anda." action={<Link className="button primary" to="/brands/new">Tambah brand <ArrowUpRight size={17} aria-hidden="true"/></Link>}/><section className="panel"><div className="list-toolbar"><div className="search-field"><Search size={17} aria-hidden="true"/><input aria-label="Cari brand" placeholder="Cari nama brand…" value={search} onChange={event=>setSearch(event.target.value)}/></div><label className="filter-label">Status <select value={status} onChange={event=>setStatus(event.target.value)}><option value="all">Semua status</option><option value="active">Aktif</option><option value="archived">Diarsipkan</option></select></label></div><div className="brand-list">{items.map(brand=><article key={brand.id} className="brand-row"><div className="brand-initial" aria-hidden="true">{brand.name.slice(0,2).toUpperCase()}</div><div className="brand-row-main"><Link className="brand-title" to={'/brands/'+encodeURIComponent(brand.id)}>{brand.name}<ArrowUpRight size={17} aria-hidden="true"/></Link><p>{brand.description || 'Deskripsi brand belum ditambahkan.'}</p></div><StatusBadge tone={brand.status==='active'?'teal':'muted'}>{brand.status==='active'?'Aktif':'Diarsipkan'}</StatusBadge></article>)}</div>{items.length===0&&<div className="state"><Layers size={30} aria-hidden="true"/><h2>{data.items.length===0?'Mulai dari brand pertama.':'Tidak ada brand yang cocok.'}</h2><p>{data.items.length===0?'Tambahkan identitas brand sebagai fondasi kerja tim. Daftar hanya memuat brand yang boleh Anda akses.':'Ubah kata pencarian atau filter status.'}</p>{data.items.length===0&&<Link to="/brands/new" className="button primary">Tambah brand</Link>}</div>}<div className="list-foot">{items.length} dari {data.items.length} brand yang diizinkan · Otorisasi diterapkan server.</div></section></>;
}
export function BrandDetail() {
 const {id=''}=useParams();
 const {session,onError}=useWorkspace();
 const navigate=useNavigate();
 const [confirm,setConfirm]=useState(false);
 const [pending,setPending]=useState(false);
 const [archiveError,setArchiveError]=useState<unknown>(null);
 async function archive(){if(pending)return;setPending(true);setArchiveError(null);try{await api.archiveBrand(id,session.csrf_token);navigate('/brands');}catch(reason){onError(reason);setArchiveError(reason);}finally{setPending(false);}}

 const load=useCallback(()=>api.brand(id),[id]);
 const {data,error,loading,retry}=useQuery(load);
 if(loading)return <Loading label="Memuat profil brand…"/>;
 if(error || !data)return <ErrorState error={error} retry={retry}/>;
 return <><Link className="back-link" to="/brands"><ArrowLeft size={16} aria-hidden="true"/>Semua brand</Link><PageHeading eyebrow="PROFIL BRAND" title={data.name} description="Satu sumber identitas untuk arah pemasaran yang konsisten." action={data.status==='active'?<Link className="button primary" to={`/brands/${encodeURIComponent(id)}/edit`}>Edit brand <ArrowUpRight size={17} aria-hidden="true"/></Link>:<StatusBadge>Diarsipkan</StatusBadge>}/><div className="detail-grid"><section className="panel profile-content"><div className="section-heading"><h2>Identitas brand</h2><StatusBadge tone={data.status==='active'?'teal':'muted'}>{data.status==='active'?'Aktif':'Diarsipkan'}</StatusBadge></div><dl><dt>Nama brand</dt><dd>{data.name}</dd><dt>Deskripsi bisnis</dt><dd className="preserve-lines">{data.description || 'Belum ditambahkan'}</dd><dt>Website</dt><dd><SafeWebsite url={data.website}/></dd></dl></section><aside className="panel context-panel"><p className="eyebrow">FONDASI BRAND</p><h2>Ruang untuk bertumbuh.</h2><p>Profil ini belum memuat produk, audiens, positioning, tone of voice, logo, atau pengelolaan anggota. Field dan izin tersebut belum tersedia pada API rilis ini.</p><div className="quiet-note">Hak akses brand ditentukan server, bukan halaman ini.</div>{data.status==='active'&&<button className="button danger-outline archive-button" onClick={()=>setConfirm(true)}>Arsipkan brand</button>}</aside></div>{confirm&&<ConfirmDialog title="Arsipkan brand ini?" pending={pending} onCancel={()=>{setConfirm(false);setArchiveError(null);}} onConfirm={()=>void archive()}><p>Brand <strong>{data.name}</strong> akan diarsipkan, bukan dihapus permanen. Perubahan baru tidak dapat dilakukan pada brand yang diarsipkan.</p>{archiveError!==null&&<ErrorState error={archiveError}/>}</ConfirmDialog>}</>;
}
