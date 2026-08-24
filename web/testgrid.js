'use strict';

(() => {
  const presets = {
    self: {
      schemaVersion: 1,
      name: 'Vault TestGrid self-test',
      description: 'Proves process, log assertion, file assertion, hashing, capture, and reports on Windows.',
      edition: 'generic',
      timeoutSeconds: 30,
      runtime: {
        kind: 'build-or-tool',
        executable: 'powershell.exe',
        arguments: ['-NoProfile', '-NonInteractive', '-Command', "Write-Output 'MMV_TESTGRID_READY'; Set-Content -LiteralPath 'testgrid-evidence.txt' -Value 'deterministic evidence' -Encoding utf8"],
        workingDirectory: '${TEMP}',
      },
      steps: [
        { name: 'runtime ready', type: 'wait-log', pattern: 'MMV_TESTGRID_READY', timeoutSeconds: 10 },
        { name: 'no PowerShell failure', type: 'deny-log', pattern: '(?i)(exception|fullyqualifiederrorid)', timeoutSeconds: 5 },
        { name: 'evidence created', type: 'file-exists', path: 'testgrid-evidence.txt', timeoutSeconds: 5 },
      ],
      artifacts: [{ name: 'self-test-evidence.txt', path: 'testgrid-evidence.txt' }],
    },
    java: {
      schemaVersion: 1,
      name: 'Java dedicated server verification',
      edition: 'java',
      gameVersion: '1.20.1',
      loader: 'forge',
      timeoutSeconds: 300,
      runtime: {
        kind: 'java-server',
        javaPath: 'java',
        jar: 'server.jar',
        jvmArguments: ['-Xms1G', '-Xmx2G'],
        arguments: ['nogui'],
        workingDirectory: 'C:\\Minecraft\\TestServer',
        stop: { rconAddress: '127.0.0.1:25575', passwordEnv: 'MMV_RCON_PASSWORD', command: 'stop', graceSeconds: 20 },
      },
      steps: [
        { name: 'server startup', type: 'wait-log', pattern: 'Done \\([0-9.]+s\\)!', timeoutSeconds: 180 },
        { name: 'Java status protocol', type: 'java-ping', address: '127.0.0.1:25565', timeoutSeconds: 30 },
        { name: 'GameTest summary', type: 'assert-log', pattern: '(?i)(all required tests passed|0 required tests failed)', timeoutSeconds: 5 },
        { name: 'no fatal crash', type: 'deny-log', pattern: '(?i)(crash report|exception in server tick loop)', timeoutSeconds: 5 },
      ],
      artifacts: [
        { name: 'latest.log', path: 'logs/latest.log' },
        { name: 'debug.log', path: 'logs/debug.log', optional: true },
        { name: 'server.properties', path: 'server.properties' },
      ],
    },
    bedrock: {
      schemaVersion: 1,
      name: 'Bedrock Dedicated Server verification',
      edition: 'bedrock',
      gameVersion: 'current-installed',
      timeoutSeconds: 180,
      runtime: {
        kind: 'bedrock-server',
        executable: 'bedrock_server.exe',
        workingDirectory: 'C:\\Minecraft\\BedrockServer',
      },
      steps: [
        { name: 'server startup', type: 'wait-log', pattern: '(?i)(server started|IPv4 supported)', timeoutSeconds: 120 },
        { name: 'Bedrock RakNet status', type: 'bedrock-ping', address: '127.0.0.1:19132', timeoutSeconds: 30 },
        { name: 'no fatal crash', type: 'deny-log', pattern: '(?i)(fatal|unhandled exception)', timeoutSeconds: 5 },
      ],
      artifacts: [
        { name: 'server.properties', path: 'server.properties' },
        { name: 'worlds', path: 'worlds', optional: true },
      ],
    },
  };

  let wired = false;
  let loaded = false;
  let selectedRun = '';
  let pollTimer = 0;

  function setPreset(name) {
    const editor = document.getElementById('testgridManifest');
    if (editor && presets[name]) editor.value = JSON.stringify(presets[name], null, 2);
    setEditorStatus('Manifest loaded. Validate it before starting the runtime.');
  }

  function setEditorStatus(message, failed = false) {
    const status = document.getElementById('testgridEditorStatus');
    if (!status) return;
    status.textContent = message;
    status.classList.toggle('failed', failed);
  }

  function readManifest() {
    const value = document.getElementById('testgridManifest')?.value || '';
    try {
      return JSON.parse(value);
    } catch (error) {
      throw new Error(`Manifest JSON: ${error.message}`);
    }
  }

  async function loadCapabilities() {
    const root = document.getElementById('testgridCapabilities');
    try {
      const capabilities = await api('/api/testgrid/capabilities');
      document.getElementById('testgridHost').textContent = `${capabilities.host.os} ${capabilities.host.arch} · schema ${capabilities.schemaVersion}`;
      root.innerHTML = (capabilities.runtimes || []).map(runtime => `<article class="card testgrid-capability"><div><span>${icon(runtime.headless === true ? 'shield' : 'gauge')}</span><strong>${esc(runtime.id)}</strong></div><p>${esc(runtime.versions)}</p><small>Headless: ${esc(runtime.headless)}</small><div>${(runtime.evidence || []).map(item => `<em>${esc(item)}</em>`).join('')}</div></article>`).join('');
    } catch (error) {
      root.innerHTML = `<div class="card empty">Capability matrix unavailable: ${esc(error.message)}</div>`;
    }
  }

  async function validateManifest() {
    try {
      const manifest = readManifest();
      const result = await api('/api/testgrid/validate', { method: 'POST', body: JSON.stringify(manifest) });
      setEditorStatus(`Valid schema ${result.schemaVersion}. Ready to run.`);
      showToast('TestGrid manifest is valid.');
      return true;
    } catch (error) {
      setEditorStatus(error.message, true);
      showToast(error.message, true);
      return false;
    }
  }

  async function startRun() {
    try {
      const manifest = readManifest();
      setEditorStatus('Starting runtime…');
      const run = await api('/api/testgrid/run', { method: 'POST', body: JSON.stringify(manifest) });
      selectedRun = run.id;
      setEditorStatus(`Started ${run.id}. Evidence is streaming to disk.`);
      await loadRuns();
      await selectRun(run.id);
      ensurePolling();
    } catch (error) {
      setEditorStatus(error.message, true);
      showToast(error.message, true);
    }
  }

  async function loadRuns() {
    const root = document.getElementById('testgridRunList');
    if (!root) return;
    try {
      const data = await api('/api/testgrid/runs');
      const runs = data.runs || [];
      document.getElementById('testgridRunCount').textContent = String(runs.length);
      root.innerHTML = runs.length ? runs.map(run => `<button class="testgrid-run ${run.id === selectedRun ? 'active' : ''}" data-testgrid-run="${esc(run.id)}"><i class="state-${esc(run.status)}"></i><span><strong>${esc(run.name)}</strong><small>${esc(run.id)} · ${esc(run.edition)} ${esc(run.gameVersion || '')}</small></span><em>${esc(run.status)}</em></button>`).join('') : '<div class="empty">No TestGrid runs yet.</div>';
      root.querySelectorAll('[data-testgrid-run]').forEach(button => button.onclick = () => selectRun(button.dataset.testgridRun));
      if (runs.some(run => run.status === 'queued' || run.status === 'running')) ensurePolling();
    } catch (error) {
      root.innerHTML = `<div class="empty">Run history unavailable: ${esc(error.message)}</div>`;
    }
  }

  function fileURL(run, kind, name = '') {
    const query = new URLSearchParams({ token: TOKEN, id: run.id, kind });
    if (name) query.set('name', name);
    return `/api/testgrid/file?${query}`;
  }

  async function selectRun(id) {
    selectedRun = id;
    const root = document.getElementById('testgridReport');
    try {
      const run = await api(`/api/testgrid/runs?id=${encodeURIComponent(id)}`);
      const terminal = ['passed', 'failed', 'canceled'].includes(run.status);
      root.innerHTML = `<div class="testgrid-report-head"><div><span class="testgrid-state state-${esc(run.status)}">${esc(run.status)}</span><h3>${esc(run.name)}</h3><p>${esc(run.id)} · ${esc(run.edition)} ${esc(run.gameVersion || '')} ${esc(run.loader || '')}</p></div><div class="testgrid-report-actions">${!terminal ? `<button class="btn danger" data-testgrid-cancel="${esc(run.id)}">${icon('pause')} Cancel</button>` : ''}<button class="btn" data-testgrid-open>${icon('folder')} Open run folder</button>${terminal ? `<a class="btn" href="${fileURL(run, 'log')}">${icon('download')} Log</a><a class="btn" href="${fileURL(run, 'report')}">${icon('download')} JSON</a><a class="btn" href="${fileURL(run, 'junit')}">${icon('download')} JUnit</a><a class="btn" href="${fileURL(run, 'html')}">${icon('download')} HTML</a>` : ''}</div></div>
        ${run.error ? `<div class="testgrid-error">${esc(run.error)}</div>` : ''}
        <div class="testgrid-process"><div><small>PID</small><strong>${esc(run.process?.pid || '—')}</strong></div><div><small>Exit</small><strong>${esc(run.process?.exitCode ?? '—')}</strong></div><div><small>Duration</small><strong>${esc(run.durationMs || 0)} ms</strong></div><div><small>Stopped by</small><strong>${esc(run.process?.stoppedBy || 'natural')}</strong></div><div><small>CPU</small><strong>${esc(run.process?.userCpu || '—')} user</strong></div></div>
        <div class="testgrid-step-list">${(run.steps || []).map(step => `<article><i class="state-${esc(step.status)}"></i><div><strong>${esc(step.name)}</strong><small>${esc(step.type)} · ${esc(step.durationMs)} ms</small><p>${esc(step.message || '')}</p></div><em>${esc(step.status)}</em></article>`).join('') || '<div class="empty">Steps have not reported yet.</div>'}</div>
        ${(run.artifacts || []).length ? `<div class="testgrid-artifacts"><h4>Captured artifacts</h4>${run.artifacts.map(artifact => `<a href="${fileURL(run, 'artifact', artifact.name)}"><span>${icon('download')}<b>${esc(artifact.name)}</b></span><small>${esc(fmtBytes(artifact.size))} · ${esc(artifact.sha256)}</small></a>`).join('')}</div>` : ''}`;
      root.querySelector('[data-testgrid-open]')?.addEventListener('click', () => openExternal(run.runDirectory, true));
      root.querySelector('[data-testgrid-cancel]')?.addEventListener('click', () => cancelRun(run.id));
      await loadRuns();
      if (!terminal) ensurePolling();
    } catch (error) {
      root.innerHTML = `<div class="empty">Run unavailable: ${esc(error.message)}</div>`;
    }
  }

  async function cancelRun(id) {
    try {
      await api('/api/testgrid/cancel', { method: 'POST', body: JSON.stringify({ id }) });
      showToast('Cancellation requested.');
      ensurePolling();
    } catch (error) {
      showToast(error.message, true);
    }
  }

  function ensurePolling() {
    if (pollTimer) return;
    pollTimer = window.setInterval(async () => {
      if (!document.querySelector('[data-view="testgrid"]')?.classList.contains('active')) return;
      await loadRuns();
      if (selectedRun) await selectRun(selectedRun);
    }, 1800);
  }

  function wire() {
    if (wired || !document.getElementById('testgridManifest')) return;
    wired = true;
    document.getElementById('testgridPreset').onchange = event => setPreset(event.target.value);
    document.getElementById('testgridValidate').onclick = validateManifest;
    document.getElementById('testgridRun').onclick = startRun;
    document.getElementById('testgridRefresh').onclick = loadRuns;
    setPreset('self');
  }

  async function activate() {
    wire();
    if (!loaded) {
      loaded = true;
      await Promise.all([loadCapabilities(), loadRuns()]);
    } else {
      await loadRuns();
    }
  }

  window.TestGridStudio = { wire, activate };
  window.setTimeout(wire, 0);
})();
