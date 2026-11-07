import { useEffect, useState } from 'react';
import { useWorkspace } from '../app/WorkspaceContext';
export function useQuery<T>(load:()=>Promise<T>) {
 const {onError}=useWorkspace();
 const [data,setData]=useState<T>();
 const [error,setError]=useState<unknown>(null);
 const [loading,setLoading]=useState(true);
 const [attempt,setAttempt]=useState(0);
 useEffect(()=>{
  let active=true;
  load().then(value=>{if(active){setData(value);setError(null);}}).catch((reason:unknown)=>{if(active){setError(reason);onError(reason);}}).finally(()=>{if(active)setLoading(false);});
  return ()=>{active=false;};
 },[load,attempt,onError]);
 function retry(){setLoading(true);setError(null);setAttempt(value=>value+1);}
 return {data,error,loading,retry};
}
