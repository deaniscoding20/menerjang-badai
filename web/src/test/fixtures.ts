// FRONTEND FIXTURES ONLY. Never imported by production code; not an authentication bypass.
export const sessionFixture = {user:{id:'frontend-user',display_name:'Nadia Putri',role:'owner'},csrf_token:'frontend-fixture-csrf'};
export const brandFixture = {id:'brand-1',name:'Laut Biru',description:'Produk perawatan berkelanjutan.',website:'https://example.test',status:'active'};
export const overviewFixture = {brands_count:1,agents_count:6,integrations:[{id:'hermes',name:'Hermes native A2A',status:'unavailable',reason:'Gate keamanan native A2A belum terselesaikan.'}],release:{sha:'frontend-fixture',scope:'foundation'}};
export const agentsFixture = [
 ['claudia','Claudia','Assistant Brand Manager & Coordinator','@MB_claudia_bot','marketingdirector'],
 ['jordan','Jordan','Marketing Strategist','@MB_jordan_bot','marketingstrategist'],
 ['kobe','Kobe','Creative Lead','@MB_kobe_bot','creativelead'],
 ['curry','Curry','Market Researcher','@MB_curry_bot','marketresearcher'],
 ['sabrina','Sabrina','Social Media & Copywriter','@MB_sabrina_bot','socialmediacopywriter'],
 ['kyrie','Kyrie','Creative Designer','@MB_kyrie_bot','creativedesigner'],
].map(([id,name,role,telegram,profile])=>({id,name,role,telegram,profile,status:'unavailable',enabled:false,reason:'Gate keamanan native A2A belum terselesaikan.'}));
export function fixtureResponse(path:string) {
 if(path==='/api/session')return sessionFixture;
 if(path==='/api/overview')return overviewFixture;
 if(path==='/api/brands')return {items:[brandFixture]};
 if(path==='/api/brands/brand-1')return brandFixture;
 if(path==='/api/agents')return {items:agentsFixture};
 throw new Error('Unexpected frontend fixture request: '+path);
}
