# Setup testing — development dulu, full testing nanti

Status: setup runner tersedia, tetapi suite panjang masih **planned/deferred**, bukan lulus. Tidak ada E2E/load/failure/restore/publication test yang dijalankan saat menambahkan setup ini.

## 1. Lihat rencana tanpa menjalankan tes

Dari root repo:

```bash
python3 scripts/test_setup.py plan
```

Menampilkan tools tersedia, inventori `testing/suites.json` dan blocker per suite. Tidak membaca credential produksi atau memulai service. Daftar suite adalah checklist executable-readiness, bukan hasil pengujian.

## 2. Prasyarat developer

- Python3.12+ (stdlib; tidak perlu pip install).
- Go sesuai `go.mod`; runner memakai `go` di PATH atau instalasi user-local `~/.local/lib/go/bin/go` bila ada.
- Node sesuai `web/package.json` dan npm; pasang dependency dengan lockfile:

```bash
npm --prefix web ci --registry=https://registry.npmjs.org
```

Jangan mengatasi masalah registry dengan `--force`/mematikan keamanan transport. Review perubahan lockfile jika perlu memperbarui dependency.

- Docker Engine tersedia dan daemon berjalan **hanya jika memilih profil database**. Tidak otomatis diinstal oleh runner, tidak memakai host PostgreSQL sebagai fallback.
- Tools full testing lain (Playwright browser+system dependencies, Authentik, Temporal, MinIO, Postiz sandbox, Superset, load/image scanners) belum dipasang/dikonfigurasi oleh setup ini. Harus disiapkan di staging terisolasi, bukan production.

## 3. Cek singkat selama development

```bash
python3 scripts/test_setup.py fast --scope backend
python3 scripts/test_setup.py fast --scope frontend
python3 scripts/test_setup.py fast --scope python
# semua cek singkat, hanya bila diperlukan:
python3 scripts/test_setup.py fast
```

- Backend: build package, unit domain/application/config. Tidak menjalankan live OIDC/DB/agent.
- Frontend: typecheck, unit tests dan production build. Bukan browser E2E.
- Python: kebijakan rilis dan runner. Tiga error lama pada `tests/test_vps_delivery.py` karena `trusted_json` belum diimplementasikan tetap harus dicatat sebagai gagal; runner tidak skip/menyamarkannya.
- Batas waktu per langkah; bukan guarantee semua dependency/build selesai pada mesin apa pun. Kegagalan missing tool/timeout ditampilkan sebagai gagal, bukan pass.
- Profil fast mengosongkan DSN integration dan menonaktifkan MB_INTEGRATION. Test skip opsional tidak boleh ditafsirkan sebagai lulus integrasi.

## 4. Database integration disposable — disiapkan, belum dieksekusi

```bash
python3 scripts/test_setup.py database --confirm-disposable-db
```

Profil ini hanya membuat **container PostgreSQL baru** dengan nama unik, password acak per run, tmpfs, batas memory/CPU, port host acak loopback, dan database `mb_test_disposable`. Image dipin digest di `suites.json` (manifest upstream diperiksa saat setup). Tidak menerima DSN host/production dari CLI.

Password hanya berada dalam temporary private file dan environment subprocess integration. Tidak masuk argument password, repo, hasil tes atau output container log. Temporary secret dibersihkan dan container dihapus di `finally`. Credential/DSN di output di-redact; audit log tambahan tetap harus tidak menulis secret.

Jika proses dibunuh SIGKILL/host crash, `finally` tidak terjamin. Periksa container label `mb.testing=disposable` secara manual dan hapus **hanya container test terkonfirmasi**; jangan menjalankan prune global. Cleanup gagal membuat run gagal.

Perintah integration yang dipanggil saat DB ready:

```bash
# Variabel DSN diisi runner, jangan salin credential ke command/chat:
# MB_INTEGRATION=1 MB_TEST_DATABASE_URL=<ephemeral DSN> go test ./internal/adapters/postgres -count=1 -timeout=90s
```

Migrations dijalankan oleh test helper terhadap DB baru. Re-run tests tidak sama dengan backup/restore/upgrade dari versi produksi; skenario itu ada di backlog.

**Keterbatasan:** Docker tidak tersedia pada host saat setup; jalur start/readiness/cleanup container belum dibuktikan end-to-end. Jangan menyebutnya integration pass sebelum perintah nyata berhasil.

## 5. Full testing — sengaja gagal tertutup saat belum siap

```bash
python3 scripts/test_setup.py full --confirm-staging
```

Saat ini exit nonzero dengan daftar suite planned. Tidak ada long test dijalankan. Mengubah status menjadi implemented hanya boleh dilakukan bersama harness/perintah nyata dan prasyarat suite, bukan untuk membuat indikator hijau.

Sebelum full diaktifkan:
1. Implementasikan suite dan isi `command` array tanpa shell interpolation/secret literals di `testing/suites.json`.
2. Review setiap command: staging-only, tidak memutasi profile/site/data produksi dan tidak publish real content.
3. Sediakan secret staging dari private files/secret store VPS; jangan memakai credential produksi sebagai default. Hanya test DB ephemeral yang memakai env DSN pendek umur.
4. Tentukan target beban, latency, RPO/RTO dan batas biaya; saat ini belum disepakati, jangan mengarang angka.
5. Selesaikan checklist [DEFERRED-CHECKLIST](../docs/testing/DEFERRED-CHECKLIST.md), lalu catat evidence per suite/commit. Tidak ada klaim production readiness hanya dari exit0 runner.

## 6. Evidence dan hasil

Runner menulis JSON bertanggal ke `artifacts/testing/` (gitignored), directory0700/file0600. Ada profile, waktu, command-step name, exit code, duration, output ter-redact, error/blocker dan `full_testing_complete` eksplisit. Jangan upload artifact raw ke repo publik; trace browser, HAR, screenshots dan DB dump bisa memuat private data.

Sukses profil fast hanya `passed_for_selected_profile_only`. `full_testing_complete=false` tetap berlaku sampai semua suite lengkap benar-benar dijalankan. Laporan tidak menjadi bukti live provider jika test memakai mock/fixture.

Lanjutkan ke [strategi](../docs/13-testing-strategy.md) dan [checklist ditunda](../docs/testing/DEFERRED-CHECKLIST.md). Development task management berjalan terpisah; setup ini tidak memicu deploy/merge/production test.

Slice tenggat/komentar menambah migration `000003_task_collaboration.sql` dan `internal/adapters/postgres/task_collaboration_test.go`. Profil database yang sama akan menjalankannya pada DB disposable; **belum dijalankan**. Rincian kasus ada di checklist T02.D1–D5 dan T02.C1–C7; UI/API fixture passes tidak mencentang acceptance DB/browser staging.
