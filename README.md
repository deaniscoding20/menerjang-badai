# MenerjangBadai — AI Marketing Office

Pusat operasi marketing multi-brand untuk manusia dan enam profil Hermes yang sudah ada: Claudia, Jordan, Kobe, Curry, Sabrina, dan Kyrie.

**Status: development Phase 1–3; belum production-ready dan belum di-deploy.** Source fondasi Go/React, OIDC/session, ACL brand, direktori enam agen, dan task management sedang dibangun. Dispatch Hermes, durable Temporal workflow dan integrasi fase berikutnya belum diaktifkan. Baca source serta hasil cek aktual; dokumen arsitektur bukan klaim fitur live.

## Stack target sesuai spesifikasi

- Go modular monolith, Hexagonal Architecture (Ports and Adapters).
- React + TypeScript strict.
- PostgreSQL + sqlc/migrations sebagai business source of truth.
- Temporal untuk durable workflow; transactional outbox untuk dispatch kritis.
- Authentik OIDC untuk human login; opaque server sessions, CSRF dan org/brand ACL.
- MinIO untuk assets/documents; Postiz untuk publishing approved; Superset untuk BI.
- Redis hanya bila kebutuhan cache/transient coordination terbukti.
- Docker Compose dan Caddy; target HTTPS `menerjangbadai.online` dengan backend loopback2023.

Tidak membuat ulang persona agent, menganggap bot Telegram sebagai workflow API, atau menampilkan metric/connection palsu.

## Baca dulu

**Snapshot pekerjaan terbaru, hasil cek aktual dan bug terbuka:** [Development status](docs/DEVELOPMENT-STATUS.md). Branch ini adalah pekerjaan berjalan, bukan Phase 3 selesai atau rilis produksi.

1. [Project overview dan status](docs/00-project-overview.md)
2. [Discovery VPS, repo, jaringan dan blocker](docs/01-environment-discovery.md)
3. [Architecture](docs/02-architecture-overview.md), [hexagonal boundaries](docs/03-hexagonal-architecture.md), [layout](docs/04-project-structure.md)
4. [Domain](docs/05-domain-model.md) dan [database/ERD](docs/06-database-design.md)
5. [Rencana Phase 0–7 dan acceptance gates](docs/07-implementation-plan.md)
6. [Authentication/security](docs/08-authentication-security.md), [Hermes](docs/09-hermes-integration.md), [Postiz](docs/10-postiz-integration.md)
7. [CI/CD dan cutover/rollback contract](docs/11-ci-cd-and-deployment.md)
8. [Task management development subset](docs/12-task-development.md) dan [startup lokal](docs/14-local-development.md)
9. [Testing strategy](docs/13-testing-strategy.md) dan [acceptance cases tertunda](docs/testing/DEFERRED-CHECKLIST.md)

Spesifikasi fase asli tetap ada di root `PHASE-00-…` sampai `PHASE-07-…` dan tidak diubah oleh PR blueprint.

## Setup dan catatan testing (development-first)

Panduan yang dipakai selama development ada di [testing/README.md](testing/README.md): prasyarat tools, perintah fast/backend/frontend, database disposable, evidence, dan inventori tes panjang yang masih ditunda.

```bash
python3 scripts/test_setup.py plan                  # inventori; tidak menjalankan tes
python3 scripts/test_setup.py fast --scope backend # cek singkat sesuai fitur
```

E2E, load, failure injection, restore/rollback dan integrasi live **tidak dianggap lulus hanya karena ditunda**. Profil `full` sengaja menolak eksekusi sampai suite/prasyarat nyata tersedia. Setup tidak memasang service atau menguji production otomatis. Evidence lokal berada di `artifacts/testing/` (private dan gitignored).

## Pemeriksaan lokal individual

Python3.12+ tanpa dependency pihak ketiga untuk policy tests:

```bash
python3 -m unittest discover -s tests -v
python3 scripts/check_repository.py
python3 scripts/release_readiness.py
python3 -m compileall -q scripts tests
for script in scripts/*.sh; do bash -n "$script"; done
git diff --check
```

Linux amd64: tool validasi user-local dengan versi/checksum pin:

```bash
bash scripts/install-ci-tools.sh
"$HOME/.local/bin/actionlint" -color
"$HOME/.local/bin/gitleaks" dir --redact --no-banner .
"$HOME/.local/bin/gitleaks" git --redact --no-banner --log-opts=--all .
```

Installer ini hanya dua tool CI di `~/.local/bin`, bukan provisioning VPS/Docker. Review sebelum eksekusi. `release_readiness.py` sukses pada Phase0 berarti marker valid **dan deployable=false**, bukan production ready.

CI yang tersedia berfokus pada matrix Python3.12/3.14, policy tests, local links, targeted scan, compile/Bash syntax, Actions lint dan Gitleaks. Unit/fixture Go+React serta database opt-in tests sekarang ada, tetapi wiring lengkap CI aplikasi/DB/browser masih pending; keberadaan file bukan hasil GitHub Actions run.

## Pull request, identitas dan deployment

- Author commit: `MB_default_agent`; PR autentikasi memakai akun owner yang mengotorisasi CLI, bukan akun agent baru.
- Izin owner terbaru mengizinkan agent merge dan deployment pertama setelah aplikasi siap; izin ini menggantikan batas PR-only awal, bukan meniadakan CI/security/readiness gates.
- Main memerlukan PR dan `CI gate`; force push/delete diblokir dan admins ikut enforcement.
- CD hanya eligible setelah successful main push CI pada SHA hasil PR yang di-merge `deaniscoding20`, serta SHA masih main terkini.
- **Produksi tetap disabled:** readiness marker dan gate CI/CD belum memenuhi release. Ada source aplikasi dan binary lokal; belum ada dashboard deployment atau container produksi yang dibuktikan. Credential GitHub tetap di VPS di luar repo, tidak dibawa ke runner.
- Merge PR blueprint tidak mengganti website. Hook/Compose/cutover baru diimplementasikan dan diuji setelah fase terkait siap serta owner menyetujui.
- Website lama pada2023 dan dashboard Hermes9119 tidak diubah.

Jangan commit `.env`, token, SSH private keys, cookie, credentials provider, atau state/transcript Hermes. Jangan membawa kredensial owner ke runner. Batas operasional PR-only tidak setara pemisahan permission akun; hard isolation memerlukan akun/App agent tersendiri.
