'use strict';

window.OmniManager = (() => {
  const state = {
    payload: null,
    selected: new Set(),
    wired: false,
    loading: false,
    enriching: false,
    generation: 0,
    view: 'grid',
  };

  try { state.view = localStorage.getItem('mmv-manager-view') === 'list' ? 'list' : 'grid'; } catch {}

  const $ = id => document.getElementById(id);
  const items = () => state.payload?.items || [];
  const itemByID = id => items().find(item => item.id === id);
  const text = value => String(value ?? '').trim();
  const list = value => Array.isArray(value) ? value.filter(Boolean) : [];
  const unique = values => [...new Set(values.filter(Boolean))];
  const humanKind = kind => ({
    mod: 'Java Mod', plugin: 'Server Plugin', resourcepack: 'Resource Pack', shader: 'Shader Pack', datapack: 'Data Pack', world: 'Java World', download: 'Vault Download',
    behaviorpack: 'Behavior Pack', 'behaviorpack-dev': 'Behavior Pack (Development)', 'resourcepack-bedrock': 'Resource Pack', 'resourcepack-bedrock-dev': 'Resource Pack (Development)', skinpack: 'Skin Pack', template: 'World Template', 'world-bedrock': 'Bedrock World', 'addon-package': 'Add-on Package', 'world-package': 'World Package', 'template-package': 'Template Package', addon: 'Bedrock Add-on',
  }[kind] || text(kind).replaceAll('-', ' ').replace(/\b\w/g, c => c.toUpperCase()));
  const humanEdition = edition => ({java: 'Java', bedrock: 'Bedrock', server: 'Server'}[edition] || edition || 'Unknown');
  const providerLabel = source => source.providerLabel || providerName(source.provider);
  const fmtDate = value => {
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? text(value) : date.toLocaleString();
  };
  const confidence = value => `${Math.round((Number(value) || 0) * 100)}%`;

  function authenticatedArt(raw) {
    raw = text(raw);
    if (!raw) return '';
    if (raw.startsWith('/')) {
      const url = new URL(raw, location.href);
      url.searchParams.set('token', TOKEN);
      return `${url.pathname}${url.search}`;
    }
    return safeImg(raw);
  }

  function itemArt(item, large = false) {
    const src = authenticatedArt(item.localArtUrl || item.remoteArtUrl);
    if (src) return `<img class="omni-art-image ${large ? 'large' : ''}" src="${esc(src)}" alt="${esc(item.name || item.filename)} artwork">`;
    const glyph = item.edition === 'bedrock' ? 'bedrock' : item.kind === 'plugin' ? 'gear' : item.kind?.includes('pack') || item.kind === 'resourcepack' || item.kind === 'shader' ? 'pack' : 'mod';
    return `<div class="omni-art-fallback ${large ? 'large' : ''}">${icon(glyph)}<span>${esc((item.name || item.filename || '?').slice(0, 1).toUpperCase())}</span></div>`;
  }

  function scanState(message, tone = '') {
    const root = $('managerScanState');
    if (!root) return;
    root.className = `omni-scan-state ${tone}`.trim();
    root.innerHTML = message;
  }

  async function refresh(options = {}) {
    if (state.loading && !options.force) return;
    const generation = ++state.generation;
    state.loading = true;
    scanState(`<span class="omni-pulse"></span><strong>Inspecting local files…</strong><span>Reading embedded identities, artwork, manifests, hashes, profiles, and reversible state.</span>`, 'working');
    try {
      const local = await api('/api/library?enrich=false');
      if (generation !== state.generation) return;
      state.payload = local;
      reconcileSelection();
      render();
      scanState(`<span class="omni-pulse"></span><strong>${local.items.length.toLocaleString()} local items loaded instantly.</strong><span>Now joining exact hashes and embedded identities across every configured source…</span>`, 'working');
      state.enriching = true;
      const enriched = await api(`/api/library?enrich=true&refresh=${options.force ? 'true' : 'false'}`);
      if (generation !== state.generation) return;
      state.payload = enriched;
      reconcileSelection();
      render();
      const warnings = list(enriched.warnings);
      scanState(`<span class="omni-scan-ok">✓</span><strong>Identity graph ready.</strong><span>${enriched.summary.exactSource.toLocaleString()} exact provider identities · ${enriched.summary.withArt.toLocaleString()} with art · ${enriched.summary.updates.toLocaleString()} safe updates${warnings.length ? ` · ${warnings.length} provider warning${warnings.length === 1 ? '' : 's'}` : ''}.</span>`, warnings.length ? 'warning' : 'ready');
    } catch (error) {
      scanState(`<span class="omni-scan-bad">!</span><strong>Scan failed.</strong><span>${esc(error.message)}</span>`, 'error');
      showToast(`OmniManager scan failed: ${error.message}`, true);
    } finally {
      if (generation === state.generation) {
        state.loading = false;
        state.enriching = false;
      }
    }
  }

  function reconcileSelection() {
    const live = new Set(items().map(item => item.id));
    for (const id of state.selected) if (!live.has(id)) state.selected.delete(id);
  }

  function render() {
    if (!state.payload) return;
    populateFilters();
    renderSummary();
    renderTrust();
    renderItems();
    renderBulk();
    $('managerGridView')?.classList.toggle('active', state.view === 'grid');
    $('managerListView')?.classList.toggle('active', state.view === 'list');
  }

  function renderSummary() {
    const root = $('managerSummary');
    if (!root) return;
    const s = state.payload.summary || {};
    const cards = [
      ['all', s.total || 0, 'Installed items', 'Everything across every profile'],
      ['update', s.updates || 0, 'Safe updates', 'Exact identity + target compatibility'],
      ['current', s.current || 0, 'Up to date', 'Newest compatible provider file'],
      ['identity', s.exactSource || 0, 'Exact identities', 'Hash, fingerprint, UUID, or declared source'],
      ['art', s.withArt || 0, 'Artwork recovered', 'Embedded or provider-backed'],
      ['bedrock', s.bedrock || 0, 'Bedrock content', 'Stable, Preview, and custom roots'],
      ['modified', s.modified || 0, 'Protected builds', 'Patched/custom files never blindly replaced'],
      ['disabled', s.disabled || 0, 'Disabled', 'Recoverable through Vault receipts'],
    ];
    root.innerHTML = cards.map(([filter, count, label, note]) => `<button class="card omni-kpi" data-summary-filter="${filter}"><span>${esc(label)}</span><strong>${Number(count).toLocaleString()}</strong><small>${esc(note)}</small></button>`).join('');
  }

  function renderTrust() {
    const total = Math.max(1, items().length);
    const identified = items().filter(item => item.provenanceConfidence >= .9 || list(item.sources).some(source => source.exact)).length;
    const withArt = items().filter(item => item.localArtUrl || item.remoteArtUrl).length;
    const score = Math.round((identified / total) * 100);
    if ($('managerTrustScore')) $('managerTrustScore').textContent = `${score}%`;
    if ($('managerTrustText')) $('managerTrustText').textContent = `${identified.toLocaleString()} exact identities and ${withArt.toLocaleString()} recovered artworks across ${total.toLocaleString()} managed items.`;
    const ring = document.querySelector('.omni-trust-ring');
    if (ring) ring.style.setProperty('--trust', `${score * 3.6}deg`);
  }

  function preserveSelect(id, html) {
    const element = $(id);
    if (!element) return;
    const previous = element.value;
    element.innerHTML = html;
    if ([...element.options].some(option => option.value === previous)) element.value = previous;
  }

  function populateFilters() {
    const payload = state.payload;
    preserveSelect('managerProfile', `<option value="all">Every profile</option>${list(payload.profiles).map(profile => `<option value="${esc(profile.name)}">${esc(profile.name)} · ${esc(profile.channel || humanEdition(profile.edition))}${profile.exists ? '' : ' (not found)'}</option>`).join('')}`);
    const kinds = unique(items().map(item => item.kind)).sort((a, b) => humanKind(a).localeCompare(humanKind(b)));
    preserveSelect('managerKind', `<option value="all">Every content type</option>${kinds.map(kind => `<option value="${esc(kind)}">${esc(humanKind(kind))}</option>`).join('')}`);
    const sources = new Map();
    for (const item of items()) for (const source of list(item.sources)) sources.set(source.provider, providerLabel(source));
    preserveSelect('managerSource', `<option value="all">Every source</option>${[...sources].sort((a, b) => a[1].localeCompare(b[1])).map(([id, label]) => `<option value="${esc(id)}">${esc(label)}</option>`).join('')}<option value="unresolved">Unresolved / local only</option>`);
  }

  function currentFilters() {
    return {
      query: text($('managerQuery')?.value).toLowerCase(),
      profile: $('managerProfile')?.value || 'all',
      edition: $('managerEdition')?.value || 'all',
      kind: $('managerKind')?.value || 'all',
      status: $('managerStatus')?.value || 'all',
      source: $('managerSource')?.value || 'all',
      sort: $('managerSort')?.value || 'name',
    };
  }

  function filteredItems() {
    const f = currentFilters();
    const result = items().filter(item => {
      if (f.profile !== 'all' && item.profile !== f.profile) return false;
      if (f.edition !== 'all' && item.edition !== f.edition) return false;
      if (f.kind !== 'all' && item.kind !== f.kind) return false;
      if (f.status === 'disabled' && item.enabled) return false;
      if (f.status !== 'all' && f.status !== 'disabled' && item.updateStatus !== f.status) return false;
      if (f.source === 'unresolved' && list(item.sources).length) return false;
      if (f.source !== 'all' && f.source !== 'unresolved' && !list(item.sources).some(source => source.provider === f.source)) return false;
      if (f.query) {
        const haystack = [item.name, item.filename, item.summary, item.description, item.modId, item.uuid, item.profile, item.edition, item.kind, item.installedVersion, item.latestVersion, ...list(item.authors), ...list(item.loaders), ...list(item.gameVersions), ...list(item.sources).flatMap(source => [source.provider, source.providerLabel, source.title, source.author, source.slug, source.projectId])].join(' ').toLowerCase();
        if (!haystack.includes(f.query)) return false;
      }
      return true;
    });
    const statusRank = {update: 0, modified: 1, review: 2, unknown: 3, current: 4};
    result.sort((a, b) => {
      if (f.sort === 'updates') return (statusRank[a.updateStatus] ?? 9) - (statusRank[b.updateStatus] ?? 9) || text(a.name).localeCompare(text(b.name));
      if (f.sort === 'identity') return (Number(b.provenanceConfidence) || 0) - (Number(a.provenanceConfidence) || 0) || list(b.sources).length - list(a.sources).length || text(a.name).localeCompare(text(b.name));
      if (f.sort === 'recent') return new Date(b.modified).getTime() - new Date(a.modified).getTime();
      if (f.sort === 'size') return Number(b.size || 0) - Number(a.size || 0);
      return text(a.name).localeCompare(text(b.name), undefined, {numeric: true, sensitivity: 'base'});
    });
    return result;
  }

  function statusLabel(item) {
    if (!item.enabled) return 'Disabled';
    return ({update: 'Update ready', current: 'Up to date', modified: 'Protected custom build', review: 'Review match', unknown: 'Resolving identity'}[item.updateStatus] || item.updateStatus || 'Unknown');
  }

  function sourceBadges(item) {
    const sources = list(item.sources);
    if (!sources.length) return `<span class="omni-source-badge unresolved">Local metadata</span>`;
    return sources.slice(0, 5).map(source => `<button class="omni-source-badge ${source.exact ? 'exact' : 'candidate'}" data-manager-source="${esc(source.pageUrl)}" title="${esc(source.evidence || '')}">${source.exact ? '<b>✓</b>' : '<b>?</b>'}${esc(providerLabel(source))}</button>`).join('') + (sources.length > 5 ? `<span class="omni-source-more">+${sources.length - 5}</span>` : '');
  }

  function versionLine(item) {
    const installed = text(item.installedVersion) || 'version unknown';
    const latest = text(item.latestVersion);
    if (!latest || latest === installed) return `<span class="omni-version"><b>${esc(installed)}</b></span>`;
    return `<span class="omni-version"><b>${esc(installed)}</b><i>→</i><strong>${esc(latest)}</strong></span>`;
  }

  function itemCard(item) {
    const selected = state.selected.has(item.id);
    const authors = list(item.authors).join(', ') || list(item.sources).map(source => source.author).filter(Boolean).join(', ') || 'Creator unresolved';
    const target = unique([...list(item.gameVersions), ...list(item.loaders)]).slice(0, 5);
    const canPort = item.edition !== 'bedrock' && /\.jar(?:\.disabled)?$/i.test(item.filename || '');
    const canUpdate = item.updateStatus === 'update' && item.safeUpdate;
    const toggleLabel = item.enabled ? 'Disable' : 'Restore';
    const exactCount = list(item.sources).filter(source => source.exact).length;
    return `<article class="card omni-item ${state.view === 'list' ? 'list' : ''} ${selected ? 'selected' : ''} status-${esc(item.updateStatus || 'unknown')} ${item.enabled ? '' : 'disabled'}" data-item-id="${esc(item.id)}">
      <label class="omni-select"><input type="checkbox" data-manager-select="${esc(item.id)}" ${selected ? 'checked' : ''}><span></span></label>
      <div class="omni-item-art">${itemArt(item)}<span class="omni-edition ${esc(item.edition)}">${esc(humanEdition(item.edition))}</span>${item.localArtUrl ? '<span class="omni-art-origin">Embedded art</span>' : item.remoteArtUrl ? '<span class="omni-art-origin">Provider art</span>' : ''}</div>
      <div class="omni-item-main">
        <div class="omni-item-heading"><div><h3>${esc(item.name || item.filename)}</h3><p>${esc(authors)}</p></div><span class="omni-status ${esc(item.enabled ? item.updateStatus : 'disabled')}"><i></i>${esc(statusLabel(item))}</span></div>
        <div class="omni-source-row">${sourceBadges(item)}</div>
        <p class="omni-summary-text">${esc(item.summary || item.description || item.updateMessage || 'Vault is building a verified identity from the local artifact and every compatible source.')}</p>
        <div class="omni-facts"><span>${esc(humanKind(item.kind))}</span>${versionLine(item)}${target.map(value => `<span>${esc(value)}</span>`).join('')}<span>${fmtBytes(item.size)}</span></div>
        <div class="omni-file-line" title="${esc(item.path)}"><span>${icon('folder')}</span><code>${esc(item.filename)}</code><em>${exactCount ? `${exactCount} exact source${exactCount === 1 ? '' : 's'}` : `${confidence(item.provenanceConfidence)} identity confidence`}</em></div>
        <div class="omni-item-actions"><button class="btn small primary" data-manager-details="${esc(item.id)}">Details</button>${canUpdate ? `<button class="btn small success" data-manager-action="update" data-item-id="${esc(item.id)}">${icon('download')} Update</button>` : ''}<button class="btn small" data-manager-action="toggle" data-item-id="${esc(item.id)}">${icon('power')} ${toggleLabel}</button>${canPort ? `<button class="btn small" data-manager-port="${esc(item.id)}">${icon('compass')} Port</button>` : ''}<button class="btn small" data-manager-open="${esc(item.id)}">${icon('folder')} Folder</button><button class="btn small danger" data-manager-action="trash" data-item-id="${esc(item.id)}">${icon('trash')}</button></div>
      </div>
    </article>`;
  }

  function renderItems() {
    const visible = filteredItems();
    const root = $('managerGroups');
    if (!root) return;
    root.classList.toggle('list-view', state.view === 'list');
    root.innerHTML = visible.length ? visible.map(itemCard).join('') : `<div class="card omni-empty"><div>${icon('search')}</div><h3>No managed content matches these filters.</h3><p>Clear a filter or resolve provider identities again. Nothing is hidden by pagination or viewport culling.</p><button class="btn" data-manager-reset>Reset filters</button></div>`;
    root.querySelectorAll('.omni-art-image').forEach(image => image.addEventListener('error', () => {
      const fallback = document.createElement('div');
      fallback.className = 'omni-art-fallback';
      fallback.innerHTML = `${icon('mod')}<span>${esc((image.alt || '?').slice(0, 1).toUpperCase())}</span>`;
      image.replaceWith(fallback);
    }, {once: true}));
    $('managerResultCount').textContent = `${visible.length.toLocaleString()} of ${items().length.toLocaleString()} items shown in full`;
    const selectAll = $('managerSelectAll');
    if (selectAll) {
      selectAll.checked = visible.length > 0 && visible.every(item => state.selected.has(item.id));
      selectAll.indeterminate = visible.some(item => state.selected.has(item.id)) && !selectAll.checked;
    }
  }

  function renderBulk() {
    const root = $('managerBulk');
    if (!root) return;
    root.hidden = state.selected.size === 0;
    $('managerSelectedCount').textContent = `${state.selected.size.toLocaleString()} selected`;
  }

  function resetFilters() {
    if ($('managerQuery')) $('managerQuery').value = '';
    for (const id of ['managerProfile', 'managerEdition', 'managerKind', 'managerStatus', 'managerSource']) if ($(id)) $(id).value = 'all';
    if ($('managerSort')) $('managerSort').value = 'name';
    renderItems();
  }

  async function performAction(action, ids) {
    ids = unique(ids).filter(Boolean);
    if (!ids.length) return;
    const actionName = ({update: 'update', toggle: 'toggle', trash: 'move to Vault trash'}[action] || action);
    if (action === 'trash' && !confirm(`Move ${ids.length} selected item${ids.length === 1 ? '' : 's'} to recoverable Vault trash?`)) return;
    try {
      showToast(`Starting ${actionName}…`);
      const response = await api('/api/library/action', {method: 'POST', body: JSON.stringify({action, ids})});
      const failures = list(response.results).filter(result => !result.ok);
      if (failures.length) throw new Error(failures.map(failure => failure.error).join('; '));
      state.selected.clear();
      showToast(`${ids.length} item${ids.length === 1 ? '' : 's'} ${action === 'update' ? 'updated' : action === 'trash' ? 'moved to recoverable Vault trash' : 'toggled'} successfully.`);
      await refresh({force: action === 'update'});
    } catch (error) {
      showToast(`${actionName} failed: ${error.message}`, true);
    }
  }

  async function toggleItems(ids) {
    const restore = ids.map(itemByID).filter(item => item && !item.enabled && item.receiptId);
    const ordinary = ids.map(itemByID).filter(item => item && !restore.includes(item)).map(item => item.id);
    try {
      for (const item of restore) await undo(item.receiptId, false);
      if (ordinary.length) await performAction('toggle', ordinary);
      else if (restore.length) {
        state.selected.clear();
        showToast(`${restore.length} disabled item${restore.length === 1 ? '' : 's'} restored from receipt.`);
        await refresh();
      }
    } catch (error) {
      showToast(`Toggle failed: ${error.message}`, true);
    }
  }

  function openDrawer(title, eyebrow, body) {
    $('managerDrawerTitle').textContent = title;
    $('managerDrawerEyebrow').textContent = eyebrow;
    $('managerDrawerBody').innerHTML = body;
    const drawer = $('managerDrawer');
    drawer.classList.add('open');
    drawer.setAttribute('aria-hidden', 'false');
  }

  function closeDrawer() {
    const drawer = $('managerDrawer');
    drawer?.classList.remove('open');
    drawer?.setAttribute('aria-hidden', 'true');
  }

  function metadataRows(item) {
    const rows = [
      ['Edition', humanEdition(item.edition)], ['Type', humanKind(item.kind)], ['Profile', item.profile], ['Installed', item.installedVersion], ['Latest compatible', item.latestVersion],
      ['Mod ID', item.modId], ['Bedrock UUID', item.uuid], ['Minimum engine', item.minEngineVersion], ['Loaders', list(item.loaders).join(', ')], ['Game versions', list(item.gameVersions).join(', ')],
      ['Modules', list(item.modules).join(', ')], ['Capabilities', list(item.capabilities).join(', ')], ['License', item.license], ['Metadata source', item.metadataBy], ['Size', fmtBytes(item.size)], ['Modified', fmtDate(item.modified)],
    ].filter(([, value]) => text(value));
    return rows.map(([label, value]) => `<div><span>${esc(label)}</span><strong>${esc(value)}</strong></div>`).join('');
  }

  function detailsBody(item) {
    const sources = list(item.sources);
    const dependencies = list(item.dependencies);
    const evidence = list(item.matchEvidence);
    const warnings = list(item.warnings);
    const worlds = items().filter(candidate => candidate.edition === 'bedrock' && candidate.kind === 'world-bedrock' && candidate.enabled);
    const canActivate = item.edition === 'bedrock' && item.uuid && !item.kind.includes('world') && worlds.length;
    const sourceCards = sources.length ? sources.map(source => `<article class="omni-source-card ${source.exact ? 'exact' : 'candidate'}"><div><strong>${source.exact ? '✓ Exact' : 'Review'} · ${esc(providerLabel(source))}</strong><span>${confidence(source.confidence)}</span></div><h4>${esc(source.title || item.name)}</h4><p>${esc(source.evidence || 'Provider match')}</p><div>${source.latestVersion ? `<span>Latest ${esc(source.latestVersion)}</span>` : ''}${source.author ? `<span>By ${esc(source.author)}</span>` : ''}</div>${source.pageUrl ? `<button class="btn small" data-manager-source="${esc(source.pageUrl)}">${icon('external')} Open exact project</button>` : ''}</article>`).join('') : '<div class="omni-detail-empty">No provider has claimed this exact artifact yet. Embedded metadata remains visible and fuzzy matches are never allowed to overwrite it automatically.</div>';
    const canPort = item.edition !== 'bedrock' && /\.jar(?:\.disabled)?$/i.test(item.filename || '');
    return `<div class="omni-detail-hero"><div class="omni-detail-art">${itemArt(item, true)}</div><div><div class="omni-detail-title-row"><span class="omni-status ${esc(item.enabled ? item.updateStatus : 'disabled')}"><i></i>${esc(statusLabel(item))}</span><span>${confidence(item.provenanceConfidence)} identity confidence</span></div><h3>${esc(item.name || item.filename)}</h3><p>${esc(item.summary || item.description || item.updateMessage || '')}</p><div class="omni-source-row">${sourceBadges(item)}</div></div></div>
      <div class="omni-detail-actions">${item.safeUpdate ? `<button class="btn success" data-manager-action="update" data-item-id="${esc(item.id)}">${icon('download')} Install verified update</button>` : ''}<button class="btn" data-manager-action="toggle" data-item-id="${esc(item.id)}">${icon('power')} ${item.enabled ? 'Disable' : 'Restore'}</button>${canPort ? `<button class="btn" data-manager-port="${esc(item.id)}">${icon('compass')} Open in Porting Lab</button>` : ''}<button class="btn" data-manager-open="${esc(item.id)}">${icon('folder')} Open containing folder</button><button class="btn danger" data-manager-action="trash" data-item-id="${esc(item.id)}">${icon('trash')} Vault trash</button></div>
      <section class="omni-detail-section"><div class="omni-detail-section-head"><span class="eyebrow">LOCAL TRUTH</span><h3>Artifact identity</h3></div><div class="omni-metadata-grid">${metadataRows(item)}</div><div class="omni-path-block"><span>Actual file</span><code>${esc(item.path)}</code></div>${item.hashes ? `<div class="omni-hash-list">${Object.entries(item.hashes).filter(([, value]) => value).map(([key, value]) => `<div><span>${esc(key)}</span><code>${esc(value)}</code></div>`).join('')}</div>` : ''}</section>
      <section class="omni-detail-section"><div class="omni-detail-section-head"><span class="eyebrow">CROSS-STORE GRAPH</span><h3>${sources.length || 0} matching source${sources.length === 1 ? '' : 's'}</h3></div><div class="omni-source-cards">${sourceCards}</div></section>
      ${dependencies.length ? `<section class="omni-detail-section"><div class="omni-detail-section-head"><span class="eyebrow">DEPENDENCIES</span><h3>${dependencies.length} declared relationship${dependencies.length === 1 ? '' : 's'}</h3></div><div class="omni-chip-cloud">${dependencies.map(dep => `<span><b>${esc(dep.id)}</b>${dep.version ? ` ${esc(dep.version)}` : ''}${dep.required ? ' · required' : ''}</span>`).join('')}</div></section>` : ''}
      ${canActivate ? `<section class="omni-detail-section"><div class="omni-detail-section-head"><span class="eyebrow">BEDROCK WORLD</span><h3>Activate this pack transactionally</h3></div><div class="omni-activation"><select id="managerActivationWorld">${worlds.map(world => `<option value="${esc(world.path)}">${esc(world.name)} · ${esc(world.profile)}</option>`).join('')}</select><button class="btn primary" data-manager-activate="${esc(item.id)}">${icon('bedrock')} Activate in world</button></div><p class="omni-detail-note">Vault edits the correct world behavior/resource-pack ledger and records the exact previous bytes for one-click undo.</p></section>` : ''}
      ${(evidence.length || warnings.length) ? `<section class="omni-detail-section"><div class="omni-detail-section-head"><span class="eyebrow">PROOF</span><h3>Why Vault believes this identity</h3></div>${evidence.length ? `<ul class="omni-evidence">${evidence.map(entry => `<li>${esc(entry)}</li>`).join('')}</ul>` : ''}${warnings.length ? `<div class="omni-warnings">${warnings.map(entry => `<p>${esc(entry)}</p>`).join('')}</div>` : ''}</section>` : ''}`;
  }

  function openDetails(id) {
    const item = itemByID(id);
    if (!item) return;
    openDrawer(item.name || item.filename, `${humanEdition(item.edition)} · ${humanKind(item.kind)}`, detailsBody(item));
  }

  async function showHistory() {
    openDrawer('Recovery history', 'TRANSACTION LEDGER', '<div class="omni-drawer-loading"><span class="omni-pulse"></span> Loading receipts…</div>');
    try {
      const response = await api('/api/library/history');
      const transactions = list(response.transactions);
      $('managerDrawerBody').innerHTML = `<div class="omni-history-intro"><strong>Nothing destructive is silent.</strong><p>Disables, trash moves, Bedrock installs, and world activations are recorded with exact source and destination paths.</p></div>${transactions.length ? `<div class="omni-history-list">${transactions.map(transaction => `<article class="card omni-history-item ${transaction.undoneAt ? 'undone' : ''}"><div><span class="omni-history-action">${esc(transaction.action)}</span><time>${esc(fmtDate(transaction.createdAt))}</time></div><h3>${esc(list(transaction.itemNames).join(', ') || list(transaction.sourcePaths).map(path => path.split(/[\\/]/).pop()).join(', ') || transaction.id)}</h3><p>${esc(list(transaction.sourcePaths).join(' · '))}</p><div><code>${esc(transaction.id)}</code>${transaction.undoneAt ? `<span>Undone ${esc(fmtDate(transaction.undoneAt))}</span>` : `<button class="btn small" data-manager-undo="${esc(transaction.id)}">Undo safely</button>`}</div></article>`).join('')}</div>` : '<div class="omni-detail-empty">No OmniManager transactions yet.</div>'}`;
    } catch (error) {
      $('managerDrawerBody').innerHTML = `<div class="omni-detail-empty error">${esc(error.message)}</div>`;
    }
  }

  async function undo(receiptID, refreshAfter = true) {
    const response = await api('/api/library/undo', {method: 'POST', body: JSON.stringify({receiptId: receiptID})});
    if (refreshAfter) {
      showToast('The transaction was restored from its exact receipt.');
      closeDrawer();
      await refresh();
    }
    return response;
  }

  function bedrockInstallerBody() {
    const profiles = list(state.payload?.profiles).filter(profile => profile.edition === 'bedrock');
    return `<div class="omni-install-intro"><strong>Native Bedrock package management</strong><p>Install .mcpack, .mcaddon, .mcworld, and .mctemplate files into Stable, Preview/Beta, or any configured custom com.mojang root. Every original is hash-preserved and every install has an undo receipt.</p></div>
      <label class="omni-install-profile">Destination profile<select id="managerBedrockProfile">${profiles.map(profile => `<option value="${esc(profile.id)}">${esc(profile.name)} · ${esc(profile.channel || 'Bedrock')}${profile.exists ? '' : ' · folder will be created'}</option>`).join('')}</select></label>
      <div class="omni-bedrock-drop" id="managerBedrockDrop"><div>${icon('bedrock')}</div><h3>Drop Bedrock packages here</h3><p>.mcpack · .mcaddon · .mcworld · .mctemplate</p><button class="btn primary" id="managerBedrockChoose">Choose packages</button><input id="managerBedrockInput" type="file" accept=".mcpack,.mcaddon,.mcworld,.mctemplate" multiple hidden></div>
      <div class="omni-install-results" id="managerBedrockResults"></div>`;
  }

  function openBedrockInstaller() {
    openDrawer('Install Bedrock content', 'BEDROCK PACKAGE LAB', bedrockInstallerBody());
    const input = $('managerBedrockInput');
    const drop = $('managerBedrockDrop');
    $('managerBedrockChoose').onclick = () => input.click();
    input.onchange = () => installBedrockFiles(input.files);
    for (const event of ['dragenter', 'dragover']) drop.addEventListener(event, e => { e.preventDefault(); drop.classList.add('drag'); });
    for (const event of ['dragleave', 'drop']) drop.addEventListener(event, e => { e.preventDefault(); drop.classList.remove('drag'); });
    drop.addEventListener('drop', event => installBedrockFiles(event.dataTransfer.files));
  }

  async function installBedrockFiles(fileList) {
    const files = [...(fileList || [])];
    if (!files.length) return;
    const profile = $('managerBedrockProfile')?.value;
    const results = $('managerBedrockResults');
    results.innerHTML = `<div class="omni-drawer-loading"><span class="omni-pulse"></span> Validating manifests, preserving originals, and installing ${files.length} package${files.length === 1 ? '' : 's'}…</div>`;
    const body = new FormData();
    body.append('profile', profile || '');
    for (const file of files) body.append('files', file, file.name);
    try {
      const response = await api('/api/library/bedrock/install', {method: 'POST', body});
      const rows = list(response.results);
      results.innerHTML = rows.map(row => row.ok ? `<article class="omni-install-result ok"><strong>✓ ${esc(row.package)}</strong><span>${esc(row.kinds?.join(', ') || 'Installed')}</span><code>${esc(row.receiptId)}</code></article>` : `<article class="omni-install-result error"><strong>! ${esc(row.package || 'Package')}</strong><span>${esc(row.error)}</span></article>`).join('');
      const success = rows.filter(row => row.ok).length;
      showToast(`${success} Bedrock package${success === 1 ? '' : 's'} installed with recovery receipts.`);
      await refresh();
    } catch (error) {
      results.innerHTML = `<article class="omni-install-result error"><strong>Installation failed</strong><span>${esc(error.message)}</span></article>`;
      showToast(`Bedrock install failed: ${error.message}`, true);
    }
  }

  async function activatePack(id) {
    const item = itemByID(id);
    const worldPath = $('managerActivationWorld')?.value;
    if (!item || !worldPath) return;
    try {
      const response = await api('/api/library/bedrock/activate', {method: 'POST', body: JSON.stringify({worldPath, packUuid: item.uuid, version: item.installedVersion || '1.0.0', packKind: item.kind})});
      showToast(`Activated ${item.name} with receipt ${response.receiptId}.`);
      await showHistory();
    } catch (error) {
      showToast(`Bedrock activation failed: ${error.message}`, true);
    }
  }

  function handleLibraryClick(event) {
    const reset = event.target.closest('[data-manager-reset]');
    if (reset) return resetFilters();
    const details = event.target.closest('[data-manager-details]');
    if (details) return openDetails(details.dataset.managerDetails);
    const source = event.target.closest('[data-manager-source]');
    if (source?.dataset.managerSource) return openExternal(source.dataset.managerSource);
    const open = event.target.closest('[data-manager-open]');
    if (open) {
      const item = itemByID(open.dataset.managerOpen);
      if (item) openExternal(item.managedRoot || item.path, true);
      return;
    }
    const port = event.target.closest('[data-manager-port]');
    if (port) {
      const item = itemByID(port.dataset.managerPort);
      if (item && typeof selectPortingFile === 'function') selectPortingFile(item.path);
      return;
    }
    const action = event.target.closest('[data-manager-action]');
    if (action) {
      const id = action.dataset.itemId;
      if (action.dataset.managerAction === 'toggle') return toggleItems([id]);
      return performAction(action.dataset.managerAction, [id]);
    }
    const undoButton = event.target.closest('[data-manager-undo]');
    if (undoButton) return undo(undoButton.dataset.managerUndo);
    const activate = event.target.closest('[data-manager-activate]');
    if (activate) return activatePack(activate.dataset.managerActivate);
  }

  function setView(view) {
    state.view = view === 'list' ? 'list' : 'grid';
    try { localStorage.setItem('mmv-manager-view', state.view); } catch {}
    render();
  }

  function applySummaryFilter(filter) {
    resetFilters();
    if (filter === 'update' || filter === 'current' || filter === 'modified' || filter === 'disabled') $('managerStatus').value = filter;
    else if (filter === 'bedrock') $('managerEdition').value = 'bedrock';
    else if (filter === 'identity') $('managerSort').value = 'identity';
    else if (filter === 'art') $('managerQuery').value = '';
    renderItems();
  }

  function wire() {
    if (state.wired) return;
    state.wired = true;
    $('resolveManager').onclick = () => refresh({force: true});
    $('updateManagerSafe').onclick = () => {
      const ids = items().filter(item => item.updateStatus === 'update' && item.safeUpdate).map(item => item.id);
      if (!ids.length) return showToast('No exact safe updates are currently ready.', true);
      performAction('update', ids);
    };
    $('installBedrockManager').onclick = openBedrockInstaller;
    $('managerHistory').onclick = showHistory;
    $('managerGridView').onclick = () => setView('grid');
    $('managerListView').onclick = () => setView('list');
    $('managerQuery').addEventListener('input', renderItems);
    for (const id of ['managerProfile', 'managerEdition', 'managerKind', 'managerStatus', 'managerSource', 'managerSort']) $(id).addEventListener('change', renderItems);
    $('managerGroups').addEventListener('click', handleLibraryClick);
    $('managerGroups').addEventListener('change', event => {
      const checkbox = event.target.closest('[data-manager-select]');
      if (!checkbox) return;
      checkbox.checked ? state.selected.add(checkbox.dataset.managerSelect) : state.selected.delete(checkbox.dataset.managerSelect);
      renderItems(); renderBulk();
    });
    $('managerSummary').addEventListener('click', event => {
      const card = event.target.closest('[data-summary-filter]');
      if (card) applySummaryFilter(card.dataset.summaryFilter);
    });
    $('managerSelectAll').addEventListener('change', event => {
      for (const item of filteredItems()) event.target.checked ? state.selected.add(item.id) : state.selected.delete(item.id);
      renderItems(); renderBulk();
    });
    $('managerClearSelection').onclick = () => { state.selected.clear(); renderItems(); renderBulk(); };
    $('managerBulkUpdate').onclick = () => performAction('update', [...state.selected].filter(id => itemByID(id)?.safeUpdate));
    $('managerBulkToggle').onclick = () => toggleItems([...state.selected]);
    $('managerBulkTrash').onclick = () => performAction('trash', [...state.selected]);
    $('managerDrawerClose').onclick = closeDrawer;
    document.querySelectorAll('[data-manager-close]').forEach(element => element.onclick = closeDrawer);
    $('managerDrawerBody').addEventListener('click', handleLibraryClick);
    document.addEventListener('keydown', event => { if (event.key === 'Escape') closeDrawer(); });
  }

  return {wire, refresh, render, closeDrawer, openBedrockInstaller, showHistory};
})();
