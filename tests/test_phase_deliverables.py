"""Required Phase0 file inventory; not a semantic acceptance/approval test."""
import tempfile
import unittest
from pathlib import Path
from scripts.check_phase_deliverables import check, REQUIRED, DIAGRAMS


class DeliverablesTest(unittest.TestCase):
    def test_empty_repository_reports_every_missing_deliverable(self):
        with tempfile.TemporaryDirectory() as folder:
            self.assertEqual(len(check(Path(folder))), len(REQUIRED))

    def test_existing_but_empty_files_do_not_pass(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            for name in REQUIRED:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.touch()
            self.assertEqual(len(check(root)), len(REQUIRED))

    def test_diagram_files_require_mermaid_block(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            for name in REQUIRED:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("# Fixture only; not real architecture\n")
            self.assertEqual(len(check(root)), len(DIAGRAMS))
            for name in DIAGRAMS:
                (root / name).write_text("```mermaid\nflowchart LR\n A --> B\n```\n")
            self.assertEqual(check(root), [])


if __name__ == "__main__":
    unittest.main()
