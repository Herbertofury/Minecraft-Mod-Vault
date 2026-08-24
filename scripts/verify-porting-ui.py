#!/usr/bin/env python3
"""Run the production Minecraft Mod Vault UI in Chromium against a real backend.

This verifier intentionally avoids direct loopback navigation so it can run on managed
hosts whose Chromium policy blocks all URLs. Production HTML/CSS/JS are rendered in
Chromium with an authenticated bridge that forwards the exact UI API calls to the
already-running local Vault backend. No application handlers are replaced besides the
transport helper and history mutation required by the about:blank harness.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

from playwright.sync_api import Page, TimeoutError as PlaywrightTimeoutError, sync_playwright


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def wait_text(page: Page, selector: str, expected: str, timeout: int = 30_000) -> str:
    page.wait_for_function(
        "([selector, expected]) => { const el = document.querySelector(selector); return !!el && (el.textContent || '').includes(expected); }",
        arg=[selector, expected],
        timeout=timeout,
    )
    return page.locator(selector).inner_text()


def wait_selector_count(page: Page, selector: str, minimum: int, timeout: int = 30_000) -> int:
    page.wait_for_function(
        "([selector, minimum]) => document.querySelectorAll(selector).length >= minimum",
        arg=[selector, minimum],
        timeout=timeout,
    )
    return page.locator(selector).count()


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
            with urllib.request.urlopen(request, timeout=60) as response:
                raw = response.read()
                status = response.status
        except urllib.error.HTTPError as error:
            raw = error.read()
            status = error.code
            try:
                parsed = json.loads(raw.decode("utf-8")) if raw else {}
            except Exception:
                parsed = {}
            message = parsed.get("error") or f"HTTP {status}"
            call_log.append(
                {
                    "method": method,
                    "path": path.split("?", 1)[0],
                    "status": status,
                    "durationMs": round((time.time() - started) * 1000),
                    "ok": False,
                }
            )
            raise RuntimeError(message) from error
        parsed = json.loads(raw.decode("utf-8")) if raw else {}
        call_log.append(
            {
                "method": method,
                "path": path.split("?", 1)[0],
                "status": status,
                "durationMs": round((time.time() - started) * 1000),
                "ok": True,
            }
        )
        return parsed

    return bridge


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-root", required=True, type=Path)
    parser.add_argument("--runtime-root", required=True, type=Path)
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--token", required=True)
    parser.add_argument("--fixture", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--screenshot", type=Path)
    parser.add_argument("--chromium", default="/usr/bin/chromium")
    args = parser.parse_args()

    source_root = args.source_root.resolve()
    runtime_root = args.runtime_root.resolve()
    fixture = args.fixture.resolve()
    output = args.output.resolve()
    screenshot = args.screenshot.resolve() if args.screenshot else None

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
    html = html.replace('<link rel="stylesheet" href="styles.css">', "")
    html = html.replace('<script src="catalog.js"></script><script src="repair-lab.js"></script><script src="app.js"></script>', "")
    css = css_path.read_text(encoding="utf-8")
    catalog = catalog_path.read_text(encoding="utf-8")
    repair = repair_path.read_text(encoding="utf-8")
    app = app_path.read_text(encoding="utf-8")

    token_replacement = "const TOKEN='ui-smoke-authenticated-bridge';"
    app, token_subs = re.subn(
        r"const TOKEN=new URLSearchParams\(location\.search\)\.get\('token'\)\|\|'';",
        token_replacement,
        app,
        count=1,
    )
    api_replacement = "async function api(path,options={}){return await window.__mmvApi(path,options)}"
    app, api_subs = re.subn(
        r"async function api\(path,options=\{\}\)\{.*?return body\}",
        api_replacement,
        app,
        count=1,
    )
    history_target = "history.replaceState(null,'',`?token=${encodeURIComponent(TOKEN)}#${id}`);"
    history_subs = app.count(history_target)
    app = app.replace(history_target, "void 0;")
    if token_subs != 1 or api_subs != 1 or history_subs != 1:
        raise SystemExit(
            f"transport harness patch mismatch: token={token_subs}, api={api_subs}, history={history_subs}"
        )

    api_calls: list[dict[str, Any]] = []
    console_errors: list[str] = []
    page_errors: list[str] = []
    request_failures: list[str] = []
    assertions: list[dict[str, Any]] = []
    screenshots: list[str] = []
    before_fixture_sha = sha256_file(fixture)
    duplicate_path = fixture.with_name("vault-fixture-copy.jar")
    before_duplicate_sha = sha256_file(duplicate_path)
    if before_fixture_sha != before_duplicate_sha:
        raise SystemExit("duplicate fixture is not byte-identical before UI smoke")

    def record(name: str, passed: bool, observed: Any) -> None:
        assertions.append({"name": name, "passed": bool(passed), "observed": observed})
        if not passed:
            raise AssertionError(f"{name}: {observed}")

    started_at = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(
            executable_path=args.chromium,
            headless=True,
            args=["--no-sandbox", "--disable-dev-shm-usage", "--disable-gpu"],
        )
        context = browser.new_context(viewport={"width": 1440, "height": 1000}, device_scale_factor=1)
        page = context.new_page()
        page.on("console", lambda msg: console_errors.append(msg.text) if msg.type == "error" else None)
        page.on("pageerror", lambda error: page_errors.append(str(error)))
        page.on(
            "requestfailed",
            lambda request: request_failures.append(f"{request.method} {request.url}: {request.failure}"),
        )
        page.expose_function("__mmvApi", make_api_bridge(args.base_url, args.token, api_calls))

        try:
            page.set_content(html, wait_until="domcontentloaded", timeout=30_000)
            page.add_style_tag(content=css)
            page.add_script_tag(content=catalog)
            page.add_script_tag(content=repair)
            page.add_script_tag(content=app)

            wait_text(page, "#versionPill", "v0.13.0", timeout=45_000)
            record("production bootstrap reports v0.13.0", True, page.locator("#versionPill").inner_text())
            record(
                "navigation includes distinct Porting Lab",
                page.locator('[data-nav="porting"]').count() == 1,
                page.locator('[data-nav="porting"]').count(),
            )
            record(
                "navigation preserves distinct source Repair Lab",
                page.locator('[data-nav="repair"]').count() == 1,
                page.locator('[data-nav="repair"]').count(),
            )

            page.locator('[data-nav="manager"]').click()
            wait_selector_count(page, "#managerGroups [data-porting-file]", 2)
            manager_buttons = page.locator("#managerGroups [data-porting-file]")
            selected_index = -1
            selected_path = ""
            for index in range(manager_buttons.count()):
                candidate = manager_buttons.nth(index).get_attribute("data-porting-file") or ""
                if candidate.endswith("/vault-fixture.jar") or candidate.endswith("\\vault-fixture.jar"):
                    selected_index = index
                    selected_path = candidate
                    break
            record("Manager exposes exact JAR Porting Lab action", selected_index >= 0, selected_path)
            manager_buttons.nth(selected_index).click()
            page.wait_for_function(
                "() => document.querySelector('[data-view=porting]')?.classList.contains('active')"
            )
            record(
                "Manager action opens real Porting Lab workspace",
                page.locator('[data-view="porting"].active').count() == 1,
                page.locator("#pageTitle").inner_text(),
            )
            record(
                "Manager transfers exact selected JAR path",
                page.locator("#portingInputJar").input_value() == selected_path,
                page.locator("#portingInputJar").input_value(),
            )
            record(
                "Manager switches starting evidence to binary",
                page.locator("#portingSourceMode").input_value() == "binary",
                page.locator("#portingSourceMode").input_value(),
            )

            page.locator("#portingSourceGame").fill("1.20.1")
            page.locator("#portingSourceLoader").select_option("forge")
            page.locator("#portingTargetGame").fill("1.21.1")
            page.locator("#portingTargetLoader").select_option("neoforge")
            project_name = f"ui-smoke-port-{int(time.time())}"
            page.locator("#portingProjectName").fill(project_name)

            page.locator("#portingProbeEnvironment").click()
            environment_text = wait_text(page, "#portingEnvironment", "Target 1.21.1", timeout=30_000)
            record("toolchain probe renders target Java evidence", "Java" in environment_text, environment_text[:500])

            page.locator("#portingBuildPlan").click()
            page.wait_for_selector("#portingPlanResult .porting-plan-hero", timeout=60_000)
            phase_count = page.locator("#portingPlanResult .porting-phase").count()
            plan_text = page.locator("#portingPlanResult").inner_text()
            record("Porting Lab renders eight gated migration phases", phase_count == 8, phase_count)
            record("live JAR forensics identifies fixture mod", "vault_fixture" in plan_text, "vault_fixture" in plan_text)
            record("plan selects current InterMed evidence", "InterMed" in plan_text, "InterMed" in plan_text)
            record("plan selects current modcrawl evidence", "modcrawl" in plan_text.lower(), "modcrawl" in plan_text.lower())
            record("plan is cross-loader critical-risk route", "critical" in plan_text.lower(), plan_text[:500])
            record(
                "workspace generation unlocks only after a plan",
                page.locator("#portingGenerateWorkspace").is_enabled(),
                page.locator("#portingGenerateWorkspace").is_enabled(),
            )

            workspace_before = page.locator("#portingWorkspaceGrid .porting-workspace").count()
            page.locator("#portingGenerateWorkspace").click()
            wait_text(page, "#portingStatus", "created with", timeout=60_000)
            page.wait_for_function(
                "minimum => document.querySelectorAll('#portingWorkspaceGrid .porting-workspace').length > minimum",
                arg=workspace_before,
                timeout=60_000,
            )
            workspace_after = page.locator("#portingWorkspaceGrid .porting-workspace").count()
            record("workspace generation produces a new persistent card", workspace_after > workspace_before, {"before": workspace_before, "after": workspace_after})
            workspace_text = page.locator("#portingWorkspaceGrid").inner_text()
            record("generated workspace preserves requested project identity", project_name in workspace_text, project_name)

            layout = page.evaluate(
                "() => ({innerWidth: window.innerWidth, rootWidth: document.documentElement.scrollWidth, bodyWidth: document.body.scrollWidth})"
            )
            record(
                "Porting Lab desktop layout has no page-level horizontal overflow",
                max(layout["rootWidth"], layout["bodyWidth"]) <= layout["innerWidth"] + 1,
                layout,
            )
            if screenshot:
                screenshot.parent.mkdir(parents=True, exist_ok=True)
                page.evaluate("window.scrollTo(0, 0)")
                document_height = page.evaluate("Math.min(3000, document.documentElement.scrollHeight)")
                page.screenshot(
                    path=str(screenshot),
                    clip={"x": 0, "y": 0, "width": 1440, "height": document_height},
                )
                screenshots.append(str(screenshot))

            page.locator('[data-nav="doctor"]').click()
            page.wait_for_function(
                "() => document.querySelector('[data-view=doctor]')?.classList.contains('active')"
            )
            page.locator("#doctorBuildRepairPlan").click()
            wait_selector_count(page, "#doctorRepairPlan [data-repair-action]", 1, timeout=60_000)
            repair_count = page.locator("#doctorRepairPlan [data-repair-action]").count()
            record("Doctor exposes only proven duplicate quarantine actions", repair_count == 1, repair_count)
            record(
                "quarantine action is selected and apply enabled",
                page.locator("#doctorRepairPlan [data-repair-action]:checked").count() == 1
                and page.locator("#doctorApplyRepairPlan").is_enabled(),
                {"checked": page.locator("#doctorRepairPlan [data-repair-action]:checked").count(), "enabled": page.locator("#doctorApplyRepairPlan").is_enabled()},
            )

            page.locator("#doctorApplyRepairPlan").click()
            receipt_text = wait_text(page, "#doctorRepairPlan", "Duplicate quarantine completed without deletion", timeout=60_000)
            record("repair apply renders recoverable receipt", "applied-recoverable" in receipt_text, receipt_text[:600])
            record("duplicate leaves managed mods directory after quarantine", not duplicate_path.exists(), str(duplicate_path))
            record("keeper remains present after quarantine", fixture.is_file(), str(fixture))

            page.locator("#doctorRepairPlan [data-restore-repair]").click()
            restored_text = wait_text(page, "#doctorRepairPlan", "restored", timeout=60_000)
            record("receipt restore returns quarantined JAR", "restored" in restored_text.lower(), restored_text)
            record("duplicate path restored", duplicate_path.is_file(), str(duplicate_path))
            record("restored duplicate remains byte-identical", sha256_file(duplicate_path) == before_duplicate_sha, sha256_file(duplicate_path))
            record("keeper hash unchanged", sha256_file(fixture) == before_fixture_sha, sha256_file(fixture))

            bridge_failures = [call for call in api_calls if not call["ok"]]
            record("all bridged authenticated UI API calls succeeded", not bridge_failures, bridge_failures)
            record("no uncaught page exceptions", not page_errors, page_errors)
            # Resource load failures from external artwork are irrelevant to the local Porting/Doctor flows;
            # keep them visible in evidence but only JavaScript console errors fail this smoke.
            meaningful_console_errors = [
                message
                for message in console_errors
                if "Failed to load resource" not in message and "ERR_BLOCKED_BY_ADMINISTRATOR" not in message
            ]
            record("no meaningful Chromium console errors", not meaningful_console_errors, meaningful_console_errors)

        except (AssertionError, PlaywrightTimeoutError, Exception) as error:
            # Preserve evidence even when the smoke fails; the caller receives a non-zero exit.
            failure = f"{type(error).__name__}: {error}"
            assertions.append({"name": "ui smoke completed", "passed": False, "observed": failure})
            passed = False
        else:
            failure = None
            passed = all(item["passed"] for item in assertions)
        finally:
            context.close()
            browser.close()

    output.parent.mkdir(parents=True, exist_ok=True)
    evidence = {
        "schema": 1,
        "name": "Minecraft Mod Vault v0.13.0 Porting Lab Chromium UI smoke",
        "startedAt": started_at,
        "completedAt": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "passed": passed,
        "harness": {
            "browser": "system Chromium",
            "navigation": "production HTML/CSS/JS rendered via set_content due managed URLBlocklist",
            "backend": "real compiled v0.13.0 loopback backend through authenticated Playwright binding",
            "applicationTransportChanges": [
                "api() forwarded to authenticated bridge",
                "history.replaceState disabled for about:blank",
                "TOKEN replaced with non-secret harness marker",
            ],
            "notClaimed": [
                "direct loopback URL navigation",
                "browser reload persistence through URL",
                "external artwork network availability",
            ],
        },
        "inputs": {
            "sourceRoot": str(source_root),
            "runtimeRoot": str(runtime_root),
            "fixture": str(fixture),
            "fixtureSha256": before_fixture_sha,
            "htmlSha256": sha256_file(html_path),
            "cssSha256": sha256_file(css_path),
            "catalogSha256": sha256_file(catalog_path),
            "repairLabSha256": sha256_file(repair_path),
            "appSha256": sha256_file(app_path),
        },
        "assertions": assertions,
        "apiCalls": api_calls,
        "browserDiagnostics": {
            "consoleErrors": console_errors,
            "pageErrors": page_errors,
            "requestFailures": request_failures,
        },
        "screenshots": screenshots,
        "failure": failure,
    }
    output.write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"passed": passed, "assertions": len(assertions), "apiCalls": len(api_calls), "output": str(output)}))
    return 0 if passed else 1


if __name__ == "__main__":
    sys.exit(main())
