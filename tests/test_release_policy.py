"""Release authorization tests. Fixtures are policy inputs, not live GitHub results."""
import unittest
from scripts.release_policy import authorize

SHA = "a" * 40
REPO = "deaniscoding20/menerjang-badai"


def run(**changes):
    value = {"event": "push", "branch": "main", "conclusion": "success", "sha": SHA, "repository": REPO,
             "status": "completed", "workflow_path": ".github/workflows/ci.yml"}
    value.update(changes)
    return value


def pull(**changes):
    value = {"merged": True, "merge_commit_sha": SHA, "base": {"ref": "main", "repo": {"full_name": REPO}}, "merged_by": {"login": "deaniscoding20", "id": 351171974}}
    value.update(changes)
    return value


class ReleasePolicyTest(unittest.TestCase):
    def test_owner_merged_main_with_successful_push_ci_is_authorized(self):
        self.assertTrue(authorize(run(), [pull()], REPO, "deaniscoding20"))

    def test_direct_push_is_not_authorized(self):
        self.assertFalse(authorize(run(), [], REPO, "deaniscoding20"))

    def test_wrong_actor_or_sha_or_open_pr_is_not_authorized(self):
        for pr in [pull(merged_by={"login": "someone"}), pull(merge_commit_sha="b" * 40), pull(merged=False)]:
            with self.subTest(pr=pr):
                self.assertFalse(authorize(run(), [pr], REPO, "deaniscoding20"))

    def test_untrusted_ci_context_is_not_authorized(self):
        for change in [{"event": "pull_request"}, {"branch": "feat/test"}, {"conclusion": "failure"}, {"repository": "fork/example"}, {"sha": "not-a-sha"}]:
            with self.subTest(change=change):
                self.assertFalse(authorize(run(**change), [pull()], REPO, "deaniscoding20"))

    def test_cross_repo_or_wrong_base_is_not_authorized(self):
        for base in [{"ref": "dev", "repo": {"full_name": REPO}}, {"ref": "main", "repo": {"full_name": "fork/example"}}]:
            self.assertFalse(authorize(run(), [pull(base=base)], REPO, "deaniscoding20"))

    def test_unrelated_workflow_or_running_ci_is_denied(self):
        for change in [{"workflow_path": ".github/workflows/untrusted.yml"}, {"status": "in_progress"}]:
            self.assertFalse(authorize(run(**change), [pull()], REPO, "deaniscoding20"))

    def test_matching_login_with_wrong_numeric_owner_is_denied(self):
        self.assertFalse(authorize(run(), [pull(merged_by={"login": "deaniscoding20", "id": 42})], REPO, "deaniscoding20"))

    def test_missing_fields_fail_closed(self):
        self.assertFalse(authorize({}, [{}], REPO, "deaniscoding20"))
        self.assertFalse(authorize(run(), [{"merged": True}], REPO, "deaniscoding20"))


if __name__ == "__main__":
    unittest.main()
