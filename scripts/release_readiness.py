"""Validate release marker; passing is necessary but not production evidence."""
import argparse
import json
import os
from pathlib import Path

REQUIRED = ["go.mod", "web/package-lock.json", "deploy/compose.production.yaml",
            "docs/reports/production-readiness.md", "scripts/deploy.sh"]


def validate(config, root):
    if type(config.get("schema_version")) is not int or config["schema_version"] != 1:
        raise ValueError("Unsupported release marker schema")
    if type(config.get("production_ready")) is not bool:
        raise ValueError("production_ready must be boolean")
    if type(config.get("phase")) is not int or not 0 <= config["phase"] <= 7:
        raise ValueError("Invalid phase")
    if (config.get("domain"), config.get("bind"), config.get("port")) != ("menerjangbadai.online", "127.0.0.1", 2023):
        raise ValueError("Unapproved deployment target")
    if config["production_ready"]:
        if config["phase"] != 7:
            raise ValueError("Production requires completed Phase 7 and owner approval")
        missing = [name for name in REQUIRED if not (root / name).is_file()]
        if missing:
            raise ValueError("Missing release inputs: " + ", ".join(missing))
    return config["production_ready"]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    with (args.root / "deploy/release.json").open(encoding="utf-8") as stream:
        deployable = validate(json.load(stream), args.root)
    if os.getenv("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as stream:
            stream.write("deployable=" + str(deployable).lower() + "\n")
    print("Release marker valid; production deployable:", deployable)


if __name__ == "__main__":
    main()
