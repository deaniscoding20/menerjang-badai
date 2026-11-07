"""Local synthetic GitHub/manifest fixtures; NOT live release evidence."""
import json
import os
import tempfile
import unittest
from pathlib import Path
from scripts import vps_delivery as delivery

SHA = "a" * 40
REPO = "deaniscoding20/menerjang-badai"


def run(**changes):
    value = {"id": 123, "event": "push", "head_branch": "main", "conclusion": "success",
             "head_sha": SHA, "head_repository": {"full_name": REPO},
             "repository": {"full_name": REPO}, "status": "completed",
             "path": ".github/workflows/ci.yml"}
    value.update(changes)
    return value


def pull(**changes):
    value = {"number": 7, "merged": True, "merge_commit_sha": SHA,
             "base": {"ref": "main", "repo": {"full_name": REPO}},
             "merged_by": {"login": "deaniscoding20", "id": 351171974}}
    value.update(changes)
    return value


class AuthorizationTest(unittest.TestCase):
    def test_exact_successful_ci_and_owner_merge(self):
        self.assertEqual(delivery.authorize(run(), pull(), SHA), SHA)

    def test_negative_ci_matrix(self):
        for change in ({"event": "pull_request"}, {"head_branch": "other"},
                       {"conclusion": "failure"}, {"status": "in_progress"},
                       {"head_sha": "b" * 40}, {"head_sha": "../main"},
                       {"path": ".github/workflows/other.yml"},
                       {"repository": {"full_name": "foreign/repo"}},
                       {"head_repository": {"full_name": "foreign/repo"}},
                       {"repository": None}, {"head_repository": []}, {"id": True}):
            with self.subTest(change=change), self.assertRaises(delivery.DeliveryError):
                delivery.authorize(run(**change), pull(), SHA)

    def test_negative_owner_pr_matrix(self):
        for change in ({"merged": False}, {"merge_commit_sha": "b" * 40},
                       {"base": {"ref": "other", "repo": {"full_name": REPO}}},
                       {"base": {"ref": "main", "repo": {"full_name": "foreign/repo"}}},
                       {"merged_by": {"login": "other", "id": 351171974}},
                       {"merged_by": {"login": "deaniscoding20", "id": 1}},
                       {"merged_by": None}, {"base": []}, {"number": True}):
            with self.subTest(change=change), self.assertRaises(delivery.DeliveryError):
                delivery.authorize(run(), pull(**change), SHA)

    def test_stale_main_denied(self):
        with self.assertRaises(delivery.DeliveryError):
            delivery.authorize(run(), pull(), "b" * 40)


class FilesystemTest(unittest.TestCase):
    def test_trusted_json_rejects_links_writable_and_oversize_inputs(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            value = root / "manifest.json"
            value.write_text('{"schema_version": 1}')
            value.chmod(0o600)
            self.assertEqual(delivery.trusted_json(value, os.getuid()), {"schema_version": 1})
            link = root / "link.json"
            link.symlink_to(value)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(link, os.getuid())
            value.chmod(0o666)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(value, os.getuid())
            value.chmod(0o600)
            value.write_text('x' * 65537)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(value, os.getuid())

    def test_path_parent_symlink_and_writable_parent_denied(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            private = root / "private"
            private.mkdir(mode=0o700)
            value = private / "data.json"
            value.write_text('{}')
            value.chmod(0o600)
            alias = root / "alias"
            alias.symlink_to(private, target_is_directory=True)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(alias / "data.json", os.getuid())
            private.chmod(0o777)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(value, os.getuid())

    def test_duplicate_json_keys_and_wrong_owner_denied(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder) / "data.json"
            path.write_text('{"sha": "a", "sha": "b"}')
            path.chmod(0o600)
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(path, os.getuid())
            path.write_text('{}')
            with self.assertRaises(delivery.DeliveryError):
                delivery.trusted_json(path, os.getuid() + 1)


if __name__ == "__main__":
    unittest.main()
