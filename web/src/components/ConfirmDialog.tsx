import { useEffect,useRef } from 'react';
export function ConfirmDialog({title,children,pending,onCancel,onConfirm}:{title:string;children:React.ReactNode;pending:boolean;onCancel:()=>void;onConfirm:()=>void}) {
 const dialog=useRef<HTMLDialogElement>(null);
 useEffect(()=>{const element=dialog.current;element?.showModal();return ()=>element?.close();},[]);
 return <dialog ref={dialog} className="confirm-dialog" aria-labelledby="confirm-title" onCancel={event=>{event.preventDefault();if(!pending)onCancel();}}><p className="eyebrow">KONFIRMASI PERUBAHAN</p><h2 id="confirm-title">{title}</h2>{children}<div className="form-actions"><button autoFocus className="button secondary" onClick={onCancel} disabled={pending}>Batal</button><button className="button danger" disabled={pending} onClick={onConfirm}>{pending?'Mengarsipkan…':'Ya, arsipkan brand'}</button></div></dialog>;
}
