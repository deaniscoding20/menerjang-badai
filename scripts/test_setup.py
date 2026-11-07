"""Explicit testing profiles. No production targets, secrets or tests by default.

plan: metadata only; fast: local short checks; database: disposable Docker PG;
full: refuse until long suite implementations/prerequisites are available.
"""
import argparse
import datetime as dt
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]


def redact(text, secrets=()):
    for value in secrets:
        if value:
            text = text.replace(value, "[REDACTED]")
    text = re.sub(r"postgres(?:ql)?://[^\s\"']+", "[DATABASE_URL_REDACTED]", text)
    text = re.sub(r"gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{40,}", "[TOKEN_REDACTED]", text)
    text = re.sub(r"(?i)(authorization\s*[:=]\s*)([^\r\n]+)", r"\1[REDACTED]", text)
    return text


def full_blockers(config):
    suites = config.get("long_suites")
    if not isinstance(suites, list) or not suites:
        return ["long-suite-inventory-missing"]
    return [str(suite.get("id", "unnamed")) for suite in suites
            if suite.get("status") != "implemented"
            or not isinstance(suite.get("command"), list)
            or not suite["command"]
            or not all(isinstance(arg, str) for arg in suite["command"])]


def go_binary():
    command = shutil.which("go")
    if command:
        return command
    local = Path.home() / ".local/lib/go/bin/go"
    return str(local) if local.is_file() else "go"


def fast_steps(root, scope):
    steps = []
    if scope in ("all", "python"):
        steps.append(("python-policy", [sys.executable, "-m", "unittest", "discover", "-s", "tests", "-v"], root))
    if scope in ("all", "backend"):
        # Explicit empty DSN below means DB tests can only SKIP in this profile.
        steps.extend([
            ("go-compile", [go_binary(), "build", "./..."], root),
            ("go-focused-units", [go_binary(), "test", "./internal/domain", "./internal/application", "./internal/platform"], root),
        ])
    if scope in ("all", "frontend"):
        for name, command in [
            ("frontend-types", ["npm", "run", "typecheck"]),
            ("frontend-unit", ["npm", "test", "--", "--reporter=dot"]),
            ("frontend-build", ["npm", "run", "build"]),
        ]:
            steps.append((name, command, root / "web"))
    return steps


def write_report(root, result):
    # Require a real local directory; do not follow an existing report-root link.
    root = Path(root)
    if root.is_symlink():
        raise ValueError("Evidence directory must not be a symlink")
    root.mkdir(parents=True, exist_ok=True, mode=0o700)
    root.chmod(0o700)
    stamp = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%S")
    path = root / (stamp + "-" + secrets.token_hex(4) + ".json")
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, "w", encoding="utf-8") as stream:
        json.dump(result, stream, indent=2)
        stream.write("\n")
    return path


def execute(name, command, cwd, env, secret_values=(), timeout=120):
    started = time.monotonic()
    try:
        run = subprocess.run(command, cwd=cwd, env=env, text=True,
                             stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=timeout)
        code, output = run.returncode, run.stdout
    except subprocess.TimeoutExpired:
        code, output = 124, "Step exceeded its bounded timeout; partial output withheld."
    except FileNotFoundError:
        code, output = 127, "Required executable unavailable; follow testing setup prerequisites."
    clean = redact(output, secret_values)
    print(f"\n[{name}] exit={code}\n{clean}")
    return {"step": name, "exit_code": code, "seconds": round(time.monotonic() - started, 2), "output": clean}


def disposable_database(config, env, report):
    """Create new memory-backed DB; never accept an external/prod DSN."""
    if not shutil.which("docker"):
        raise RuntimeError("Docker is required; no host PostgreSQL fallback is allowed")
    image = config.get("database_image", "")
    if not re.fullmatch(r"postgres:17@sha256:[0-9a-f]{64}", image):
        raise ValueError("Database image must have a reviewed immutable digest")
    name = "mb-testing-" + secrets.token_hex(6)
    password = secrets.token_urlsafe(36)
    with tempfile.TemporaryDirectory(prefix="mb-test-secret-") as folder:
        private = Path(folder)
        private.chmod(0o700)
        password_file = private / "postgres-password"
        password_file.write_text(password, encoding="utf-8")
        password_file.chmod(0o600)
        created = False
        cleanup_ok = False
        try:
            command = ["docker", "run", "--detach", "--name", name,
                       "--label", "mb.testing=disposable", "--memory", "512m", "--cpus", "1",
                       "--tmpfs", "/var/lib/postgresql/data:rw,nosuid,nodev,size=384m",
                       "--mount", f"type=bind,src={password_file},dst=/run/secrets/pg_password,readonly",
                       "--env", "POSTGRES_PASSWORD_FILE=/run/secrets/pg_password",
                       "--env", "POSTGRES_USER=mb_testing", "--env", "POSTGRES_DB=mb_test_disposable",
                       "--publish", "127.0.0.1::5432", image]
            # Mark before attempt so even a timed-out docker client gets cleanup.
            created = True
            step = execute("disposable-postgres-start", command, ROOT, env, [password], 180)
            report["steps"].append(step)
            if step["exit_code"]:
                raise RuntimeError("Disposable database could not start")
            for _ in range(40):
                probe = subprocess.run(["docker", "exec", name, "pg_isready", "-U", "mb_testing", "-d", "mb_test_disposable"],
                                       capture_output=True, timeout=5)
                if probe.returncode == 0:
                    break
                time.sleep(0.5)
            else:
                raise RuntimeError("Disposable database readiness timed out")
            port = subprocess.run(["docker", "port", name, "5432/tcp"], text=True, capture_output=True, check=True, timeout=5).stdout.strip()
            match = re.fullmatch(r"127\.0\.0\.1:(\d+)", port)
            if not match:
                raise RuntimeError("Disposable database is not loopback-only")
            dsn = f"postgres://mb_testing:{password}@127.0.0.1:{match[1]}/mb_test_disposable?sslmode=disable"
            db_env = dict(env, MB_INTEGRATION="1", MB_TEST_DATABASE_URL=dsn)
            report["steps"].append(execute("postgres-integration", [go_binary(), "test", "./internal/adapters/postgres", "-count=1", "-timeout=90s"], ROOT, db_env, [password, dsn], 120))
        finally:
            if created:
                try:
                    result = subprocess.run(["docker", "rm", "--force", name], capture_output=True, timeout=20)
                    cleanup_ok = result.returncode == 0
                except (OSError, subprocess.TimeoutExpired):
                    cleanup_ok = False
                report["steps"].append({"step": "disposable-postgres-cleanup", "exit_code": 0 if cleanup_ok else 1,
                                        "output": "Removed isolated database" if cleanup_ok else "Cleanup incomplete: remove container " + name})
            if created and not cleanup_ok:
                raise RuntimeError("Disposable database cleanup requires attention")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("profile", choices=["plan", "fast", "database", "full"])
    parser.add_argument("--scope", choices=["all", "python", "backend", "frontend"], default="all")
    parser.add_argument("--confirm-disposable-db", action="store_true")
    parser.add_argument("--confirm-staging", action="store_true")
    args = parser.parse_args()
    config = json.loads((ROOT / "testing/suites.json").read_text())
    if args.profile == "plan":
        print(json.dumps({"mode": config["mode"], "tools": {name: bool(shutil.which(name)) for name in ["docker", "node", "npm"]},
                          "go_available": shutil.which(go_binary()) is not None,
                          "long_tests": config["long_suites"], "tests_executed": False}, indent=2))
        return 0
    report = {"profile": args.profile, "started_at": dt.datetime.now(dt.timezone.utc).isoformat(), "steps": [],
              "long_tests_executed": False, "full_testing_complete": False}
    env = dict(os.environ, MB_INTEGRATION="0", MB_TEST_DATABASE_URL="", MB_DATABASE_URL="", MB_DATABASE_URL_FILE="")
    try:
        if args.profile == "fast":
            for name, command, cwd in fast_steps(ROOT, args.scope):
                report["steps"].append(execute(name, command, cwd, env))
        elif args.profile == "database":
            if not args.confirm_disposable_db:
                raise ValueError("Use --confirm-disposable-db; this profile creates and removes a new isolated container")
            disposable_database(config, env, report)
        else:
            blockers = full_blockers(config)
            if blockers:
                report["blockers"] = blockers
                raise ValueError("Full testing NOT READY: long suites still planned or lack executable commands")
            if not args.confirm_staging:
                raise ValueError("Full testing requires an isolated staging review and --confirm-staging")
            report["long_tests_executed"] = True
            for suite in config["long_suites"]:
                report["steps"].append(execute(suite["id"], suite["command"], ROOT, env, timeout=900))
            report["full_testing_complete"] = bool(report["steps"]) and all(s["exit_code"] == 0 for s in report["steps"])
    except (ValueError, RuntimeError, OSError, subprocess.SubprocessError) as exc:
        # Unexpected command errors can contain arguments; emit no repr/traceback.
        message = str(exc) if isinstance(exc, (ValueError, RuntimeError)) else "Harness prerequisite/operation failed; inspect local tools safely"
        report["error"] = redact(message)
        print(report["error"], file=sys.stderr)
    report["status"] = "failed" if report.get("error") or any(s["exit_code"] != 0 for s in report["steps"]) else "passed_for_selected_profile_only"
    print("Evidence:", write_report(ROOT / "artifacts/testing", report))
    return 1 if report["status"] == "failed" else 0


if __name__ == "__main__":
    raise SystemExit(main())
