"""CLI contract tests: GitHub HTTP boundary mocked; no real authorization claims."""
import contextlib
import io
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
from scripts import release_policy

SHA = "a" * 40
REPO = "deaniscoding20/menerjang-badai"


class ReleaseCLITest(unittest.TestCase):
    def invoke(self, current_sha=SHA, merged=True):
        event = {"workflow_run": {"event": "push", "head_branch": "main", "conclusion": "success", "head_sha": SHA, "head_repository": {"full_name": REPO}, "status": "completed", "path": ".github/workflows/ci.yml", "id": 123}}
        responses = [
            event["workflow_run"],
            [{"number": 1}],
            {"merged": merged, "merge_commit_sha": SHA, "base": {"ref": "main", "repo": {"full_name": REPO}}, "merged_by": {"login": "deaniscoding20", "id": 351171974}},
            {"object": {"sha": current_sha}},
        ]
        with tempfile.TemporaryDirectory() as folder:
            event_path = Path(folder) / "event.json"
            output = Path(folder) / "output"
            event_path.write_text(json.dumps(event))
            args = ["release_policy", "--event-file", str(event_path), "--repository", REPO, "--owner", "deaniscoding20"]
            # Unit boundary: deterministic GitHub metadata, not an API integration test.
            def reply(request, timeout):
                self.assertTrue(request.full_url.startswith("https://api.github.com/repos/" + REPO + "/"))
                self.assertEqual(timeout, 25)
                return io.BytesIO(json.dumps(responses.pop(0)).encode())
            with patch("sys.argv", args), patch.dict(os.environ, {"GH_TOKEN": "unit-fixture-only", "GITHUB_OUTPUT": str(output)}), patch("urllib.request.urlopen", side_effect=reply), contextlib.redirect_stdout(io.StringIO()):
                release_policy.main()
            return output.read_text()

    def test_owner_merge_outputs_exact_sha(self):
        self.assertEqual(self.invoke(), "eligible=true\nsha=" + SHA + "\n")

    def test_late_ci_cannot_redeploy_old_main(self):
        self.assertIn("eligible=false", self.invoke(current_sha="b" * 40))

    def test_open_pr_is_denied(self):
        self.assertIn("eligible=false", self.invoke(merged=False))


if __name__ == "__main__":
    unittest.main()
