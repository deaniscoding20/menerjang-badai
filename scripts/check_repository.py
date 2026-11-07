"""Dependency-free repository checks; reports locations, never secret values.

This targeted secret scan complements, not replaces, Gitleaks/full-history scanning.
"""
import argparse
import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit

SKIP = {".git", ".venv", "node_modules", "__pycache__", ".next", "dist"}
PATTERNS = [
    re.compile(r"gh[pousr]_[A-Za-z0-9]{36,}"),
    re.compile(r"github_pat_[A-Za-z0-9_]{60,}"),
    re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"),
    re.compile(r"\b\d{8,12}:[A-Za-z0-9_-]{30,}\b"),
    re.compile(r"\bAKIA[A-Z0-9]{16}\b"),
]


def files(root):
    return sorted(p for p in root.rglob("*") if p.is_file() and not any(part in SKIP for part in p.relative_to(root).parts))


def check_links(root):
    errors = []
    for path in files(root):
        if path.suffix != ".md":
            continue
        text = path.read_text(encoding="utf-8")
        for match in re.finditer(r"\]\(([^\s)]+)(?:\s+[^)]*)?\)", text):
            target = match.group(1).strip("<>")
            parsed = urlsplit(target)
            if parsed.scheme or parsed.netloc or not parsed.path:
                continue
            local = (root / unquote(parsed.path.lstrip("/"))) if parsed.path.startswith("/") else path.parent / unquote(parsed.path)
            if not local.exists():
                number = text.count("\n", 0, match.start()) + 1
                errors.append(f"{path.relative_to(root)}:{number}: broken local link")
    return errors


def check_secrets(root):
    errors = []
    for path in files(root):
        relative = str(path.relative_to(root))
        if path.name == ".env" or (path.name.startswith(".env.") and not path.name.endswith(".example")):
            errors.append(relative + ": environment secret file is forbidden")
        if path.stat().st_size > 2_000_000:
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeError:
            continue
        for number, line in enumerate(text.splitlines(), 1):
            if any(pattern.search(line) for pattern in PATTERNS):
                errors.append(f"{relative}:{number}: credential pattern detected (redacted)")
    return errors


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    root = args.root.resolve()
    errors = check_links(root) + check_secrets(root)
    for error in errors:
        print(error, file=sys.stderr)
    if errors:
        raise SystemExit(1)
    print("Local documentation links and targeted secret scan passed")


if __name__ == "__main__":
    main()
