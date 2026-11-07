import { ApiError } from '../lib/api';
import { Link } from 'react-router-dom';
import { AlertCircle, RotateCw } from 'lucide-react';
export function Loading({label='Memuat ruang kerja…'}:{label?:string}) { return <div className="state loading" role="status"><span className="loader" aria-hidden="true"/><p>{label}</p></div>; }
export function ErrorState({error,retry}:{error:unknown;retry?:()=>void}) {
 const denied=error instanceof ApiError && error.status===403;
 const missing=error instanceof ApiError && error.status===404;
 return <section className="state error" role="alert"><AlertCircle size={28} aria-hidden="true"/><h2>{denied?'Akses tidak diizinkan':missing?'Data tidak ditemukan':'Belum dapat memuat data'}</h2><p>{denied?'Anda tidak memiliki izin untuk melihat atau mengubah data ini. Hubungi administrator untuk meminta akses.':missing?'Data mungkin sudah diarsipkan atau tidak tersedia bagi akun Anda.':error instanceof Error?error.message:'Terjadi kesalahan. Coba lagi.'}</p>{!denied && !missing && retry?<button className="button secondary" onClick={retry}><RotateCw size={16} aria-hidden="true"/>Coba lagi</button>:<Link className="button secondary" to="/">Kembali ke ringkasan</Link>}</section>;
}
export function StatusBadge({children,tone='muted'}:{children:React.ReactNode;tone?:'muted'|'teal'|'coral'}) {return <span className={`badge ${tone}`}><span aria-hidden="true"/>{children}</span>;}
export function PageHeading({eyebrow,title,description,action}:{eyebrow:string;title:string;description:string;action?:React.ReactNode}) {return <header className="page-heading"><div><p className="eyebrow">{eyebrow}</p><h1 tabIndex={-1}>{title}</h1><p>{description}</p></div>{action}</header>;}
