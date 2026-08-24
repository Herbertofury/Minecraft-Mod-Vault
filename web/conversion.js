'use strict';

let conversionStatus=null;
let conversionCurrent=null;
let conversionLoading=false;
let conversionInitialized=false;

async function loadConversionStudio(force=false){
  if(conversionLoading)return;
  conversionLoading=true;setConversionWorking(true);
  try{
    conversionStatus=await api('/api/conversion/status');
    renderConversionStatus();
    if(!conversionInitialized){wireConversionStudio();conversionInitialized=true}
    const wanted=conversionCurrent?.id||sessionStorage.getItem('mmv-conversion-session')||conversionStatus.sessions?.[0]?.id;
    if(wanted&&(!conversionCurrent||conversionCurrent.id!==wanted||force))await loadConversionSession(wanted,false);
    else renderConversionShell();
  }catch(err){showToast(`OmniBridge: ${err.message}`,true);document.getElementById('conversionSessionList').innerHTML=`<div class="empty-inline">${esc(err.message)}</div>`}
  finally{conversionLoading=false;setConversionWorking(false)}
}

function wireConversionStudio(){
  const picker=document.getElementById('conversionSourceFile');
  const choose=()=>picker.click();
  document.getElementById('conversionChooseSource').onclick=choose;
  document.getElementById('conversionEmptyChoose').onclick=e=>{e.stopPropagation();choose()};
  picker.onchange=()=>{const file=picker.files?.[0];if(file)importConversionFile(file);picker.value=''};
  document.getElementById('conversionRefresh').onclick=()=>loadConversionStudio(true);
  document.getElementById('conversionOpenRoot').onclick=()=>conversionStatus?.root&&openExternal(conversionStatus.root,true);
  document.getElementById('conversionImportPath').onclick=importConversionPath;
  document.getElementById('conversionBuildPlan').onclick=buildConversionPlanUI;
  document.getElementById('conversionRun').onclick=runConversionUI;
  document.getElementById('conversionOpenSession').onclick=()=>conversionCurrent?.paths?.root&&openExternal(conversionCurrent.paths.root,true);
  document.getElementById('conversionDeleteSession').onclick=deleteConversionSession;
  document.getElementById('conversionDownloadReport').onclick=()=>{if(conversionCurrent)location.href=conversionDownloadURL(conversionCurrent.id,'report')};
  document.getElementById('conversionTarget').onchange=syncConversionTargetDefaults;
  const drop=document.getElementById('conversionDropZone');
  drop.onclick=e=>{if(!e.target.closest('button'))choose()};
  drop.onkeydown=e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();choose()}};
  for(const type of ['dragenter','dragover'])drop.addEventListener(type,e=>{e.preventDefault();drop.classList.add('dragging')});
  for(const type of ['dragleave','drop'])drop.addEventListener(type,e=>{e.preventDefault();drop.classList.remove('dragging')});
  drop.addEventListener('drop',e=>{const file=e.dataTransfer?.files?.[0];if(file)importConversionFile(file)});
}

function setConversionWorking(active,message=''){
  const view=document.querySelector('[data-view="conversion"]');if(view)view.classList.toggle('bridge-working',active);
  if(message){const node=document.getElementById('conversionActionStatus');if(node)node.textContent=message}
}

function renderConversionStatus(){
  const targets=document.getElementById('conversionTarget');
  const selected=targets.value;
  targets.innerHTML=(conversionStatus.targets||[]).map(t=>`<option value="${esc(t.id)}">${esc(t.name)} · ${esc(t.maturity)}</option>`).join('');
  if(selected&&[...targets.options].some(o=>o.value===selected))targets.value=selected;
  renderConversionSessions();renderConversionTools();
}

function renderConversionSessions(){
  const root=document.getElementById('conversionSessionList'),sessions=conversionStatus?.sessions||[];
  root.innerHTML=sessions.length?sessions.map(s=>`<button class="bridge-session-button ${conversionCurrent?.id===s.id?'active':''}" data-conversion-session="${esc(s.id)}"><div><b>${esc(s.name)}</b><small>${esc(s.sourceFormat||'source')} → ${esc(s.targetFormat||'choose target')} · ${Number(s.automated||0).toFixed(1)}%</small></div><span>${esc(s.state)}</span></button>`).join(''):'<div class="empty-inline">No conversion sessions yet.</div>';
  root.querySelectorAll('[data-conversion-session]').forEach(button=>button.onclick=()=>loadConversionSession(button.dataset.conversionSession));
}

function renderConversionTools(){
  const tools=conversionStatus?.tools||[],ready=tools.filter(t=>t.ready).length;
  document.getElementById('conversionToolCount').textContent=`${ready}/${tools.length} ready`;
  const root=document.getElementById('conversionTools');
  root.className='bridge-tool-list';
  root.innerHTML=tools.map(t=>{const runnable=t.ready&&t.canExecute&&conversionToolCompatible(t);return `<div class="bridge-tool ${t.ready?'ready':''}" title="${esc(t.role)}"><i></i><div><b>${esc(t.name)}</b><small>${esc(t.role)}</small><div class="bridge-tool-actions"><button class="btn small" data-conversion-tool-config="${esc(t.id)}">${t.configured?'Change path':'Configure'}</button>${runnable?`<button class="btn primary small" data-conversion-tool-run="${esc(t.id)}">Run adapter</button>`:''}</div></div><em>${t.ready?'detected':esc(t.maturity)}</em></div>`}).join('');
  root.querySelectorAll('[data-conversion-tool-config]').forEach(button=>button.onclick=()=>configureConversionTool(button.dataset.conversionToolConfig));
  root.querySelectorAll('[data-conversion-tool-run]').forEach(button=>button.onclick=()=>runConversionAdapterUI(button.dataset.conversionToolRun));
}

function conversionToolCompatible(tool){
  const target=conversionCurrent?.plan?.target?.format||'';
  if(!target)return false;
  if(tool.id==='chunker')return ['bedrock-world','bedrock-template','java-world'].includes(target);
  if(['je2be-resource','packconverter'].includes(tool.id))return ['bedrock-resource','bedrock-addon','bedrock-project','bedrock-world-product'].includes(target);
  if(tool.id==='regolith')return ['bedrock-addon','bedrock-behavior','bedrock-resource','bedrock-project','bedrock-world-product'].includes(target);
  return false;
}

async function configureConversionTool(toolId){
  const tool=(conversionStatus?.tools||[]).find(item=>item.id===toolId);if(!tool)return;
  const value=prompt(`Path to ${tool.name} executable/JAR/script. Leave blank to clear the saved path.`,tool.detectedPath||'');if(value===null)return;
  setConversionWorking(true,`Validating ${tool.name}…`);
  try{await api('/api/conversion/tool',{method:'POST',body:JSON.stringify({toolId,path:value.trim()})});await loadConversionStudio(true);showToast(value.trim()?`${tool.name} configured.`:`${tool.name} path cleared.`)}catch(err){showToast(err.message,true)}finally{setConversionWorking(false)}
}

async function runConversionAdapterUI(toolId){
  if(!conversionCurrent?.plan){showToast('Build a conversion plan first.',true);return}
  const tool=(conversionStatus?.tools||[]).find(item=>item.id===toolId);if(!tool)return;
  setConversionWorking(true,`Running ${tool.name} in an isolated session workspace…`);
  try{const result=await api('/api/conversion/adapter/run',{method:'POST',body:JSON.stringify({sessionId:conversionCurrent.id,toolId,options:{}})});conversionCurrent=result.session;renderConversionSession();await loadConversionStudio(false);showToast(`${tool.name} finished and its output was hash-recorded.`)}catch(err){showToast(err.message,true);await loadConversionSession(conversionCurrent.id,false)}finally{setConversionWorking(false)}
}

function renderConversionShell(){
  const empty=document.getElementById('conversionEmpty'),session=document.getElementById('conversionSession');
  empty.hidden=!!conversionCurrent;session.hidden=!conversionCurrent;
  renderConversionSessions();
}

async function importConversionFile(file){
  if(!file)return;setConversionWorking(true,`Importing and fingerprinting ${file.name}…`);
  try{
    const form=new FormData();form.append('source',file,file.name);
    conversionCurrent=await api('/api/conversion/import',{method:'POST',body:form});
    sessionStorage.setItem('mmv-conversion-session',conversionCurrent.id);
    await loadConversionStudio(true);showToast(`OmniBridge profiled ${conversionCurrent.source?.name||file.name}`);
  }catch(err){showToast(err.message,true)}finally{setConversionWorking(false)}
}

async function importConversionPath(){
  const path=document.getElementById('conversionPath').value.trim();if(!path){showToast('Enter a managed Minecraft or Vault path.',true);return}
  setConversionWorking(true,'Importing the managed path into an immutable conversion session…');
  try{conversionCurrent=await api('/api/conversion/import-path',{method:'POST',body:JSON.stringify({path})});sessionStorage.setItem('mmv-conversion-session',conversionCurrent.id);await loadConversionStudio(true);showToast('Managed content imported into OmniBridge.')}catch(err){showToast(err.message,true)}finally{setConversionWorking(false)}
}

async function loadConversionSession(id,show=true){
  if(!id)return;setConversionWorking(true,'Loading conversion session…');
  try{conversionCurrent=await api(`/api/conversion/session?id=${encodeURIComponent(id)}`);sessionStorage.setItem('mmv-conversion-session',id);renderConversionSession();renderConversionSessions();if(show)showToast(`Loaded ${conversionCurrent.name}`)}catch(err){showToast(err.message,true);conversionCurrent=null;renderConversionShell()}finally{setConversionWorking(false)}
}

function renderConversionSession(){
  renderConversionShell();if(!conversionCurrent)return;
  const source=conversionCurrent.source||{},graph=conversionCurrent.graph||{},summary=graph.summary||{};
  document.getElementById('conversionSourceKicker').textContent=`${String(source.edition||'unknown').toUpperCase()} · ${String(source.kind||'content').toUpperCase()} · ${String(source.format||'archive').toUpperCase()}`;
  document.getElementById('conversionSourceName').textContent=source.name||conversionCurrent.name;
  document.getElementById('conversionSourceDescription').textContent=source.description||'OmniBridge built an immutable Universal Minecraft Content Graph from this source.';
  document.getElementById('conversionSourceFileName').textContent=`${source.filename||''} · SHA-256 ${shortHash(source.sha256)} · ${fmtBytes(source.size||0)}`;
  const chips=[source.edition,source.kind,source.loader,source.gameVersion,source.version,source.minimumEngine&&`engine ${source.minimumEngine}`].filter(Boolean);
  for(const warning of source.warnings||[])chips.push({value:warning,className:'warning'});
  document.getElementById('conversionSourceChips').innerHTML=chips.map(c=>typeof c==='string'?`<span class="bridge-chip ${c==='java'||c==='bedrock'?c:''}">${esc(c)}</span>`:`<span class="bridge-chip ${c.className}">${esc(c.value)}</span>`).join('');
  document.getElementById('conversionName').value=conversionCurrent.plan?.target?.name||source.name||conversionCurrent.name;
  document.getElementById('conversionNamespace').value=conversionCurrent.plan?.target?.namespace||source.namespace||'converted';
  renderConversionGraph(summary);
  if(conversionCurrent.plan){
    const t=conversionCurrent.plan.target||{};setSelectValue('conversionTarget',t.format);document.getElementById('conversionGameVersion').value=t.gameVersion||'';setSelectValue('conversionLoader',t.loader||'vanilla');setSelectValue('conversionStrategy',t.strategy||'balanced');
  }else syncConversionTargetDefaults(true);
  renderConversionPlan();renderConversionOutputs();
}

function renderConversionGraph(s){
  const cards=[['Nodes',s.total||0,'rgba(110,232,220,.13)'],['Assets',s.assets||0,'rgba(114,167,255,.13)'],['Data',s.data||0,'rgba(184,153,255,.13)'],['Logic',s.logic||0,'rgba(242,204,114,.13)'],['World',s.world||0,'rgba(126,226,139,.13)'],['Review',s.reviewRequired||0,'rgba(255,155,155,.13)']];
  document.getElementById('conversionGraphStats').innerHTML=cards.map(([name,value,color])=>`<div class="bridge-graph-stat" style="--stat-accent:${color}"><b>${fmtNum(value)}</b><span>${esc(name)}</span></div>`).join('');
}

function syncConversionTargetDefaults(preserve=false){
  const id=document.getElementById('conversionTarget').value,option=conversionStatus?.targets?.find(t=>t.id===id);
  if(!option)return;
  const version=document.getElementById('conversionGameVersion'),loader=document.getElementById('conversionLoader');
  if(!preserve||!version.value)version.value=option.edition==='bedrock'?'1.26.30':(conversionCurrent?.source?.gameVersion||currentSettings?.gameVersion||'1.21.1');
  if(id==='java-fabric')loader.value='fabric';else if(id==='java-neoforge')loader.value='neoforge';else if(id==='java-forge')loader.value='forge';else if(id==='java-multiloader')loader.value='multiloader';else if(id==='java-world-mod'&&loader.value==='vanilla')loader.value='fabric';else if(!id.startsWith('java-')||['java-datapack','java-resourcepack','java-world'].includes(id))loader.value='vanilla';
}

function conversionTargetRequest(){return{format:document.getElementById('conversionTarget').value,gameVersion:document.getElementById('conversionGameVersion').value.trim(),loader:document.getElementById('conversionLoader').value,name:document.getElementById('conversionName').value.trim(),namespace:document.getElementById('conversionNamespace').value.trim(),strategy:document.getElementById('conversionStrategy').value,includeResourcePack:true,includeDataPack:true,includeSource:true}}

async function buildConversionPlanUI(){
  if(!conversionCurrent){showToast('Import Minecraft content first.',true);return}
  setConversionWorking(true,'Classifying every content node against the selected target…');
  try{conversionCurrent=await api('/api/conversion/plan',{method:'POST',body:JSON.stringify({sessionId:conversionCurrent.id,target:conversionTargetRequest()})});renderConversionSession();showToast('Evidence-backed conversion plan ready.')}catch(err){showToast(err.message,true)}finally{setConversionWorking(false)}
}

async function runConversionUI(){
  if(!conversionCurrent?.plan){showToast('Build a conversion plan first.',true);return}
  setConversionWorking(true,'Generating target-native artifacts, contracts and proof…');
  try{conversionCurrent=await api('/api/conversion/run',{method:'POST',body:JSON.stringify({sessionId:conversionCurrent.id})});renderConversionSession();await loadConversionStudio(false);showToast(conversionCurrent.state==='review-required'?'Artifacts generated; semantic review items remain visible.':'Conversion artifacts generated and validated.')}catch(err){showToast(err.message,true);await loadConversionSession(conversionCurrent.id,false)}finally{setConversionWorking(false)}
}

function renderConversionPlan(){
  const plan=conversionCurrent?.plan,run=document.getElementById('conversionRun'),state=document.getElementById('conversionPlanState');
  run.disabled=!plan;
  if(!plan){state.textContent='Not planned';state.className='bridge-plan-state';document.getElementById('conversionCoveragePercent').textContent='—';document.getElementById('conversionCoverage').innerHTML='<div class="empty-inline">Build a plan to classify every graph node.</div>';document.getElementById('conversionSteps').innerHTML='<div class="empty-inline">No plan yet.</div>';document.getElementById('conversionReview').innerHTML='<div class="empty-inline">No review queue until a target is planned.</div>';document.getElementById('conversionReviewCount').textContent='0 items';return}
  const c=plan.coverage||{},outstanding=(c.review||0)+(c.blocked||0)+(c.toolAssisted||0);
  state.textContent=outstanding?`${outstanding} items need review/tooling`:'Fully automated plan';state.className=`bridge-plan-state ${outstanding?'review':'ready'}`;
  document.getElementById('conversionPlanState').title=plan.planSha256||'';
  document.getElementById('conversionCoveragePercent').textContent=`${Number(c.automatedPercent||0).toFixed(1)}%`;
  const levels=[['exact',c.exact,'#75e28a'],['translated',c.translated,'#72c8ff'],['generated',c.generated,'#b899ff'],['tool-assisted',c.toolAssisted,'#f2cc72'],['review',c.review,'#ffb082'],['blocked',c.blocked,'#ff8e8e']];
  const total=Math.max(1,c.total||0);
  document.getElementById('conversionCoverage').innerHTML=`<div class="bridge-coverage-score"><div class="bridge-coverage-ring" style="--coverage:${Number(c.automatedPercent||0)}"><b>${Math.round(Number(c.automatedPercent||0))}%</b></div><div class="bridge-coverage-bars">${levels.map(([name,value,color])=>`<div class="bridge-coverage-row"><span>${esc(name)}</span><div class="bridge-coverage-track"><i style="--value:${Number(value||0)/total*100}%;--bar:${color}"></i></div><b>${Number(value||0)}</b></div>`).join('')}</div></div>`;
  document.getElementById('conversionSteps').innerHTML=(plan.steps||[]).map(step=>`<div class="bridge-step"><span class="bridge-step-index">${step.order}</span><div><b>${esc(step.title)}</b><p>${esc(step.description)}</p>${step.toolId?`<small>Adapter: ${esc(step.toolId)}</small>`:''}</div><span class="bridge-level ${esc(step.level)}">${esc(step.level)}</span></div>`).join('')||'<div class="empty-inline">No execution steps.</div>';
  const reviews=plan.reviewQueue||[];document.getElementById('conversionReviewCount').textContent=`${reviews.length} item${reviews.length===1?'':'s'}`;
  document.getElementById('conversionReview').innerHTML=reviews.length?reviews.map(item=>`<article class="bridge-review-item ${item.level==='blocked'?'blocked':''}"><span class="bridge-level ${esc(item.level)}">${esc(item.level)}</span><div><b>${esc(item.category)} · ${esc(item.title)}</b><p>${esc(item.reason)}</p><small>${esc(item.suggestedRoute)}</small></div><span>${esc(item.severity)}</span></article>`).join(''):'<div class="empty-inline">No unresolved semantic work for this target plan.</div>';
  document.getElementById('conversionActionStatus').textContent=`Plan ${shortHash(plan.planSha256)} · ${c.completenessState||'ready'}`;
}

function renderConversionOutputs(){
  const outputs=conversionCurrent?.outputs||[],runs=conversionCurrent?.adapterRuns||[],root=document.getElementById('conversionOutputs');document.getElementById('conversionDownloadReport').disabled=!conversionCurrent;
  const cards=outputs.map((o,index)=>`<article class="bridge-output"><span class="bridge-output-icon">${icon(o.kind==='proof'?'shield':'download')}</span><div><b title="${esc(o.name)}">${esc(o.name)}</b><small>${fmtBytes(o.size)} · ${shortHash(o.sha256)} · ${esc(o.kind)}</small><div class="bridge-output-validation ${o.validated?'':'bad'}">${esc(o.validation||'Not validated')}</div></div><a class="btn primary small" href="${conversionDownloadURL(conversionCurrent.id,'output',index)}" download><span data-icon="download"></span> Download</a></article>`);
  runs.forEach((run,index)=>cards.push(`<article class="bridge-output bridge-adapter-run"><span class="bridge-output-icon">${icon(run.state==='succeeded'?'gear':'warning')}</span><div><b>${esc(run.toolName)}</b><small>${esc(run.state)} · ${esc(run.completedAt||run.startedAt||'')}</small><div class="bridge-output-validation ${run.state==='succeeded'?'':'bad'}">${esc(run.error||`${run.outputs?.length||0} validated adapter output(s)`)}</div></div>${run.logPath?`<a class="btn small" href="${conversionDownloadURL(conversionCurrent.id,'adapter-log',index)}" download>Log</a>`:''}</article>`));
  root.innerHTML=cards.length?cards.join(''):'<div class="empty-inline">Run the plan to create outputs.</div>';
  renderConversionTools();if(typeof injectIcons==='function')injectIcons(root);
}

async function deleteConversionSession(){
  if(!conversionCurrent)return;
  if(!confirm(`Delete conversion session “${conversionCurrent.name}”? The original uploaded artifact is only removed from this conversion session; your managed Minecraft file is untouched.`))return;
  const id=conversionCurrent.id;setConversionWorking(true,'Removing conversion session…');
  try{await api('/api/conversion/reset',{method:'POST',body:JSON.stringify({sessionId:id})});conversionCurrent=null;sessionStorage.removeItem('mmv-conversion-session');await loadConversionStudio(true);showToast('Conversion session removed.')}catch(err){showToast(err.message,true)}finally{setConversionWorking(false)}
}

function conversionDownloadURL(id,kind,index){const params=new URLSearchParams({id,kind,token:TOKEN});if(index!==undefined)params.set('index',String(index));return `/api/conversion/download?${params}`}
function shortHash(value){value=String(value||'');return value?`${value.slice(0,10)}…${value.slice(-6)}`:'—'}
function setSelectValue(id,value){const select=document.getElementById(id);if(select&&[...select.options].some(o=>o.value===value))select.value=value}
