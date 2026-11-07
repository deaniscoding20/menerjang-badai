import { z } from 'zod';
const sessionSchema = z.object({user:z.object({id:z.string().min(1),display_name:z.string(),role:z.string()}),csrf_token:z.string().min(1)});
export type Session = z.infer<typeof sessionSchema>;
const brandSchema=z.object({id:z.string().min(1),name:z.string(),description:z.string(),website:z.string(),status:z.enum(['active','archived'])});
export type Brand=z.infer<typeof brandSchema>;
export type BrandInput=Pick<Brand,'name'|'description'|'website'>;
const agentSchema=z.object({id:z.string().min(1),name:z.string(),role:z.string(),telegram:z.string(),profile:z.string(),status:z.string(),enabled:z.literal(false),reason:z.string()});
export type Agent=z.infer<typeof agentSchema>;
export const taskPrioritySchema = z.enum(['low','normal','high','urgent']);
export const taskStatusSchema = z.enum(['draft','ready','assigned','queued','running','waiting_input','waiting_approval','completed','failed','cancelled']);
const deadlineSchema = z.string().datetime({offset:true});
export const taskSchema = z.object({
 id:z.string().min(1),brand_id:z.string().min(1),title:z.string(),objective:z.string(),instructions:z.string(),
 priority:taskPrioritySchema,status:taskStatusSchema,agent_id:z.string().nullable().optional(),
 version:z.number().int().positive(),deadline_at:deadlineSchema.nullable().optional(),created_at:z.string(),updated_at:z.string(),
});
export const taskEventSchema = z.object({
 id:z.string().min(1),from_status:taskStatusSchema.or(z.literal('')).nullable().optional(),
 to_status:taskStatusSchema,actor_id:z.string(),created_at:z.string(),
});
export const taskInputSchema = z.object({
 brand_id:z.string().uuid(),title:z.string().trim().min(1),objective:z.string(),instructions:z.string(),
 priority:taskPrioritySchema,agent_id:z.string().min(1).nullable().optional(),deadline_at:deadlineSchema.nullable().optional(),
});
export const taskUpdateSchema = taskInputSchema.omit({brand_id:true}).partial().extend({version:z.number().int().positive()});
export type Task = z.infer<typeof taskSchema>;
export type TaskStatus = z.infer<typeof taskStatusSchema>;
export type TaskPriority = z.infer<typeof taskPrioritySchema>;
export type TaskEvent = z.infer<typeof taskEventSchema>;
export type TaskInput = z.infer<typeof taskInputSchema>;
export type TaskUpdate = z.infer<typeof taskUpdateSchema>;
export const taskCommentInputSchema=z.object({body:z.string().trim().min(1).refine(body=>Array.from(body).length<=4000,{message:'Isi pesan maksimal 4000 karakter.'}),kind:z.enum(['comment','clarification'])});
const taskCommentSchema=z.object({id:z.string().min(1),task_id:z.string().min(1),author_id:z.string().min(1),kind:z.enum(['comment','clarification']),body:z.string(),created_at:z.string()});
export type TaskCommentInput=z.infer<typeof taskCommentInputSchema>;
const taskDependenciesSchema=z.object({version:z.number().int().positive(),items:z.array(z.object({task_id:z.string().uuid(),title:z.string(),status:taskStatusSchema}))});
export type TaskDependencyInput={dependency_id:string;version:number};
const deliverableKindSchema=z.enum(['research_report','marketing_strategy','creative_brief','caption_package','design_specification','campaign_plan','performance_report']);
// strings.TrimSpace uses Unicode White_Space: NEL is blank, but U+FEFF is not.
const deliverableText=(max:number)=>z.string().refine(value=>/[^\p{White_Space}]/u.test(value)&&!value.includes('\0')&&Array.from(value).length<=max);
const deliverableSourceURL=z.string().refine(value=>{
 try {
  const url=new URL(value);
  // Go checks the original UTF-8 bytes and RawQuery, not WHATWG-normalized search.
  const reference=value.split('#',1)[0]??'';
  const queryIndex=reference.indexOf('?');
  return new TextEncoder().encode(value).length<=2048&&value.startsWith('https://')&&url.protocol==='https:'&&!!url.hostname&&!url.username&&!url.password&&!value.slice(8).split(/[/?#]/)[0]?.includes('@')&&(queryIndex<0||queryIndex===reference.length-1)&&!Array.from(value).some(char=>char.charCodeAt(0)<0x20||char.charCodeAt(0)===0x7f);
 } catch { return false; }
});
export const structuredDeliverableSchema=z.strictObject({schema_version:z.literal(1),kind:deliverableKindSchema,title:deliverableText(200),sections:z.array(z.strictObject({heading:deliverableText(160),body:deliverableText(12000)})).min(1).max(40),sources:z.array(z.strictObject({label:deliverableText(200),url:deliverableSourceURL})).max(40).optional()});
const taskDeliverableSummarySchema=z.object({id:z.string().uuid(),task_id:z.string().uuid(),revision:z.number().int().positive(),author_id:z.string().uuid(),origin:z.literal('human_import'),validation_status:z.enum(['validated','invalid']),title:z.string(),kind:z.string(),created_at:z.string().datetime({offset:true})}).refine(value=>value.validation_status==='invalid'?value.kind==='':deliverableKindSchema.safeParse(value.kind).success);
const taskDeliverablesSchema=z.object({version:z.number().int().positive(),items:z.array(taskDeliverableSummarySchema).max(100)});
const taskDeliverableSchema=taskDeliverableSummarySchema.safeExtend({raw_output:z.string(),structured:structuredDeliverableSchema.nullable()}).refine(value=>value.validation_status==='invalid'?value.structured===null:value.structured!==null&&value.kind===value.structured.kind&&value.title===value.structured.title);
export type TaskDeliverable=z.infer<typeof taskDeliverableSchema>;
export type TaskDeliverableInput={raw_output:string;version:number};
const revisionRequestBody=deliverableText(4000);
export const taskRevisionRequestInputSchema=z.strictObject({deliverable_id:z.string().uuid(),body:z.string().transform(value=>value.replace(/^\p{White_Space}+|\p{White_Space}+$/gu,'')).pipe(revisionRequestBody),version:z.number().int().positive()});
const taskRevisionRequestSchema=z.strictObject({id:z.string().uuid(),task_id:z.string().uuid(),deliverable_id:z.string().uuid(),deliverable_revision:z.number().int().positive(),author_id:z.string().uuid(),body:revisionRequestBody,created_at:z.string().datetime({offset:true})});
const taskRevisionRequestsSchema=z.strictObject({version:z.number().int().positive(),items:z.array(taskRevisionRequestSchema).max(100)});
export type TaskRevisionRequestInput=z.infer<typeof taskRevisionRequestInputSchema>;
const overviewSchema = z.object({brands_count:z.number().int().nonnegative(),agents_count:z.number().int().nonnegative(),integrations:z.array(z.object({id:z.string(),name:z.string(),status:z.string(),reason:z.string()})),release:z.object({sha:z.string(),scope:z.string()})});
export type Overview = z.infer<typeof overviewSchema>;
const errorSchema = z.object({error:z.object({code:z.string(),message:z.string()})});
export class ApiError extends Error {
 readonly status:number;
 readonly code:string;
 constructor(status:number,code:string,message:string){super(message);this.name='ApiError';this.status=status;this.code=code;}
}
export class ApiClient {
 private readonly transport:typeof fetch;
 constructor(transport:typeof fetch=(...args)=>fetch(...args)){this.transport=transport;}
 private async request<T>(path:string, schema:z.ZodType<T>, method = 'GET', token?:string, body?:unknown):Promise<T> {
  const headers:Record<string,string> = {Accept:'application/json'};
  if (method !== 'GET') {
   if (!token) throw new ApiError(403,'csrf_missing','Token keamanan tidak tersedia. Masuk kembali.');
   headers['X-CSRF-Token']=token;
  }
  if (body !== undefined) headers['Content-Type']='application/json';
  let response:Response;
  try { response = await this.transport(path,{method,credentials:'same-origin',redirect:'error',cache:'no-store',headers,body:body === undefined ? undefined : JSON.stringify(body)}); }
  catch { throw new ApiError(0,'network_error','Server tidak dapat dijangkau. Periksa koneksi lalu coba lagi.'); }
  let data:unknown;
  try { data = response.status === 204 ? undefined : await response.json(); }
  catch { throw new ApiError(response.status,'invalid_response','Respons server tidak dapat dibaca.'); }
  if (!response.ok) {
   const parsed = errorSchema.safeParse(data);
   throw new ApiError(response.status,parsed.success ? parsed.data.error.code : 'request_failed',parsed.success ? parsed.data.error.message : 'Permintaan tidak dapat diproses.');
  }
  const parsed = schema.safeParse(data);
  if (!parsed.success) throw new ApiError(502,'invalid_response','Respons server tidak sesuai kontrak. Tidak ada data pengganti yang ditampilkan.');
  return parsed.data;
 }
 session() { return this.request('/api/session',sessionSchema); }
 overview() { return this.request('/api/overview',overviewSchema); }
 agents() { return this.request('/api/agents',z.object({items:z.array(agentSchema)})); }
 brands() { return this.request('/api/brands',z.object({items:z.array(brandSchema)})); }
 brand(id:string) { return this.request('/api/brands/'+encodeURIComponent(id),brandSchema); }
 createBrand(input:BrandInput,token:string) { return this.request('/api/brands',brandSchema,'POST',token,input); }
 updateBrand(id:string,input:BrandInput,token:string) { return this.request('/api/brands/'+encodeURIComponent(id),brandSchema,'PATCH',token,input); }
 archiveBrand(id:string,token:string) { return this.request('/api/brands/'+encodeURIComponent(id),z.unknown(),'DELETE',token); }
 tasks(brandId?:string) { return this.request('/api/tasks'+(brandId?'?'+new URLSearchParams({brand_id:brandId}).toString():''),z.object({items:z.array(taskSchema)})); }
 task(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id),taskSchema); }
 createTask(input:TaskInput,token:string) { return this.request('/api/tasks',taskSchema,'POST',token,input); }
 updateTask(id:string,input:TaskUpdate,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id),taskSchema,'PATCH',token,input); }
 transitionTask(id:string,status:TaskStatus,token:string,version:number) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/transition',taskSchema,'POST',token,{status,version}); }
 taskEvents(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/events',z.object({items:z.array(taskEventSchema)})); }
 taskDependencies(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/dependencies',taskDependenciesSchema); }
 addTaskDependency(id:string,input:TaskDependencyInput,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/dependencies',taskSchema,'POST',token,input); }
 removeTaskDependency(id:string,input:TaskDependencyInput,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/dependencies',taskSchema,'DELETE',token,input); }
 taskDeliverables(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/deliverables',taskDeliverablesSchema); }
 taskDeliverable(id:string,deliverableId:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/deliverables/'+encodeURIComponent(deliverableId),taskDeliverableSchema); }
 importTaskDeliverable(id:string,input:TaskDeliverableInput,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/deliverables',taskDeliverableSchema,'POST',token,input); }
 taskRevisionRequests(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/revision-requests',taskRevisionRequestsSchema); }
 requestTaskRevision(id:string,input:TaskRevisionRequestInput,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/revision-requests',taskRevisionRequestSchema,'POST',token,input); }
 taskComments(id:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/comments',z.object({items:z.array(taskCommentSchema)})); }
 addTaskComment(id:string,input:TaskCommentInput,token:string) { return this.request('/api/tasks/'+encodeURIComponent(id)+'/comments',taskCommentSchema,'POST',token,input); }
 logout(token:string) { return this.request('/auth/logout',z.undefined(),'POST',token); }
}
export const api = new ApiClient();
