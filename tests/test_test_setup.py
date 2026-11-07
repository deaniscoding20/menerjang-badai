"""Test the test harness itself; no Docker, external APIs or production data."""
import tempfile
import unittest
from pathlib import Path
from scripts import test_setup


class HarnessTest(unittest.TestCase):
    def test_redaction_removes_ephemeral_secret_and_database_url(self):
        secret = "unit-" + "sensitive" + "-fixture"
        raw = f"error {secret} postgres://someone:password@localhost/test and ghp_" + "A" * 36
        safe = test_setup.redact(raw, [secret])
        self.assertNotIn(secret, safe)
        self.assertNotIn("password", safe)
        self.assertNotIn("A" * 36, safe)

    def test_long_suite_plan_cannot_claim_full_readiness(self):
        config = {"long_suites": [{"id": "real-oidc", "status": "planned"}]}
        self.assertEqual(test_setup.full_blockers(config), ["real-oidc"])

    def test_unknown_or_missing_long_suite_status_is_blocked(self):
        config = {"long_suites": [{"id": "missing"}, {"id": "ambiguous", "status": "passed"}]}
        self.assertEqual(test_setup.full_blockers(config), ["missing", "ambiguous"])
        self.assertEqual(test_setup.full_blockers({}), ["long-suite-inventory-missing"])

    def test_only_explicitly_implemented_suites_with_commands_are_available(self):
        config = {"long_suites": [{"id": "ready", "status": "implemented", "command": ["python3", "-V"]}]}
        self.assertEqual(test_setup.full_blockers(config), [])

    def test_fast_commands_do_not_enable_external_integrations(self):
        steps = test_setup.fast_steps(Path("/example"), "all")
        self.assertTrue(steps)
        self.assertFalse(any("docker" in command or "MB_INTEGRATION=1" in command for _, command, _ in steps))
        self.assertIn("build", " ".join(arg for _, command, _ in steps for arg in command))

    def test_report_is_private_and_json(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            report = test_setup.write_report(root, {"status": "not_run"})
            self.assertEqual(report.stat().st_mode & 0o777, 0o600)
            self.assertEqual(report.parent.stat().st_mode & 0o777, 0o700)
            self.assertIn('"not_run"', report.read_text())


if __name__ == "__main__":
    unittest.main()
