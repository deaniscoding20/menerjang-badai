"""Delivery workflow contract: no production credential transport or execution."""
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


class DeployBoundaryTest(unittest.TestCase):
    def test_actions_is_attestation_only_without_credential_transport(self):
        text = (ROOT / ".github/workflows/deploy.yml").read_text()
        for forbidden in ("secrets.", "ssh ", "scp ", "rsync ", "VPS_", "self-hosted",
                          "environment:", "scripts/deploy", "vps_release_hook", "sudo "):
            with self.subTest(forbidden=forbidden):
                self.assertNotIn(forbidden, text)
        self.assertIn("name: Release eligibility (no deployment)", text)
        self.assertIn("actions: read", text)
        self.assertIn("persist-credentials: false", text)
        self.assertIn("--owner deaniscoding20", text)
        self.assertNotIn("deployable", text)


if __name__ == "__main__":
    unittest.main()
