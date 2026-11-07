import { useCallback, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { Link,useNavigate,useParams } from 'react-router-dom';
import { ArrowLeft,ShieldCheck } from 'lucide-react';
import { api } from '../lib/api';
import type { BrandInput } from '../lib/api';
import { useWorkspace } from '../app/WorkspaceContext';
import { useQuery } from '../lib/useQuery';
import { PageHeading,ErrorState,Loading } from '../components/Feedback';
export function BrandCreate() {
 const navigate=useNavigate();
 const {session,onError}=useWorkspace();
 return <BrandForm title="Mulai cerita brand Anda." onSave={async input=>{try{const brand=await api.createBrand(input,session.csrf_token);navigate('/brands/'+encodeURIComponent(brand.id));}catch(reason){onError(reason);throw reason;}}}/>;
}
export function BrandEdit() {
 const {id=''}=useParams();
 const navigate=useNavigate();
 const {session,onError}=useWorkspace();
 const load=useCallback(()=>api.brand(id),[id]);
 const {data,error,loading,retry}=useQuery(load);
 if(loading)return <Loading label="Memuat informasi brand…"/>;
 if(error || !data)return <ErrorState error={error} retry={retry}/>;
 if(data.status==='archived')return <ErrorState error={new Error('Brand yang diarsipkan tidak dapat diedit.')}/>;
 return <BrandForm key={data.id} title="Perbarui arah brand." initial={data} onSave={async input=>{try{await api.updateBrand(id,input,session.csrf_token);navigate('/brands/'+encodeURIComponent(id));}catch(reason){onError(reason);throw reason;}}}/>;
}
export function BrandForm({title,initial={name:'',description:'',website:''},onSave}:{title:string;initial?:BrandInput;onSave:(input:BrandInput)=>Promise<void>}) {
 const [values,setValues]=useState(initial);
 const [errors,setErrors]=useState<Partial<Record<keyof BrandInput,string>>>({});
 const [pending,setPending]=useState(false);
 const [error,setError]=useState<unknown>(null);
 const form=useRef<HTMLFormElement>(null);
 async function submit(event:FormEvent){
  event.preventDefault(); if(pending)return;
  const next:typeof errors={};
  if(!values.name.trim())next.name='Nama brand wajib diisi.';
  else if(values.name.trim().length>120)next.name='Nama brand maksimal 120 karakter.';
  if(values.description.length>4000)next.description='Deskripsi maksimal 4000 karakter.';
  if(values.website.trim()){try{const url=new URL(values.website.trim());if(!['http:','https:'].includes(url.protocol)||url.username||url.password)throw new Error('invalid');}catch{next.website='Gunakan URL lengkap http:// atau https://.';}}
  setErrors(next);
  if(Object.keys(next).length){const first=Object.keys(next)[0];form.current?.querySelector<HTMLInputElement>(`[name="${first}"]`)?.focus();return;}
  setPending(true);setError(null);
  try{await onSave({name:values.name.trim(),description:values.description.trim(),website:values.website.trim()});}catch(reason){setError(reason);}finally{setPending(false);}
 }
 return <><Link to="/brands" className="back-link"><ArrowLeft size={16} aria-hidden="true"/>Semua brand</Link><PageHeading eyebrow="IDENTITAS / BRAND" title={title} description="Mulai dari informasi dasar. Detail dapat diperbarui saat brand berkembang."/><div className="detail-grid"><section className="panel form-panel"><div className="section-heading"><h2>Informasi dasar</h2><span className="mini-label">* WAJIB DIISI</span></div><form ref={form} noValidate onSubmit={event=>void submit(event)} aria-busy={pending}><fieldset disabled={pending}><label htmlFor="brand-name">Nama brand <span aria-hidden="true">*</span></label><input id="brand-name" name="name" required autoComplete="organization" maxLength={120} value={values.name} aria-invalid={!!errors.name} aria-describedby={errors.name?'name-error':'name-help'} onChange={event=>setValues({...values,name:event.target.value})}/>{errors.name?<p className="field-error" id="name-error">{errors.name}</p>:<p id="name-help" className="field-help">Nama yang digunakan untuk mengenali brand di ruang kerja.</p>}<label htmlFor="brand-description">Deskripsi bisnis</label><textarea id="brand-description" name="description" rows={5} maxLength={4000} value={values.description} aria-invalid={!!errors.description} aria-describedby="description-help" onChange={event=>setValues({...values,description:event.target.value})}/><p className="field-help" id="description-help">{errors.description || 'Apa yang ditawarkan brand, dan apa yang membuatnya berbeda? Maks. 4000 karakter.'}</p><label htmlFor="brand-website">Website</label><input id="brand-website" name="website" type="url" inputMode="url" autoComplete="url" placeholder="https://" value={values.website} aria-invalid={!!errors.website} aria-describedby={errors.website?'website-error':'website-help'} onChange={event=>setValues({...values,website:event.target.value})}/>{errors.website?<p className="field-error" id="website-error">{errors.website}</p>:<p className="field-help" id="website-help">Opsional. Gunakan alamat lengkap http:// atau https://.</p>}</fieldset>{error!==null&&<ErrorState error={error}/>}<div className="form-actions"><Link className="button secondary" to="/brands" aria-disabled={pending} onClick={event=>{if(pending)event.preventDefault();}}>Batal</Link><button className="button primary" type="submit" disabled={pending}>{pending?'Menyimpan…':'Simpan brand'}</button></div></form></section><aside className="panel context-panel"><ShieldCheck size={28} aria-hidden="true"/><h2>Konteks yang tepat dimulai di sini.</h2><p>Informasi ini disimpan melalui API organisasi. Server memeriksa izin dan perlindungan CSRF pada setiap perubahan.</p><div className="quiet-note">Menyimpan brand tidak menjalankan agen, membuat kampanye, atau menerbitkan konten.</div></aside></div></>;
}
