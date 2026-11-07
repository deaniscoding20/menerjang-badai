"""Trusted installed VPS metadata gate. Never import/execute candidate repository code."""
import json
import os
import re
import stat
from contextlib import contextmanager
from pathlib import Path


REPOSITORY = "deaniscoding20/menerjang-badai"
OWNER = "deaniscoding20"
OWNER_ID = 351171974
CI_PATH = ".github/workflows/ci.yml"


class DeliveryError(Exception):
    """Only fixed public reason codes may cross the CLI boundary."""


def authorize(run, pull, latest):
    """Require authoritative same-repo CI and owner PR at the exact current main."""
    if not isinstance(run, dict) or not isinstance(pull, dict):
        raise DeliveryError("metadata_invalid")
    sha = run.get("head_sha")
    if not (type(run.get("id")) is int and run["id"] > 0
            and isinstance(sha, str) and re.fullmatch(r"[0-9a-f]{40}", sha)
            and isinstance(run.get("repository"), dict)
            and isinstance(run.get("head_repository"), dict)):
        raise DeliveryError("metadata_invalid")
    if not (isinstance(sha, str) and re.fullmatch(r"[0-9a-f]{40}", sha)
            and sha == latest and run.get("event") == "push"
            and run.get("head_branch") == "main" and run.get("status") == "completed"
            and run.get("conclusion") == "success" and run.get("path") == CI_PATH
            and run.get("repository", {}).get("full_name") == REPOSITORY
            and run.get("head_repository", {}).get("full_name") == REPOSITORY):
        raise DeliveryError("ci_or_main_denied")
    merged_by = pull.get("merged_by")
    base = pull.get("base")
    if not (type(pull.get("number")) is int and pull["number"] > 0
            and isinstance(merged_by, dict) and isinstance(base, dict)
            and isinstance(base.get("repo"), dict)):
        raise DeliveryError("owner_merge_denied")
    if not (pull.get("merged") is True and pull.get("merge_commit_sha") == sha
            and base.get("ref") == "main"
            and (base.get("repo") or {}).get("full_name") == REPOSITORY
            and merged_by.get("login") == OWNER
            and type(merged_by.get("id")) is int and merged_by["id"] == OWNER_ID):
        raise DeliveryError("owner_merge_denied")
    return sha
