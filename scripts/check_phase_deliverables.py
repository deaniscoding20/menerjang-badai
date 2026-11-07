"""Inventory only: file existence does not establish phase acceptance."""
from pathlib import Path

REQUIRED = ["docs/00-project-overview.md", "docs/02-architecture-overview.md", "docs/03-hexagonal-architecture.md", "docs/04-project-structure.md", "docs/05-domain-model.md", "docs/06-database-design.md", "docs/08-authentication-security.md", "docs/09-hermes-integration.md", "docs/10-postiz-integration.md", "docs/architecture/system-context.md", "docs/architecture/container-diagram.md", "docs/architecture/component-diagram.md", "docs/architecture/deployment-diagram.md", "docs/architecture/data-flow.md", "docs/adr/ADR-0001-initial-architecture.md", "docs/adr/ADR-0002-human-authentication.md", "docs/adr/ADR-0003-agent-integration.md", "docs/adr/ADR-0004-workflow-orchestration.md"]
DIAGRAMS = [name for name in REQUIRED if "/architecture/" in name] + ["docs/06-database-design.md"]


def check(root):
    errors = []
    for name in REQUIRED:
        path = root / name
        if not path.is_file() or not path.read_text(encoding="utf-8").strip():
            errors.append(name + ": missing or empty")
        elif name in DIAGRAMS and "```mermaid" not in path.read_text(encoding="utf-8"):
            errors.append(name + ": Mermaid block required")
    return errors


if __name__ == "__main__":
    errors = check(Path(__file__).resolve().parents[1])
    if errors:
        raise SystemExit("\n".join(errors))
    print("18 required Phase0 deliverables present; semantic approval remains manual")
