"""Repository checks exercise real temporary files; no credentials/network."""
import tempfile
import unittest
from pathlib import Path
from scripts.check_repository import check_links, check_secrets


class RepositoryChecksTest(unittest.TestCase):
    def test_missing_local_document_link_is_reported(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            (root / "README.md").write_text("[missing](docs/absent.md)")
            self.assertEqual(len(check_links(root)), 1)

    def test_broken_local_link_diagnostic_does_not_leak_query_token(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            fake = "ghp_" + "B" * 36
            target = "docs/absent.md?token=" + fake
            (root / "README.md").write_text("# Example\n\n[missing](" + target + ")\n")
            errors = check_links(root)
            self.assertEqual(len(errors), 1)
            self.assertNotIn(target, "\n".join(errors))
            self.assertNotIn(fake, "\n".join(errors))
            self.assertEqual(errors, ["README.md:3: broken local link"])

    def test_existing_local_link_and_external_link_are_valid(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            (root / "other.md").write_text("# Example")
            (root / "README.md").write_text("[doc](other.md#example) [web](https://example.org) [email](mailto:a@example.org)")
            self.assertEqual(check_links(root), [])

    def test_secret_detection_reports_location_not_value(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            fake = "ghp_" + "A" * 36
            (root / "bad.txt").write_text(fake)
            errors = check_secrets(root)
            self.assertEqual(len(errors), 1)
            self.assertNotIn(fake, str(errors))

    def test_private_key_and_dotenv_are_rejected(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            (root / ".env").write_text("EXAMPLE=value")
            (root / "key.txt").write_text("-----BEGIN " + "OPENSSH PRIVATE KEY-----")
            self.assertEqual(len(check_secrets(root)), 2)

    def test_placeholder_example_is_not_a_credential(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            (root / ".env.example").write_text("CLIENT_SECRET=replace-at-deploy-time")
            self.assertEqual(check_secrets(root), [])


if __name__ == "__main__":
    unittest.main()
