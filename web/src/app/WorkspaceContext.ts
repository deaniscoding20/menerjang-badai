import { createContext, useContext } from 'react';
import type { Session } from '../lib/api';
export const WorkspaceContext = createContext<{session:Session;onError:(reason:unknown)=>void}|null>(null);
export function useWorkspace() { const value=useContext(WorkspaceContext); if(!value)throw new Error('Workspace context is required'); return value; }
