# CI/CD dan deployment — fondasi Phase 0

## Status yang tidak boleh disalahartikan

PR ini mengaktifkan **CI untuk dokumen dan release policy**, serta menambahkan **jalur CD yang sengaja dormant**. Belum ada Go API/React app, Compose produksi atau hook deploy di VPS. Tidak ada build frontend/backend/container yang pura-pura hijau karena di-skip. Tes aplikasi, scan dependency/image, migration dan restore menjadi kewajiban CI pada fase masing-masing; lihat [rencana implementasi](07-implementation-plan.md).

Target akhir: browser HTTPS `menerjangbadai.online` → Caddy → loopback `127.0.0.1:2023` aplikasi. Port2023 bukan layanan publik tanpa TLS yang perlu dibuka baru. Caddy saat ini memakai port itu untuk static site; cutover belum dilakukan. Dashboard Hermes9119→9120 tidak berubah.

## CI saat ini

`.github/workflows/ci.yml` berjalan untuk PR ke main, push main dan branch agent. Job:

1. Policy tests pada Python3.12 dan3.14: owner-merged release, deny direct push/foreign actor/foreign repo/wrong SHA/failed CI; readiness marker dan scanner/link fixtures.
2. Local markdown link validation + targeted credential scan.
3. Python compile dan Bash syntax.
4. `actionlint` semantic workflow validation, checksum-pinned binary.
5. Gitleaks working tree **dan full Git history**, redacted, checksum-pinned binary.
6. `CI gate` gagal jika salah satu dependency gagal/cancelled/skipped; ini status check utama untuk branch protection.

Actions dipin SHA; checkout tidak mempertahankan credential. PR tidak memiliki production secrets, tidak memakai `pull_request_target`, tidak menjalankan persistent self-hosted runner di VPS produksi. Auto-upgrade pins harus direview; checksum pin tidak menggantikan review provenance.

## CD setelah merge owner — fail closed

`.github/workflows/deploy.yml` hanya menerima successful workflow_run dari CI push main pada repo ini. Tidak memakai artifact dari PR/fork.

1. Checkout exact SHA dari same-repo CI push main (bukan ref PR/fork); refetch run lewat API dan wajibkan status completed, workflow path `.github/workflows/ci.yml`, repository/branch/event/conclusion benar.
2. REST GitHub mengambil PR terkait commit, lalu detail PR; wajib merged=true, base main repo ini, merge_commit_sha sama SHA yang diuji, merged_by tepat `deaniscoding20` dan numeric ID `351171974`.
3. Verifikasi SHA masih head main, untuk menolak CI lama yang selesai terlambat.
4. `deploy/release.json` wajib schema valid, phase7, production_ready=true, target domain/loopback2023 tetap, dan input release benar-benar tersedia.
5. Repo variable `PRODUCTION_DEPLOY_ENABLED` wajib literal `true`; default sengaja absent/false.
6. Job masuk environment `production` dengan reviewer owner (konfigurasi repo diverifikasi melalui API, bukan berasal dari YAML saja). Environment secrets hanya diberikan setelah gate terpenuhi. Setelah approval, SHA main dicek ulang sebelum SSH; server tetap wajib melakukan check terkunci sebelum promosi.
7. SSH dengan strict verified known-hosts + dedicated non-root key mengirim **hanya SHA 40 hex** ke pre-provisioned `/usr/local/libexec/mb-deploy`.
8. HTTPS `/api/readyz` wajib menyatakan `status=ready` dan commit tepat release SHA; HTTP200 website lama bukan health aplikasi.

**Phase0 merge tidak melakukan deploy.** Fase berikutnya yang belum ready juga tidak. Marker hanyalah salah satu syarat konfigurasi, bukan bukti kesiapan. Tidak ada workflow_dispatch deploy yang melewati merge owner, dan tidak ada auto-merge oleh agent.

`user.name=MB_default_agent` hanya identitas commit. PR dibuat melalui akun GitHub owner yang mengotorisasi CLI, bukan akun MB_default_agent terpisah. Principal OAuth owner secara teknis memiliki hak owner; aturan agent tidak merge adalah batas operasional, bukan separation of duties berbasis credential. Untuk pemisahan keras, ganti dengan GitHub App/account agent khusus tanpa bypass merge setelah owner menyediakan otorisasinya. Jangan membagikan PAT owner ke runner.

## Secrets dan variables yang diperlukan nanti

Environment `production` secrets (belum dibuat nilainya):
- `VPS_HOST`, `VPS_USER` dedicated deployment principal.
- `VPS_DEPLOY_KEY` khusus forced-command deployment, bukan private key Git fetch atau credential Hermes.
- `VPS_KNOWN_HOSTS` host key VPS hasil verifikasi out-of-band.

Repo variable: `PRODUCTION_DEPLOY_ENABLED=false` sampai readiness/cutover approved. OIDC/Postgres/MinIO/Temporal/Postiz/Superset credentials disuntikkan pada runtime secret store di VPS, tidak dibawa UI atau committed YAML. Detail Auth di [security](08-authentication-security.md).

## Kontrak hook VPS — belum diimplementasikan/dipasang

Hook tidak boleh sekadar `git pull && docker compose up`. Phase7 harus mengimplementasikan dan menguji:

- Dedicated deployment user, file hook dimiliki admin dan tidak writable oleh akun itu. SSH key `restrict,command=...`; larang forwarding/PTY/general shell. Validasi SSH_ORIGINAL_COMMAND hanya path hook+SHA, semua path fixed. Jangan beri unrestricted sudo/docker-group yang ekuivalen root jika tidak diperlukan.
- Satu lock deploy, monotonic/current-main check juga **di server** sebelum switch. Client check saja masih punya TOCTOU.
- Validasi release dari trusted repo/registry, SHA exact, image digest approved, readiness report dan manifest. Jangan download mutable latest.
- Tarik/persiapan release di staging path baru; environment runtime di luar repo tidak tertimpa.
- Backup metadata-preserving Caddy/static config dan DB sebelum migration; integrity+restore rehearsal terlebih dulu. Jangan `chmod -R`, `chown -R`, reset/clean pada workspace manusia, atau menyalin .hermes.
- Expand/contract migration dengan schema compatibility; hindari destructive down migration otomatis. Simpan schema version/backup ID.
- Build/test/image scan sebelum release dipromosikan; ideal image once CI then deploy immutable digest. Implement artifact attestation/verifikasi sebelum dipakai CD.
- Boot candidate isolated bind, dependency health+smoke, lalu switch loopback2023 dan Caddy validated config pada cutover pertama yang owner approve. Jangan matikan default gateway atau site sebelum candidate valid.
- Poll bounded readiness identity di loopback dan HTTPS, sesi OIDC + smoke authenticated sandbox. Jika gagal, kembalikan container/config release sebelumnya dan verifikasi HTTPS. DB rollback hanya jika aman sesuai kompatibilitas; hentikan dan eskalasi jika irreversible, jangan menghapus data.
- Rollback juga bila public HTTPS smoke gagal; client curl gagal saja tidak otomatis memulihkan server. Hook perlu memegang transaksi promosi sampai kedua probe lulus dan menyimpan receipt audit redacted.
- Tidak mengubah :9119, firewall atau service Hermes. Admin interfaces/storage/db/workflow tidak public.

Karena hook/container/app belum ada, PR ini **tidak dapat disebut CD operasional end-to-end**. Wiring ini dapat ditinjau sekarang tanpa mengganti website. Tes rehearsal deploy/rollback/restore nyata baru bisa dilaporkan setelah implementasi dan persetujuan infrastruktur.

## Proteksi branch dan owner review

Sebelum perubahan konfigurasi, inspect current policy. Target minimal: require PR, `CI gate` strict/up-to-date, enforce admins, block force push/deletion. Jangan meminta approval self-review yang membuat owner tunggal tak bisa merge PR miliknya sendiri; review count 0 tetap membutuhkan owner review secara proses, sementara deployment secara API memeriksa merged_by owner. Jika owner ingin two-person rule, gunakan akun terpisah dengan reviewer berbeda.

Repo settings dan environment enforcement dicatat berdasarkan fresh API reads di laporan verifikasi. File YAML tidak otomatis mengaktifkan branch protection atau reviewer. Owner tetap dapat mengubah policy; tidak ada klaim bahwa identitas commit membatasi permission akun.

## Menambah gates aplikasi bertahap

Phase1 harus memperbarui CI sebelum app accepted: gofmt check, go vet, go test -race, TypeScript strict+lint+unit+production build, sqlc diff, OpenAPI lint/contract, migration empty Postgres+integration+Playwright OIDC, govulncheck/frontend dependency scan, Docker build+health+image vulnerability scan. Phase3 menambah Temporal replay/restart dan outbox/idempotency/bridge isolation. Phase5 Postiz uncertainty/reconciliation sandbox. Phase6 data source/tenant-RLS/token-expiry. Phase7 backup restore/failure injection/load budget dan deployment rollback.

Tidak menambahkan job kosong `echo tests passed`, `continue-on-error` untuk critical gate, atau tests yang diam-diam mengizinkan missing service.
