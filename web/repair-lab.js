'use strict';

let repairLabStatus=null;
let repairCurrentSession=null;
let repairPollTimer=null;
let repairWired=false;
let repairStatusLoading=false;
let repairRefreshQueued=false;

function wireRepairLab(){
  if(repairWired)return;
  repairWired=true;
  const source=document.getElementById('repairSourceFile');
  const drop=document.getElementById('repairDropZone');
  document.getElementById('repairChooseSource').onclick=()=>source.click();
  source.onchange=()=>{const file=source.files?.[0];if(file)importRepairSource(file);source.value=''};
  ['dragenter','dragover'].forEach(name=>drop.addEventListener(name,event=>{event.preventDefault();drop.classList.add('drag')}));
  ['dragleave','drop'].forEach(name=>drop.addEventListener(name,event=>{event.preventDefault();drop.classList.remove('drag')}));
  drop.ondrop=event=>{const file=event.dataTransfer?.files?.[0];if(file)importRepairSource(file)};
  drop.onclick=event=>{if(!event.target.closest('a,button,input,select'))source.click()};
  document.getElementById('repairRefresh').onclick=()=>loadRepairLab(true);
  document.getElementById('repairSessionList').onclick=event=>{const button=event.target.closest('[data-repair-session]');if(button)loadRepairSession(button.dataset.repairSession)};
  document.getElementById('repairPrepare').onclick=prepareRepairMigration;
  document.getElementById('repairRunBuild').onclick=()=>runRepairAction('build');
  document.getElementById('repairRunTest').onclick=()=>runRepairAction('test');
  document.getElementById('repairRunClean').onclick=()=>runRepairAction('clean');
  document.getElementById('repairCancel').onclick=cancelRepairAction;
  document.getElementById('repairReset').onclick=resetRepairSession;
  document.getElementById('repairExport').onclick=exportRepairBundle;
  document.getElementById('repairBrainSearch').onclick=searchRepairBrain;
  document.getElementById('repairBrainQuery').onkeydown=event=>{if(event.key==='Enter')searchRepairBrain()};
}

async function loadRepairLab(force=false){
  if(repairStatusLoading){if(force)repairRefreshQueued=true;return;}
  if(repairLabStatus&&!force){
    renderRepairLabStatus();
    if(repairCurrentSession)renderRepairSession(repairCurrentSession);
    return;
  }
  repairStatusLoading=true;
  try{
    repairLabStatus=await api('/api/repair-lab/status');
    renderRepairLabStatus();
    const id=repairCurrentSession?.id||repairLabStatus.sessions?.[0]?.id;
    if(id)await loadRepairSession(id,false);
  }catch(err){
    showToast(`Repair Lab: ${err.message}`,true);
    document.getElementById('repairMetrics').innerHTML=`<div class="card empty">${esc(err.message)}</div>`;
  }finally{repairStatusLoading=false;if(repairRefreshQueued){repairRefreshQueued=false;setTimeout(()=>loadRepairLab(true),0)}}
}

function renderRepairLabStatus(){
  if(!repairLabStatus)return;
  const atlas=repairLabStatus.atlas||{},brain=repairLabStatus.brain||{};
  document.getElementById('repairMetrics').innerHTML=[
    repairMetric('Version Atlas',fmtNum(atlas.mojangVersions||0),`${esc(atlas.latestRelease||'unknown')} release · ${fmtNum(atlas.runtimeLibraryRows||0)} runtime library rows`,'compass'),
    repairMetric('Compatibility Brain',fmtNum(brain.knowledgeDocuments||0),`${fmtNum(brain.toolchainReleases||0)} toolchain releases · ${fmtNum(brain.repairRecords||0)} repair records`,'radar'),
    repairMetric('Durable sessions',fmtNum(repairLabStatus.sessions?.length||0),'Restart-safe manifests, logs, receipts and exports','save'),
    repairMetric('Security boundary','8 gates','Immutable input · strict ZIP intake · explicit execution','shield')
  ].join('');
  document.getElementById('repairTrustState').textContent=repairLabStatus.securityBoundaries?.immutableOriginal?'Immutable intake verified by design':'Safety metadata unavailable';
  document.getElementById('repairSessionCount').textContent=repairLabStatus.sessions?.length||0;
  document.getElementById('repairBrainStatus').textContent=brain.ready?`${fmtNum(brain.minecraftVersions)} Minecraft versions · SQLite ${brain.sqliteVersion||''}`:(repairLabStatus.brainError||'Brain unavailable');
  renderRepairSessions();
}

function repairMetric(label,value,note,ic){return `<article class="card repair-metric"><span>${icon(ic)}</span><div><small>${esc(label)}</small><strong>${esc(value)}</strong><p>${note}</p></div></article>`}

function renderRepairSessions(){
  const root=document.getElementById('repairSessionList'),rows=repairLabStatus?.sessions||[];
  root.innerHTML=rows.length?rows.map(row=>`<button class="repair-session-item ${repairCurrentSession?.id===row.id?'active':''}" data-repair-session="${esc(row.id)}"><span class="repair-session-dot state-${esc(row.state)}"></span><span><strong>${esc(row.name||row.source||row.id)}</strong><small>${esc(row.loader||'unknown')} · ${esc(row.gameVersion||'version unknown')}${row.target?` → ${esc(row.target)}`:''}</small><em>${esc(row.phase||row.state)}</em></span></button>`).join(''):'<div class="empty">No repair sessions yet.</div>';
}

async function importRepairSource(file){
  if(!file?.name?.toLowerCase().endsWith('.zip')){showToast('Repair Lab accepts source ZIP archives.',true);return}
  const button=document.getElementById('repairChooseSource'),old=button.innerHTML;
  button.disabled=true;button.innerHTML=`${icon('refresh')} Verifying archive…`;
  document.getElementById('repairDropZone').classList.add('busy');
  try{
    const form=new FormData();form.append('source',file,file.name);
    repairCurrentSession=await api('/api/repair-lab/import',{method:'POST',body:form});
    renderRepairSession(repairCurrentSession);
    showToast(`Imported ${file.name} as an immutable repair session.`);
    await loadRepairLab(true);
  }catch(err){showToast(err.message,true)}finally{button.disabled=false;button.innerHTML=old;document.getElementById('repairDropZone').classList.remove('busy')}
}

async function loadRepairSession(id,renderStatus=true){
  if(!id)return;
  try{
    repairCurrentSession=await api(`/api/repair-lab/session?id=${encodeURIComponent(id)}`);
    renderRepairSession(repairCurrentSession);
    if(renderStatus)renderRepairSessions();
    scheduleRepairPoll(repairCurrentSession.state==='running');
  }catch(err){showToast(err.message,true)}
}

function renderRepairSession(session){
  if(!session)return;
  document.getElementById('repairEmpty').hidden=true;
  document.getElementById('repairSessionWorkspace').hidden=false;
  document.getElementById('repairSessionTitle').textContent=session.name||session.source?.filename||session.id;
  const state=document.getElementById('repairSessionState');state.textContent=session.state||'unknown';state.className=`repair-state state-${session.state||'unknown'}`;
  document.getElementById('repairSessionPhase').textContent=session.phase||'';
  document.getElementById('repairSessionIdentity').innerHTML=`Session <code>${esc(session.id)}</code> · source <code>${esc(shortHash(session.source?.sha256))}</code> · immutable tree <code>${esc(shortHash(session.source?.treeSha256))}</code> · ${fmtNum(session.source?.fileCount||0)} files / ${fmtBytes(session.source?.extractedBytes||0)}`;
  document.getElementById('repairReceiptJSON').href=repairDownloadURL(session.id,'receipt-json');
  document.getElementById('repairReceiptMarkdown').href=repairDownloadURL(session.id,'receipt-markdown');
  const project=session.project||{};
  document.getElementById('repairConfidence').textContent=`${project.confidence||'unknown'} confidence`;
  document.getElementById('repairProjectSummary').innerHTML=`<div class="repair-fact-grid">${repairFact('Build',project.buildSystem||'unknown')}${repairFact('Wrapper',project.wrapper||'none')}${repairFact('Loader',project.loader||'unknown')}${repairFact('Minecraft',project.gameVersion||'unknown')}${repairFact('Java',project.javaMajor?`Java ${project.javaMajor}`:'unknown')}${repairFact('Modules',(project.modules||[]).length||1)}</div><div class="repair-signal-row">${(project.signals||[]).map(x=>`<span class="tag source">${esc(x)}</span>`).join('')}${(project.metadataFiles||[]).slice(0,5).map(x=>`<span class="tag">${esc(x)}</span>`).join('')}</div><p class="repair-fingerprint">Project fingerprint <code>${esc(shortHash(project.fingerprint,18))}</code>${project.wrapperSha256?` · wrapper <code>${esc(shortHash(project.wrapperSha256,18))}</code>`:''}</p>`;
  if(!document.getElementById('repairTargetGame').dataset.touched)document.getElementById('repairTargetGame').value=session.target?.gameVersion||project.gameVersion||currentSettings?.gameVersion||'1.21.1';
  if(!document.getElementById('repairTargetLoader').dataset.touched)document.getElementById('repairTargetLoader').value=session.target?.loader||(['fabric','forge','neoforge','quilt','vanilla'].includes(project.loader)?project.loader:(currentSettings?.loader||'fabric'));
  renderRepairResolution(session);
  renderRepairChanges(session.changes||[]);
  renderRepairArtifacts(session);
  renderRepairExports(session);
  renderRepairWarnings(session);
  renderRepairExecution(session);
  renderRepairSessions();
}

function repairFact(label,value){return `<div><small>${esc(label)}</small><strong>${esc(value)}</strong></div>`}
function shortHash(value,length=12){value=String(value||'');return value?`${value.slice(0,length)}${value.length>length?'…':''}`:'unavailable'}

function renderRepairResolution(session){
  const root=document.getElementById('repairResolution'),resolution=session.resolution;
  document.getElementById('repairResolutionBadge').textContent=resolution?`${resolution.gameVersion} / ${resolution.loader}`:'Not staged';
  if(!resolution){root.innerHTML='<div class="mini-note">Choose a target to resolve exact Java, loader, mappings, game artifacts, build plugins, and legacy routes.</div>';return}
  const groups=[['Game artifacts',resolution.gameArtifacts],['Build toolchains',resolution.buildToolchains],['Mappings',resolution.mappings],['Compatibility routes',resolution.compatibilityRoutes]].filter(([,rows])=>rows?.length);
  root.innerHTML=`<div class="repair-resolution-head">${repairFact('Exists',resolution.exists?'Yes':'No')}${repairFact('Loader support',resolution.loaderSupported?'Proven':'Review')}${repairFact('Java',`Java ${resolution.javaMajor||'?'}`)}${repairFact('Loader version',resolution.loaderVersion||'manual')}</div>${groups.map(([label,rows])=>`<div class="repair-resolution-group"><small>${esc(label)}</small>${rows.map(row=>`<span title="${esc(row.reason||'')}"><b>${esc(row.name||row.id)}</b><em>${esc(row.version||row.channel||'route')}</em></span>`).join('')}</div>`).join('')}`;
}

function renderRepairChanges(changes){
  document.getElementById('repairChangeCount').textContent=changes.length;
  document.getElementById('repairChanges').innerHTML=changes.length?changes.map(change=>`<article class="repair-change"><div><strong>${esc(change.field)}</strong><small>${esc(change.file)}</small></div><code>${esc(change.before||'∅')}</code><span>→</span><code>${esc(change.after||'∅')}</code><p>${esc(change.reason||'')}</p></article>`).join(''):'<div class="empty">No migration staged.</div>';
}

function renderRepairArtifacts(session){
  const rows=session.artifacts||[];
  document.getElementById('repairArtifactCount').textContent=rows.length;
  document.getElementById('repairArtifacts').innerHTML=rows.length?rows.map((item,index)=>`<article class="repair-output"><div><strong>${esc(item.name)}</strong><small>${esc(item.kind)} · ${fmtBytes(item.size)} · Java ${esc(item.javaMajor||'?')} · ${fmtNum(item.classCount||0)} classes</small><code>${esc(shortHash(item.sha256,18))}</code></div><a class="btn small" href="${repairDownloadURL(session.id,'artifact',index)}">${icon('download')} Download</a></article>`).join(''):'<div class="empty">Build outputs appear here after a successful wrapper run.</div>';
}

function renderRepairExports(session){
  const rows=session.exports||[];
  document.getElementById('repairExportCount').textContent=rows.length;
  document.getElementById('repairExports').innerHTML=rows.length?rows.map((item,index)=>`<article class="repair-output"><div><strong>${esc(item.name)}</strong><small>${esc(item.kind)} · ${fmtBytes(item.size)}</small><code>${esc(shortHash(item.sha256,18))}</code></div><a class="btn small" href="${repairDownloadURL(session.id,'export',index)}">${icon('download')} Download</a></article>`).join(''):'<div class="empty">No proof bundle exported.</div>';
}

function renderRepairWarnings(session){
  const rows=[...(session.warnings||[])];if(session.lastError)rows.unshift(session.lastError);
  document.getElementById('repairWarnings').innerHTML=rows.length?`<div class="repair-panel-head"><div><span class="eyebrow">REMAINING PROOF + WARNINGS</span><h3>Do not hide uncertainty</h3></div><span>${rows.length}</span></div><div class="repair-warning-list">${rows.map(row=>`<p>${icon('shield')}<span>${esc(row)}</span></p>`).join('')}</div>`:`<div class="repair-clear">${icon('shield')} No active warnings in the current stage.</div>`;
}

function renderRepairExecution(session){
  const running=session.state==='running',latest=(session.runs||[]).at(-1);
  document.getElementById('repairExecutionState').textContent=latest?`${latest.action} · ${latest.state}`:(session.state||'Not running');
  document.getElementById('repairCancel').disabled=!running;
  ['repairRunBuild','repairRunTest','repairRunClean','repairPrepare','repairReset','repairExport'].forEach(id=>document.getElementById(id).disabled=running);
  document.getElementById('repairLog').textContent=latest?.logTail||latest?.error||'No build log yet.';
}

async function prepareRepairMigration(){
  if(!repairCurrentSession){showToast('Import or select a repair session first.',true);return}
  const targetGame=document.getElementById('repairTargetGame').value.trim(),targetLoader=document.getElementById('repairTargetLoader').value;
  if(!targetGame){showToast('Enter a target Minecraft version.',true);return}
  const button=document.getElementById('repairPrepare'),old=button.innerHTML;button.disabled=true;button.innerHTML=`${icon('refresh')} Resolving…`;
  try{
    repairCurrentSession=await api('/api/repair-lab/prepare',{method:'POST',body:JSON.stringify({sessionId:repairCurrentSession.id,targetGameVersion:targetGame,targetLoader})});
    renderRepairSession(repairCurrentSession);showToast(`Staged ${repairCurrentSession.changes?.length||0} traceable migration edits.`);await loadRepairLab(true);
  }catch(err){showToast(err.message,true)}finally{button.disabled=false;button.innerHTML=old}
}

async function runRepairAction(action){
  if(!repairCurrentSession){showToast('Select a repair session first.',true);return}
  if(!document.getElementById('repairExecutionAck').checked){showToast('Acknowledge build-script code execution for this action.',true);return}
  try{
    repairCurrentSession=await api('/api/repair-lab/run',{method:'POST',body:JSON.stringify({sessionId:repairCurrentSession.id,action,confirmCode:repairLabStatus?.executionPhrase||'',timeoutMinutes:30})});
    document.getElementById('repairExecutionAck').checked=false;renderRepairSession(repairCurrentSession);scheduleRepairPoll(true);showToast(`${action[0].toUpperCase()+action.slice(1)} started in the disposable working copy.`);
  }catch(err){showToast(err.message,true)}
}

async function cancelRepairAction(){
  if(!repairCurrentSession)return;
  try{await api('/api/repair-lab/cancel',{method:'POST',body:JSON.stringify({sessionId:repairCurrentSession.id})});showToast('Cancellation requested.');scheduleRepairPoll(true)}catch(err){showToast(err.message,true)}
}

async function resetRepairSession(){
  if(!repairCurrentSession)return;
  try{repairCurrentSession=await api('/api/repair-lab/reset',{method:'POST',body:JSON.stringify({sessionId:repairCurrentSession.id})});renderRepairSession(repairCurrentSession);showToast('Working copy rolled back to the verified immutable source.');await loadRepairLab(true)}catch(err){showToast(err.message,true)}
}

async function exportRepairBundle(){
  if(!repairCurrentSession)return;
  const button=document.getElementById('repairExport'),old=button.innerHTML;button.disabled=true;button.innerHTML=`${icon('refresh')} Packaging…`;
  try{repairCurrentSession=await api('/api/repair-lab/export',{method:'POST',body:JSON.stringify({sessionId:repairCurrentSession.id,includeArtifacts:true})});renderRepairSession(repairCurrentSession);showToast('Prepared-source and proof bundle exports are ready.')}catch(err){showToast(err.message,true)}finally{button.disabled=false;button.innerHTML=old}
}

function scheduleRepairPoll(active){
  clearTimeout(repairPollTimer);repairPollTimer=null;
  if(!active||!repairCurrentSession)return;
  repairPollTimer=setTimeout(async()=>{const id=repairCurrentSession?.id;if(!id)return;await loadRepairSession(id,false);if(repairCurrentSession?.state==='running')scheduleRepairPoll(true);else loadRepairLab(true)},1100);
}

function repairDownloadURL(id,kind,index){const params=new URLSearchParams({id,kind,token:TOKEN});if(index!==undefined)params.set('index',String(index));return `/api/repair-lab/download?${params}`}

async function searchRepairBrain(){
  const query=document.getElementById('repairBrainQuery').value.trim(),kind=document.getElementById('repairBrainKind').value,root=document.getElementById('repairBrainResults');
  if(!query){showToast('Enter a repair, version, API, or tool question.',true);return}
  root.innerHTML='<div class="card loading">Searching the local compatibility brain…</div>';
  try{
    const data=await api(`/api/brain/search?q=${encodeURIComponent(query)}&kind=${encodeURIComponent(kind)}&limit=30`);
    root.innerHTML=data.results?.length?data.results.map(item=>`<article class="card repair-brain-result"><div><span class="tag source">${esc(item.kind||'knowledge')}</span>${item.category?`<span class="tag">${esc(item.category)}</span>`:''}</div><h3>${esc(item.name)}</h3><p>${esc(item.snippet||item.summary||'')}</p><div class="repair-result-actions">${item.url?`<button class="btn small" data-repair-open="${esc(item.url)}">${icon('external')} Official</button>`:''}${item.repository?`<button class="btn small" data-repair-open="${esc(item.repository)}">${icon('external')} Repository</button>`:''}<span>${esc(item.maturity||item.status||'')}</span></div></article>`).join(''):'<div class="card empty">No exact knowledge record matched. Try a loader, Minecraft version, error symbol, or tool name.</div>';
    root.querySelectorAll('[data-repair-open]').forEach(button=>button.onclick=()=>openExternal(button.dataset.repairOpen));
  }catch(err){root.innerHTML=`<div class="card empty">${esc(err.message)}</div>`;showToast(err.message,true)}
}
