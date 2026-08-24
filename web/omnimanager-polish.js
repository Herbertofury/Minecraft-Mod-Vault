(()=>{
'use strict';
const STORE='mmv.omnimanager.ui.v1';
const state={view:'grid',density:'comfortable'};
try{Object.assign(state,JSON.parse(localStorage.getItem(STORE)||'{}'))}catch(_){ }
const save=()=>{try{localStorage.setItem(STORE,JSON.stringify(state))}catch(_){}};
const manager=()=>document.querySelector('.view[data-view="manager"]');
function decorate(){
 const root=manager(); if(!root)return;
 root.classList.add('omni-premium');
 root.querySelectorAll('button,a,input,select').forEach(el=>{
  if(!el.getAttribute('aria-label')&&!el.textContent.trim()&&el.getAttribute('title'))el.setAttribute('aria-label',el.getAttribute('title'));
 });
 const library=root.querySelector('#omniLibrary,.omni-library,#managerGroups,.manager-groups');
 if(library){library.classList.toggle('grid',state.view==='grid');library.classList.toggle('list',state.view==='list');library.dataset.density=state.density;}
 root.querySelectorAll('[data-omni-view]').forEach(btn=>{
  btn.classList.toggle('active',btn.dataset.omniView===state.view);
  btn.setAttribute('aria-pressed',String(btn.dataset.omniView===state.view));
  if(!btn.dataset.omniBound){btn.dataset.omniBound='1';btn.addEventListener('click',()=>{state.view=btn.dataset.omniView==='list'?'list':'grid';save();decorate();});}
 });
}
function focusSearch(){manager()?.querySelector('#omniSearch,#managerSearch,input[type="search"]')?.focus()}
document.addEventListener('keydown',ev=>{
 const root=manager(); if(!root?.classList.contains('active'))return;
 if(ev.key==='/'&&!/INPUT|TEXTAREA|SELECT/.test(document.activeElement?.tagName||'')){ev.preventDefault();focusSearch();}
 if(ev.key==='Escape'){root.querySelector('[data-close-drawer],#omniCloseDetails,.drawer-close')?.click();}
 if((ev.ctrlKey||ev.metaKey)&&ev.key.toLowerCase()==='a'&&!/INPUT|TEXTAREA|SELECT/.test(document.activeElement?.tagName||'')){
  const all=root.querySelector('[data-select-all],#omniSelectAll');if(all){ev.preventDefault();all.click();}
 }
});
const observer=new MutationObserver(decorate);
document.addEventListener('DOMContentLoaded',()=>{decorate();const r=manager();if(r)observer.observe(r,{subtree:true,childList:true,attributes:true,attributeFilter:['class']});});
window.OmniManagerPolish={decorate,focusSearch};
})();
