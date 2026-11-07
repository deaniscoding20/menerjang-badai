"""Release marker guards: fixture filesystem, no production operations."""
import tempfile
import unittest
from pathlib import Path
from scripts.release_readiness import validate


def config(**changes):
    value = {"schema_version": 1, "production_ready": False, "phase": 0,
             "domain": "menerjangbadai.online", "bind": "127.0.0.1", "port": 2023}
    value.update(changes)
    return value


class ReadinessTest(unittest.TestCase):
    def test_phase_zero_is_valid_but_not_deployable(self):
        with tempfile.TemporaryDirectory() as folder:
            self.assertFalse(validate(config(), Path(folder)))

    def test_early_phase_cannot_enable_production(self):
        with tempfile.TemporaryDirectory() as folder:
            with self.assertRaises(ValueError):
                validate(config(production_ready=True), Path(folder))

    def test_unknown_or_unsafe_deployment_target_is_rejected(self):
        with tempfile.TemporaryDirectory() as folder:
            for change in [{"port": 80}, {"bind": "0.0.0.0"}, {"domain": "elsewhere.test"}, {"production_ready": "false"}, {"schema_version": 2}]:
                with self.subTest(change=change), self.assertRaises(ValueError):
                    validate(config(**change), Path(folder))

    def test_phase_seven_requires_actual_release_artifacts(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            with self.assertRaises(ValueError):
                validate(config(phase=7, production_ready=True), root)
            for relative in ["go.mod", "web/package-lock.json", "deploy/compose.production.yaml", "docs/reports/production-readiness.md", "scripts/deploy.sh"]:
                target = root / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("test fixture")
            self.assertTrue(validate(config(phase=7, production_ready=True), root))


if __name__ == "__main__":
    unittest.main()
