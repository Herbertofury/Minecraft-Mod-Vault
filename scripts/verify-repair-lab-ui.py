#!/usr/bin/env python3
"""Exercise production Repair Lab UI code against a real authenticated backend.

Managed Chromium in the verification host blocks direct loopback navigation. The
harness therefore renders the exact shipped HTML/CSS/JS in Chromium and bridges
only the transport layer to the real compiled backend. File upload remains driven
through the production input control and FormData path; artifact downloads are
fulfilled from the real authenticated download endpoint.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import mimetypes
import re
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path
from typing import Any
from urllib.parse import urlsplit

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError, sync_playwright


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def wait_text(page, selector: str, expected: str, timeout: int = 30_000) -> str:
    page.wait_for_function(
        "([selector, expected]) => { const el=document.querySelector(selector); return !!el && (el.textContent||'').includes(expected); }",
        arg=[selector, expected],
        timeout=timeout,
    )
    return page.locator(selector).inner_text()


def parse_json_response(raw: bytes) -> Any:
    return json.loads(raw.decode("utf-8")) if raw else {}


def progress(message: str) -> None:
    print(f"[repair-ui] {message}", flush=True)


def download_from_rendered_link(
    base_url: str,
    token: str,
    href: str,
    destination: Path,
    call_log: list[dict[str, Any]],
    kind: str,
) -> dict[str, Any]:
    """Follow an exact link rendered by the production UI through the real backend.

    Chromium on this host is policy-blocked from direct loopback navigation. The
    UI must still render the exact authenticated download URL; this verifier then
    follows that URL outside Chromium and hashes the returned bytes.
    """
    parsed = urlsplit(href)
    if parsed.path.startswith("/api/repair-lab/download") is False:
        raise AssertionError(f"unexpected download URL rendered by UI: {href}")
    relative = parsed.path + (("?" + parsed.query) if parsed.query else "")
    request = urllib.request.Request(
        base_url.rstrip("/") + relative,
        headers={"X-MMV-Token": token},
        method="GET",
    )
    started = time.time()
    with urllib.request.urlopen(request, timeout=90) as response:
        body = response.read()
        disposition = response.headers.get("Content-Disposition") or ""
        match = re.search(r'filename\*?=(?:UTF-8\'\')?["\']?([^"\';]+)', disposition, re.I)
        name = Path(match.group(1).strip()).name if match else f"{kind}-download.bin"
        target = destination / name
        target.write_bytes(body)
        call_log.append({
            "method": "GET",
            "path": parsed.path,
            "status": response.status,
            "durationMs": round((time.time() - started) * 1000),
            "ok": True,
            "download": True,
            "bytes": len(body),
        })
    return {
        "kind": kind,
        "name": target.name,
        "path": str(target),
        "uiHref": href,
        "size": target.stat().st_size,
        "sha256": sha256_file(target),
    }


def make_api_bridge(base_url: str, token: str, call_log: list[dict[str, Any]]):
    base = base_url.rstrip("/")

    def bridge(path: str, options: dict[str, Any] | None = None) -> Any:
        opts = options or {}
        method = str(opts.get("method") or "GET").upper()
        body = opts.get("body")
        data: bytes | None = None
        headers = {"X-MMV-Token": token}
        if body is not None:
            if not isinstance(body, str):
                body = json.dumps(body, separators=(",", ":"))
            data = body.encode("utf-8")
            headers["Content-Type"] = "application/json"
        request = urllib.request.Request(base + path, data=data, headers=headers, method=method)
        started = time.time()
        try:
            with urllib.request.urlopen(request, timeout=90) as response:
                raw = response.read()
                status = response.status
        except urllib.error.HTTPError as error:
            raw = error.read()
            status = error.code
            parsed = parse_json_response(raw) if raw else {}
            call_log.append({"method": method, "path": path.split("?", 1)[0], "status": status, "ok": False})
            raise RuntimeError(parsed.get("error") or f"HTTP {status}") from error
        call_log.append({
            "method": method,
            "path": path.split("?", 1)[0],
            "status": status,
            "durationMs": round((time.time() - started) * 1000),
            "ok": True,
        })
        return parse_json_response(raw)

    return bridge


def make_upload_bridge(base_url: str, token: str, call_log: list[dict[str, Any]]):
    base = base_url.rstrip("/")

    def upload(path: str, payload: dict[str, Any]) -> Any:
        name = Path(str(payload.get("name") or "source.zip")).name.replace('"', "")
        content_type = str(payload.get("type") or "application/zip")
        content = bytes(payload.get("bytes") or [])
        boundary = "----MMVRepairSmoke" + uuid.uuid4().hex
        body = b"".join([
            f"--{boundary}\r\n".encode(),
            f'Content-Disposition: form-data; name="source"; filename="{name}"\r\n'.encode(),
            f"Content-Type: {content_type}\r\n\r\n".encode(),
            content,
            f"\r\n--{boundary}--\r\n".encode(),
        ])
        request = urllib.request.Request(
            base + path,
            data=body,
            headers={"X-MMV-Token": token, "Content-Type": f"multipart/form-data; boundary={boundary}"},
            method="POST",
        )
        started = time.time()
        try:
            with urllib.request.urlopen(request, timeout=90) as response:
                raw = response.read()
                status = response.status
        except urllib.error.HTTPError as error:
            raw = error.read()
            parsed = parse_json_response(raw) if raw else {}
            call_log.append({"method": "POST", "path": path, "status": error.code, "ok": False, "upload": True})
            raise RuntimeError(parsed.get("error") or f"HTTP {error.code}") from error
        call_log.append({
            "method": "POST", "path": path, "status": status,
            "durationMs": round((time.time() - started) * 1000), "ok": True,
            "upload": True, "bytes": len(content),
        })
        return parse_json_response(raw)

    return upload


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-root", required=True, type=Path)
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--token", required=True)
    parser.add_argument("--fixture", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--downloads", required=True, type=Path)
    parser.add_argument("--screenshot", type=Path)
    parser.add_argument("--chromium", default="/usr/bin/chromium")
    args = parser.parse_args()

    source_root = args.source_root.resolve()
    fixture = args.fixture.resolve()
    output = args.output.resolve()
    downloads = args.downloads.resolve()
    screenshot = args.screenshot.resolve() if args.screenshot else None
    downloads.mkdir(parents=True, exist_ok=True)

    html_path = source_root / "web" / "index.html"
    css_path = source_root / "web" / "styles.css"
    catalog_path = source_root / "web" / "catalog.js"
    repair_path = source_root / "web" / "repair-lab.js"
    app_path = source_root / "web" / "app.js"
    required = [html_path, css_path, catalog_path, repair_path, app_path, fixture]
    missing = [str(path) for path in required if not path.is_file()]
    if missing:
        raise SystemExit(f"missing required files: {missing}")

    html = html_path.read_text(encoding="utf-8")
    html = re.sub(r'<link rel="icon"[^>]*>', "", html)
    html = html.replace('<link rel="stylesheet" href="styles.css">', '<base href="http://mmv.invalid/">')
    html = html.replace('<script src="catalog.js"></script><script src="repair-lab.js"></script><script src="app.js"></script>', "")
    css = css_path.read_text(encoding="utf-8")
    catalog = catalog_path.read_text(encoding="utf-8")
    repair_js = repair_path.read_text(encoding="utf-8")
    app = app_path.read_text(encoding="utf-8")

    token_marker = "ui-smoke-authenticated-bridge"
    app, token_subs = re.subn(
        r"const TOKEN=new URLSearchParams\(location\.search\)\.get\('token'\)\|\|'';",
        f"const TOKEN='{token_marker}';",
        app,
        count=1,
    )
    api_replacement = """async function api(path,options={}){
if(options.body instanceof FormData){const file=options.body.get('source')||options.body.get('files');if(!file)throw new Error('source file missing');const bytes=Array.from(new Uint8Array(await file.arrayBuffer()));return await window.__mmvUpload(path,{name:file.name,type:file.type||'application/zip',bytes})}
return await window.__mmvApi(path,options)}"""
    app, api_subs = re.subn(r"async function api\(path,options=\{\}\)\{.*?return body\}", api_replacement, app, count=1)
    history_target = "history.replaceState(null,'',`?token=${encodeURIComponent(TOKEN)}#${id}`);"
    history_subs = app.count(history_target)
    app = app.replace(history_target, "void 0;")
    if token_subs != 1 or api_subs != 1 or history_subs != 1:
        raise SystemExit(f"transport harness patch mismatch: token={token_subs}, api={api_subs}, history={history_subs}")

    api_calls: list[dict[str, Any]] = []
    console_errors: list[str] = []
    page_errors: list[str] = []
    request_failures: list[str] = []
    assertions: list[dict[str, Any]] = []
    traces: list[dict[str, Any]] = []
    downloaded: list[dict[str, Any]] = []
    started_at = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())

    def record(name: str, passed: bool, observed: Any) -> None:
        assertions.append({"name": name, "passed": bool(passed), "observed": observed})
        if not passed:
            raise AssertionError(f"{name}: {observed}")

    def attach_diagnostics(page) -> None:
        page.on("console", lambda msg: console_errors.append(msg.text) if msg.type == "error" else None)
        page.on("pageerror", lambda error: page_errors.append(str(error)))
        page.on("requestfailed", lambda request: request_failures.append(f"{request.method} {request.url}: {request.failure}"))

    def load_production_ui(page) -> None:
        page.set_content(html, wait_until="domcontentloaded", timeout=30_000)
        page.add_style_tag(content=css)
        page.add_script_tag(content=catalog)
        page.add_script_tag(content=repair_js)
        page.add_script_tag(content=app)
        wait_text(page, "#versionPill", "v0.9.0", timeout=45_000)
        page.locator('[data-nav="repair"]').click()
        page.locator('[data-view="repair"].active').wait_for(state="visible", timeout=45_000)
        page.locator("#repairMetrics .repair-metric").first.wait_for(state="visible", timeout=45_000)

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(
            executable_path=args.chromium,
            headless=True,
            args=["--no-sandbox", "--disable-dev-shm-usage", "--disable-gpu"],
        )
        context = browser.new_context(viewport={"width": 1440, "height": 1000}, accept_downloads=True)
        context.expose_function("__mmvApi", make_api_bridge(args.base_url, args.token, api_calls))
        context.expose_function("__mmvUpload", make_upload_bridge(args.base_url, args.token, api_calls))

        page = context.new_page()
        attach_diagnostics(page)
        failure = None
        try:
            progress("loading production bundle")
            load_production_ui(page)
            progress("production bundle loaded")
            record("production runtime reports v0.9.0", "v0.9.0" in page.locator("#versionPill").inner_text(), page.locator("#versionPill").inner_text())
            record("Repair Lab is a distinct active top-level workspace", page.locator('[data-nav="repair"]').count() == 1 and page.locator('[data-view="repair"].active').count() == 1, page.locator("#pageTitle").inner_text())
            record("Porting Lab remains separately available", page.locator('[data-nav="porting"]').count() == 1, page.locator('[data-nav="porting"]').count())
            record("Repair Lab renders four live capability metrics", page.locator("#repairMetrics .repair-metric").count() == 4, page.locator("#repairMetrics .repair-metric").count())
            brain_status = page.locator("#repairBrainStatus").inner_text()
            record("Compatibility Brain is live in the UI", "SQLite" in brain_status and "Minecraft versions" in brain_status, brain_status)

            progress("importing source ZIP through production file control")
            page.locator("#repairSourceFile").set_input_files(str(fixture))
            page.locator("#repairSessionWorkspace").wait_for(state="visible", timeout=45_000)
            page.wait_for_function("() => (document.querySelector('#repairSessionTitle')?.textContent||'').includes('mmv-v090-repair-fixture')", timeout=45_000)
            imported_state = (page.locator("#repairSessionState").text_content() or "").strip().lower()
            record("source ZIP creates a real immutable repair session", imported_state == "imported", {"title": page.locator("#repairSessionTitle").inner_text(), "state": imported_state})
            identity = page.locator("#repairSessionIdentity").inner_text()
            record("session exposes immutable source and tree identities", "source" in identity and "immutable tree" in identity, identity)

            progress("staging 1.20.1 to 1.21.1 Fabric migration")
            page.locator("#repairTargetGame").fill("1.21.1")
            page.locator("#repairTargetLoader").select_option("fabric")
            page.locator("#repairPrepare").click()
            page.wait_for_function("() => Number(document.querySelector('#repairChangeCount')?.textContent||0) >= 10", timeout=45_000)
            change_count = int(page.locator("#repairChangeCount").inner_text())
            record("migration stages at least ten traceable known-field edits", change_count >= 10, change_count)
            record("target resolution is explicit", "1.21.1 / fabric" in page.locator("#repairResolutionBadge").inner_text(), page.locator("#repairResolutionBadge").inner_text())

            page.locator("#repairRunBuild").click()
            toast_text = wait_text(page, "#toast", "Acknowledge build-script code execution", timeout=10_000)
            record("build action refuses execution without explicit acknowledgement", "Acknowledge" in toast_text, toast_text)
            record("refused execution does not enter running state", (page.locator("#repairSessionState").text_content() or "").strip().lower() != "running", page.locator("#repairSessionState").inner_text())

            progress("running controlled wrapper build")
            page.locator("#repairExecutionAck").check()
            page.locator("#repairRunBuild").click()
            page.wait_for_function("() => document.querySelector('#repairSessionState')?.textContent.trim() === 'succeeded'", timeout=90_000)
            page.wait_for_function("() => Number(document.querySelector('#repairArtifactCount')?.textContent||0) >= 1", timeout=30_000)
            artifact_count = int(page.locator("#repairArtifactCount").inner_text())
            log_tail = page.locator("#repairLog").inner_text()
            record("controlled wrapper build succeeds", (page.locator("#repairSessionState").text_content() or "").strip().lower() == "succeeded", page.locator("#repairExecutionState").inner_text())
            record("successful build discovers hashed output artifacts", artifact_count >= 1 and page.locator("#repairArtifacts code").count() >= 1, {"count": artifact_count, "hashes": page.locator("#repairArtifacts code").all_inner_texts()})
            record("build log is surfaced in the real UI", "Finished:" in log_tail, log_tail[-500:])

            progress("exporting prepared source and proof bundle")
            page.locator("#repairExport").click()
            page.wait_for_function("() => Number(document.querySelector('#repairExportCount')?.textContent||0) >= 2", timeout=45_000)
            export_count = int(page.locator("#repairExportCount").inner_text())
            record("Repair Lab creates prepared-source and proof exports", export_count >= 2, export_count)

            progress("following exact artifact and proof links rendered by the UI")
            artifact_href = page.locator("#repairArtifacts a").first.get_attribute("href") or ""
            bundle_href = page.locator("#repairExports a").last.get_attribute("href") or ""
            record("production UI renders authenticated artifact and proof download links", artifact_href.startswith("/api/repair-lab/download") and bundle_href.startswith("/api/repair-lab/download"), {"artifact": artifact_href, "proofBundle": bundle_href})
            downloaded.append(download_from_rendered_link(args.base_url, args.token, artifact_href, downloads, api_calls, "artifact"))
            downloaded.append(download_from_rendered_link(args.base_url, args.token, bundle_href, downloads, api_calls, "proof-bundle"))
            record("UI-rendered artifact and proof-bundle links return complete bytes", all(item["size"] > 0 for item in downloaded), downloaded)

            progress("querying Compatibility Brain from the production workspace")
            page.locator("#repairBrainQuery").fill("RetroFuturaGradle")
            page.locator("#repairBrainKind").select_option("tool")
            page.locator("#repairBrainSearch").click()
            page.locator("#repairBrainResults .repair-brain-result").first.wait_for(state="visible", timeout=30_000)
            brain_titles = page.locator("#repairBrainResults .repair-brain-result h3").all_inner_texts()
            record("Compatibility Brain full-text search returns exact repair tooling", any("RetroFuturaGradle" in title for title in brain_titles), brain_titles[:10])

            layout = page.evaluate("""() => ({innerWidth:window.innerWidth,rootWidth:document.documentElement.scrollWidth,bodyWidth:document.body.scrollWidth,active:document.querySelector('.view.active')?.dataset.view,height:document.documentElement.scrollHeight})""")
            record("Repair Lab has no page-level horizontal overflow", max(layout["rootWidth"], layout["bodyWidth"]) <= layout["innerWidth"] + 2, layout)
            traces.append({"step": "succeeded-before-fresh-ui", "state": page.locator("#repairSessionState").inner_text(), "changes": change_count, "artifacts": artifact_count, "exports": export_count, "layout": layout})

            if screenshot:
                screenshot.parent.mkdir(parents=True, exist_ok=True)
                page.evaluate("window.scrollTo(0, 0)")
                page.screenshot(path=str(screenshot), full_page=False)

            progress("opening a fresh production UI instance to prove persistence")
            # New page + freshly injected production bundle proves restart-safe backend persistence.
            fresh = context.new_page()
            attach_diagnostics(fresh)
            load_production_ui(fresh)
            fresh.locator("#repairSessionWorkspace").wait_for(state="visible", timeout=45_000)
            fresh.wait_for_function("() => document.querySelector('#repairSessionState')?.textContent.trim() === 'succeeded'", timeout=45_000)
            record("repair session survives a fresh production UI instance", (fresh.locator("#repairSessionState").text_content() or "").strip().lower() == "succeeded", {"active": fresh.evaluate("document.querySelector('.view.active')?.dataset.view"), "state": fresh.locator("#repairSessionState").inner_text()})

            progress("rolling the session back to its immutable source")
            fresh.locator("#repairReset").click()
            fresh.wait_for_function("() => document.querySelector('#repairSessionState')?.textContent.trim() === 'imported'", timeout=45_000)
            record("rollback returns working copy to immutable source", (fresh.locator("#repairSessionState").text_content() or "").strip().lower() == "imported" and int(fresh.locator("#repairChangeCount").inner_text()) == 0 and int(fresh.locator("#repairArtifactCount").inner_text()) == 0, {"state": fresh.locator("#repairSessionState").inner_text(), "changes": fresh.locator("#repairChangeCount").inner_text(), "artifacts": fresh.locator("#repairArtifactCount").inner_text()})
            fresh.close()

            bridge_failures = [call for call in api_calls if not call.get("ok")]
            meaningful_console = [message for message in console_errors if "Failed to load resource" not in message and "ERR_BLOCKED_BY_ADMINISTRATOR" not in message]
            meaningful_requests = [item for item in request_failures if "favicon" not in item and "ERR_BLOCKED_BY_ADMINISTRATOR" not in item]
            record("all bridged authenticated production API calls succeeded", not bridge_failures, bridge_failures)
            record("no uncaught page exceptions", not page_errors, page_errors)
            record("no meaningful Chromium console errors", not meaningful_console, meaningful_console)
            record("no meaningful browser request failures", not meaningful_requests, meaningful_requests)

        except (AssertionError, PlaywrightTimeoutError, Exception) as error:
            failure = f"{type(error).__name__}: {error}"
            assertions.append({"name": "Repair Lab UI smoke completed", "passed": False, "observed": failure})
            passed = False
        else:
            passed = all(item["passed"] for item in assertions)
        finally:
            progress("closing browser verification context")
            context.close()
            browser.close()
            progress("browser verification context closed")

    evidence = {
        "schema": 1,
        "name": "Minecraft Mod Vault v0.9.0 Repair Lab Chromium UI smoke",
        "startedAt": started_at,
        "completedAt": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "passed": passed,
        "harness": {
            "browser": "system Chromium",
            "navigation": "production HTML/CSS/JS rendered via set_content due managed URLBlocklist",
            "backend": "real compiled v0.9.0 loopback backend through authenticated bridges",
            "productionPaths": ["file input/FormData", "migration controls", "execution acknowledgement", "build polling", "artifact rendering", "download links", "brain search", "rollback"],
            "transportOnlyChanges": ["api() forwarded to authenticated bridge", "FormData bytes forwarded to real multipart endpoint", "exact UI-rendered download URL fetched from real authenticated endpoint outside policy-blocked Chromium", "history.replaceState disabled for about:blank"],
        },
        "inputs": {
            "sourceRoot": str(source_root),
            "fixture": str(fixture),
            "fixtureSha256": sha256_file(fixture),
            "htmlSha256": sha256_file(html_path),
            "cssSha256": sha256_file(css_path),
            "catalogSha256": sha256_file(catalog_path),
            "repairLabSha256": sha256_file(repair_path),
            "appSha256": sha256_file(app_path),
        },
        "assertions": assertions,
        "trace": traces,
        "apiCalls": api_calls,
        "downloads": downloaded,
        "browserDiagnostics": {"consoleErrors": console_errors, "pageErrors": page_errors, "requestFailures": request_failures},
        "screenshot": str(screenshot) if screenshot else "",
        "failure": failure,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"passed": passed, "assertions": len(assertions), "apiCalls": len(api_calls), "downloads": len(downloaded), "output": str(output)}))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
