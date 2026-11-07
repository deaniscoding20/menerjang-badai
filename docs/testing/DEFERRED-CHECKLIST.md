# Checklist pengujian tertunda — Phase 01–07

**Status awal seluruh checkbox: deferred / belum dijalankan, bukan pass.** Pengembangan didahulukan. Dokumen ini tidak mencatat hasil tes; tidak ada tes, build, Docker, login, jaringan, publikasi, atau operasi produksi yang dijalankan dalam penyusunannya.

Acuan: [strategi dan inventaris berkas](../13-testing-strategy.md), serta acceptance criteria dan scope [Phase 1](../../PHASE-01-APPLICATION-FOUNDATION.md), [Phase 2](../../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md), [Phase 3](../../PHASE-03-HERMES-INTEGRATION-AND-WORKFLOWS.md), [Phase 4](../../PHASE-04-KNOWLEDGE-ASSETS-AND-CONTENT.md), [Phase 5](../../PHASE-05-APPROVAL-AND-POSTIZ-PUBLISHING.md), [Phase 6](../../PHASE-06-ANALYTICS-AND-MARKETING-INTELLIGENCE.md), dan [Phase 7](../../PHASE-07-PRODUCTION-HARDENING-AND-OPERATIONS.md). ID induk kasus mengikuti matriks strategi §5. Cakupan lokal adalah snapshot inventaris 6 November 2026; source dan test baru dapat bertambah selama pengembangan paralel.

## Cara memakai register

- `[ ]` berarti belum diterima: deferred, blocked, fail, not_run atau skipped. Tambahkan status+blocker aman bila perlu; **jangan centang karena berkas tes ada, fixture lulus, konfigurasi tersedia, atau service di-skip**.
- `[x]` baru boleh dipakai setelah kasus benar-benar dieksekusi pada revisi/lingkungan tercatat, assertion dan exit code ditinjau, serta evidence teredaksi diterima reviewer. Jangan mencentang kasus lain dalam kelompok secara otomatis.
- Bukti hanya di `artifacts/testing/<run-id>/`, gitignored melalui `artifacts/`. Direktori private **0700**, file **0600**; tidak ada secret, raw DSN, token, cookie, presigned query, PII atau raw HAR/browser storage state dalam results maupun Git. Di checklist tracked cukup ID run/suite, hasil ringkas aman dan reviewer; bukti private tidak perlu dipublikasi.
- Owner di bawah adalah **peran usulan**. Owner proyek menetapkan nama, target tanggal, prioritas dan reviewer; semuanya masih pending. Satu blocker harus punya next action, bukan dihapus dengan bypass/credential produksi.
- Live integration berarti layanan staging/sandbox versi aktual, bukan `httptest`, frontend fixture, synthetic GitHub metadata, atau dokumentasi vendor. `provider_test.go` adalah **signed mocked OIDC lokal, bukan Authentik live**.
- `full` harus fail-closed sampai seluruh long gate terimplementasi dan preflight sah. Bahkan full nonpublik yang berhasil bukan bukti social publishing nyata atau readiness produksi.

## 0. Preflight, runner, dan evidence

**Owner usulan:** maintainer runner + DevOps + reviewer keamanan. **Berkas tersedia:** `tests/test_test_setup.py`; `scripts/test_setup.py` tersedia dan enam unit harness pernah lulus. Database/full acceptance dan isolasi runtime masih memerlukan review prerequisites; hasil harness bukan hasil full testing.

### Scope, alat dan isolasi

- [ ] PF01 — Tetapkan operator, reviewer, revisi/build identity dan daftar kasus; source yang masih berubah tidak memakai hasil revisi lain sebagai bukti.
- [ ] PF02 — Setujui waktu, batas resource, stop condition, dan fault yang diizinkan; tidak ada izin gangguan implisit.
- [ ] PF03 — Verifikasi Go sesuai `go.mod`, Python yang didukung runner, Node sesuai engines, npm/lockfile dan dependency lokal; catat versi aktual tanpa install/download diam-diam.
- [ ] PF04 — Untuk DB, verifikasi Docker daemon yang diotorisasi, image PostgreSQL yang dipin/disetujui, kapasitas serta resource disposable; bukan daemon/volume produksi.
- [ ] PF05 — Untuk browser/full, siapkan staging TLS/CA, redirect URI, browser Playwright/binary yang sesuai dan alat gate yang benar-benar dipakai; dependency Playwright bukan suite E2E.
- [ ] PF06 — Pisahkan domain, database, MinIO bucket/prefix, Temporal namespace/task queue, cache bila dipakai, account/channel, dan sink telemetry/notifikasi staging.
- [ ] PF07 — Pastikan tidak memakai situs lama, listener aktif, data produksi, dashboard Hermes aktif, atau profile/home/session/storage Hermes lama sebagai fixture atau target failure test.
- [ ] PF08 — Tetapkan allowlist sandbox dan blokir egress ke akun nyata, publishing, spend, Telegram/email pengguna nyata; hentikan full jika batas target tidak bisa dibuktikan.
- [ ] PF09 — Siapkan fixture sintetis: dua organisasi, dua brand dalam satu organisasi, seluruh role, user tanpa grant, revoked/expired session, versi dokumen/asset/content berbeda; tidak memakai PII/backup produksi.
- [ ] PF10 — Admin terkait menyediakan credential reference melalui secret manager/file private di luar repo, scope minimum dan cleanup/revoke plan; tidak meminta credential dalam chat atau memakai admin produksi.

### Kontrak runner yang harus diuji kemudian

- [ ] RUN01 — `python3 scripts/test_setup.py plan` hanya menghasilkan rencana/prasyarat/blocker; tidak mulai service, Docker, login, instalasi atau suite.
- [ ] RUN02 — `python3 scripts/test_setup.py fast` hanya suite pendek unit/lint/typecheck/build; tidak memanggil DB, Docker, browser live, failure injection atau load test.
- [ ] RUN03 — Fast menghapus/menolak `MB_TEST_DATABASE_URL`/`MB_INTEGRATION` dan credential integrasi warisan; default Go `Skip` tidak dicatat sebagai database pass.
- [ ] RUN04 — `python3 scripts/test_setup.py database` hanya membuat PostgreSQL Docker disposable dengan credential acak sementara; menolak seluruh DSN eksternal/manual, bukan menerima staging/produksi.
- [ ] RUN05 — Database mode menyuntikkan DSN+`MB_INTEGRATION=1` hanya ke child test, bind loopback dan identifikasi container milik run sebelum migrasi; DSN/credential tidak ke log/results/argv/Git.
- [ ] RUN06 — DB tidak tersedia/gagal provision/migrasi/test menghasilkan fail/blocked, bukan skip sukses; timeout dan exit code nyata tercatat.
- [ ] RUN07 — Cleanup hanya container/network/volume milik run; interrupt/timeout tidak meninggalkan credential/resource atau menghapus resource lain; tidak `prune`.
- [ ] RUN08 — `python3 scripts/test_setup.py full` perlu pemilihan eksplisit dan staging preflight, tidak terpicu environment flag atau fallback ke fast.
- [ ] RUN09 — Full menolak long-suite inventory kosong, missing/unknown/planned/deferred/blocked status, atau command hilang; tulisan `passed` dalam config tidak membuka gate implemented.
- [ ] RUN10 — Setiap long gate implemented punya command nyata, ID kasus, fixture, owner, timeout dan evidence; seluruh command benar-benar berjalan, missing/skipped mandatory gate memblokir kesimpulan full selesai.
- [ ] RUN11 — Long tests tetap dinonaktifkan sampai implementasi lengkap; bukan memalsukan implemented untuk melewati fail-closed.

### Evidence dan hasil

- [ ] EV01 — `artifacts/testing` dan folder run milik operator, private 0700; file 0600; symlink/ownership diperiksa dan `artifacts/` tetap gitignored.
- [ ] EV02 — Redaction dijalankan sebelum stdout/stderr disimpan/ditampilkan; uji dengan synthetic secret/DSN/header/cookie/token/presigned query, tanpa credential nyata.
- [ ] EV03 — Report memuat mode, revisi, waktu, versi alat/image, target label nonsecret, suite/case ID, status terpisah pass/fail/blocked/not_run/skipped, exit code dan durasi.
- [ ] EV04 — Credential ephemeral hanya di memory/secret storage terpisah; tidak ada raw env dump, shell trace, HAR/token screenshot, browser storage state, plaintext backup atau secret di evidence.
- [ ] EV05 — Setujui retention, reviewer access, cleanup dan kebijakan upload hasil; tidak upload artifact sensitif otomatis ke CI/public report.
- [ ] EV06 — Kebocoran menghentikan distribusi, dieskalasi ke owner keamanan dan ditangani revoke/rotate terotorisasi; redaction retroaktif saja bukan pemulihan.

## 1. Foundation — auth/session/CSRF/org+brand

**Owner usulan:** backend/auth + frontend + security; admin Authentik untuk live. **Tersedia parsial:** `internal/application/auth_test.go`, `internal/adapters/oidc/provider_test.go`, `internal/adapters/httpapi/server_test.go`, `internal/platform/config_test.go`, `internal/application/brands_test.go`, `internal/adapters/postgres/{auth,store}_test.go`, `internal/domain/brand_test.go`, dan frontend fixture tests. **Belum dibuktikan:** live Authentik/browser, seluruh RBAC/mutasi, runtime/audit/migration hardening. Sumber: Phase 1 §2 dan §5; Phase 2 §2–3.

### OIDC — A01

- [ ] A01.1 — Permitted human user login lewat **Authentik staging live**, identity dipetakan dengan issuer+subject dan session aplikasi dibuat; tidak menyimpulkan dari signed local OIDC.
- [ ] A01.2 — Authorization Code+PKCE S256, state, nonce dan verifier terikat browser/login transaction; callback tanpa/salah binding ditolak.
- [ ] A01.3 — Replay state/code, state expired/missing/mismatch dan callback error ditolak tanpa session baru.
- [ ] A01.4 — Issuer, audience, nonce, signature, expiry dan required claims yang salah ditolak; JWKS/key rotation/error tidak fail-open.
- [ ] A01.5 — Identity belum diprovision, subject dari issuer lain, atau email sama tidak memperoleh membership/privilege otomatis.
- [ ] A01.6 — MFA/step-up policy dan keterbatasan logout provider dibuktikan terhadap konfigurasi Authentik aktual, bukan diasumsikan dari adapter.

### Session — A02

- [ ] A02.1 — Cookie opaque Secure/HttpOnly/SameSite/Path benar pada HTTPS; tidak ada provider token/client secret ke React/localStorage.
- [ ] A02.2 — Session ID dirotasi sesudah login, ID lama tidak tetap valid dan fixation ditolak.
- [ ] A02.3 — Privilege/membership berubah memicu rotasi/revocation sesuai policy; stale session tidak mempertahankan akses lama.
- [ ] A02.4 — Idle expiry benar dan tidak melampaui absolute expiry; kedua expiry diuji terpisah dengan waktu terkendali.
- [ ] A02.5 — Logout/server-side revoke efektif pada request berikutnya dan browser lain yang memakai session sama; provider logout limitation terlihat.
- [ ] A02.6 — Sensitive operation memerlukan reauthentication sesuai policy yang disepakati; policy belum ada tetap blocker, bukan skip.

### CSRF, API dan browser — A03

- [ ] A03.1 — Seluruh protected backend request tanpa session/expired/revoked session ditolak, bukan hanya halaman UI disembunyikan.
- [ ] A03.2 — **Setiap mutasi** menolak missing/wrong CSRF dan origin asing/kosong sesuai policy; session/CSRF silang tidak diterima.
- [ ] A03.3 — CORS/cookie/origin policy diuji dari browser; tidak membolehkan credential lintas origin sembarang.
- [ ] A03.4 — 401/403/404 dan validation/error schema sesuai kontrak; no HTML fallback pada API dan no cross-tenant data dalam detail error.
- [ ] A03.5 — Frontend session-expired/access-denied/loading/empty/error state benar; fetch fixture dicatat sebagai fixture, bukan login backend.
- [ ] A03.6 — Browser tidak menyimpan bearer/long-lived secret; XSS/unsafe URL/rendered content tidak mencuri session atau melewati same-origin boundary.

### Organisasi, brand dan role — A04

- [ ] A04.1 — Role × list/read/create/update/archive brand memiliki explicit allow/deny di backend; unknown role default-deny.
- [ ] A04.2 — Cross-org list/detail/write dan forged organization/brand ID ditolak; error/count tidak membocorkan keberadaan data asing.
- [ ] A04.3 — Same-org user tanpa explicit brand grant ditolak termasuk owner jika policy membutuhkan grant; grant di brand A tidak memberi brand B.
- [ ] A04.4 — Membership/invitation/role inheritance assignment hanya oleh principal berwenang; privilege escalation/self-grant ditolak dan diaudit.
- [ ] A04.5 — Isolation berlaku pada list/count/search/filter/export/related resource dan worker/context, bukan frontend filtering saja.
- [ ] A04.6 — Revoked membership langsung membatasi penggunaan data/worker/cache sesuai policy; ID dari client tidak dipercaya.

### Database, runtime dan audit — A05/A06

- [ ] A05.1 — Semua migrasi berjalan dari PostgreSQL disposable kosong; constraints dan indeks yang diperlukan benar-benar dibuat.
- [ ] A05.2 — Apply ulang aman; checksum drift ditolak; concurrent migration/failed migration tidak meninggalkan schema setengah jadi.
- [ ] A05.3 — Role aplikasi least privilege tidak dapat melakukan DDL/admin; migration role terpisah dan tidak bocor ke aplikasi/frontend.
- [ ] A05.4 — Parameterized SQL dan input adversarial tidak melewati filter organisasi/brand atau mengubah query.
- [ ] A06.1 — API dan frontend start dengan perintah runtime terdokumentasi; graceful shutdown menutup request/koneksi dengan benar.
- [ ] A06.2 — Liveness berbeda dari readiness; kegagalan DB/OIDC/dependency menghasilkan readiness tidak siap, bukan selalu 200.
- [ ] A06.3 — Readiness mencantumkan exact build/release identity yang diuji; website lama 200 tidak diterima sebagai health aplikasi.
- [ ] A06.4 — Request/correlation ID dan centralized errors konsisten; OpenAPI benar-benar cocok dengan response/auth/error aktual.
- [ ] A06.5 — Audit mencatat login/logout/revoke, invitation/role/permission/config dan aksi penting dengan actor+scope tanpa secret/token/cookie.
- [ ] A06.6 — Static serving hanya UI/asset yang diizinkan; file arbitrary/path traversal dan API-to-SPA fallback ditolak.
- [ ] A06.7 — Secret-pattern/history/dependency/config review selesai tanpa nilai secret dalam laporan; “tidak ada secret” tidak diklaim dari pembacaan docs saja.
- [ ] A06.8 — Formatter/linter/unit/build yang berlaku dieksekusi dengan hasil nyata; hasil revisi lama tidak dipakai sebagai acceptance saat ini.

## 2. Brand dan registry enam agen — B01

**Owner usulan:** backend brand/registry + frontend; owner agen untuk identity/sandbox. **Tersedia parsial:** brand/domain/application/PG/frontend tests di strategi §2. **Gap:** membership UI/API, seed/status evidence dan live behavior. Sumber: Phase 2 §2–9.

- [ ] B01.1 — Admin berizin dapat CRUD/archive brand dan mengelola member; unauthorized update/member assignment ditolak backend.
- [ ] B01.2 — Profile/metadata/logo/website brand tervalidasi dan tetap scoped ke organisasi+brand.
- [ ] B01.3 — Keenam agen tampil dengan nama, role dan Telegram tepat: Claudia, Jordan, Kobe, Curry, Sabrina, Kyrie; periksa **semua**, bukan sebagian assertion.
- [ ] B01.4 — Application agent ID stabil, unik dan independen dari Telegram/Hermes internal ID; seed/config apply ulang idempotent.
- [ ] B01.5 — Registry tidak mengubah prompt/persona/profile/config lama; validasi tanpa menulis home/profile agen aktif.
- [ ] B01.6 — Disabled agent tidak menerima dispatch; enabled tanpa bukti koneksi tidak diberi label online/ready.
- [ ] B01.7 — Configured/reachable/ready/busy/unavailable/error/unknown mengikuti sinyal terdokumentasi dan staleness; bot Telegram ada bukan heartbeat.
- [ ] B01.8 — Directory/detail/status error/empty state jujur; diagram kolaborasi tidak memaksa semua task lewat Claudia.
- [ ] B01.9 — Dokumentasi model brand/registry dan aturan worker authorization cocok dengan schema/source yang kelak diuji.

## 3. Tasks, agent dispatch, idempotency dan Temporal

**Owner usulan:** backend tasks + maintainer integrasi + workflow/DevOps; owner agen menyediakan sandbox dan izin. **Tersedia saat pembacaan:** `internal/domain/tasks_test.go` kontrak lifecycle pengembangan/disabled execution; **bukan** suite dispatch/live Temporal. Source/task tests lain dapat ditambahkan paralel; jangan menandai full lifecycle implemented tanpa membaca dan menjalankannya. Sumber: Phase 3 §2–7 dan §10.

### Task/transitions/assignment/deliverable — T01/T02

- [ ] T01.1 — Create/edit/list/detail task authorized, field domain valid dan terikat requester/org/brand; forged brand/task ID ditolak.
- [ ] T01.2 — Draft → ready → assigned/cancelled subset pengembangan benar; transition eksekusi yang masih disabled ditolak jujur.
- [ ] T01.3 — Seluruh lifecycle aktif, termasuk queued/running/waiting input/waiting approval/completed/failed/cancelled, punya matriks transisi allow/deny eksplisit.
- [ ] T01.4 — Stale version/concurrent edit/transition menghasilkan conflict dan tidak overwrite state/history; terminal task tidak dibuka ulang tanpa aturan sah.
- [ ] T01.5 — History/audit actor, alasan, timestamp dan version sesuai setiap aksi; failure terlihat di API/UI, bukan diam-diam selesai.
- [ ] T02.1 — Assignment langsung ke **setiap** configured specialist didukung tanpa wajib Claudia; disabled/unready destination tetap ditolak.
- [ ] T02.2 — Reassignment, priority/deadline, dependency selesai/belum selesai/cycle, comments dan clarification memiliki validasi/permission yang benar.
- [ ] T02.3 — Deliverable submission/revision/cancel authorized, raw output dan validated representation dibedakan; malformed/schema-version tidak dikenal bukan structured success.
- [ ] T02.4 — Deliverable versions/source references immutable dan tersimpan; correlation ID/external execution ID cocok task+attempt+brand.
- [ ] T02.5 — Cancellation internal dibedakan dari external cancel acknowledgement; unsupported external cancellation dinyatakan, bukan dipalsukan.

### Dependency graph — rincian T02.2

Owner: backend/workflow maintainer; scope database disposable dan browser staging, bukan website/agent produksi. Setup: [testing runner](../../testing/README.md), [task contract](../12-task-development.md); `python3 scripts/test_setup.py database --confirm-disposable-db` hanya jika prerequisites runner terpenuhi. Target DB test `TestRealTaskDependencyPersistenceCycleAndAudit` sudah tersedia tetapi **belum dieksekusi**. E2E graph belum punya harness; langkah berikutnya siapkan fixtures same-brand/cross-brand/revoked grant dan dua session browser di staging. Bukti: migration checksums, SQL snapshot, HTTP status/version, history/audit count, hasil concurrent writes dan cleanup tanpa credential.

- [ ] T02.G1 — Migrasi 000001→000004 baru/upgrade disposable berhasil; composite FK menolak edge lintas org/brand dan self, index/allowlist history+audit benar, data lama dipertahankan.
- [ ] T02.G2 — GET dependencies mengembalikan title/status dan source version dari satu SQL snapshot; empty array sah, no session/expired/revoked grant/cross-org/cross-brand tidak bocor.
- [ ] T02.G3 — POST/DELETE wajib current write role, same-origin CSRF, valid dependency_id/current version; claimed org/user/role/unknown fields ditolak. Draft/ready/assigned boleh; queued/running/waiting/terminal ditolak tanpa lifecycle change.
- [ ] T02.G4 — Add/remove atomic edge+version+history+audit; duplicate add/missing remove current version no-op; stale retry tetap409, no auto-retry mutation UI, rollback tidak meninggalkan edge atau event parsial.
- [ ] T02.G5 — Self, direct/transitive cycle ditolak; concurrent A→B/B→A dan longer opposing graph writes menghasilkan satu consistent acyclic graph setelah restart. Semua graph writers wajib memegang lock org+brand; uji deadlock/timeouts bersama metadata edit/grant revocation di isolated DB.
- [ ] T02.G6 — UI memperlihatkan prerequisite completed vs belum selesai, escaped title/status, same-brand candidate list (maks500), refresh sesudah mutation/conflict dan loaded snapshot version; readonly/terminal tidak menampilkan controls. Error API tidak disamarkan sebagai graph kosong. Refresh dependency saat brief dirty tidak membuang edit; edit mulai ketika mutasi dependency pending tidak hilang saat response datang. Koordinasikan refresh antar-panel dengan optimistic edit snapshot tetap asli.
- [ ] T02.G7 — Dependency metadata tidak mengeksekusi agen atau menandai task running/completed; gate dispatch dan pembukaan saat prasyarat selesai tetap **belum diimplementasikan**, verifikasi setelah Temporal/outbox terpasang.

### Human-imported versioned deliverables — rincian T02.3/T02.4

Owner: backend/workflow maintainer + security reviewer; fixtures sintetis di disposable DB, bukan output privat/credential agent produksi. Setup runner database di atas, test `TestRealVersionedDeliverablesPreserveRawAndRejectStale` tersedia **belum dieksekusi**. Browser import/viewer source sudah tersedia dan fixtures lulus; browser acceptance belum dijalankan. Status short checks: UTF8 raw request dan response contract sudah diperbaiki serta review ulang bersih; koordinasi dirty form terintegrasi dengan198 frontend fixtures/typecheck/lint/build lulus, review ulang koordinasi bersih (passed=true). Next action: jalankan upgrade000005 serta fixture ACL/version saat environment disposable siap, lalu browser acceptance. Bukti: raw-byte equality, revision list/detail, task version/status, audit/history, rejected SQL update/delete dan cleanup.

- [ ] T02.V1 — Upgrade database000001→000005 mempertahankan data lama; unique revision/FK task-author/immutable UPDATE+DELETE trigger berfungsi, permission DB runtime tidak bypass immutability.
- [ ] T02.V2 — Import raw1–128KiB nonblank UTF8 tanpa literal NUL dipertahankan persis; unknown schema/malformed JSON/arbitrary text/escaped NUL raw-only invalid structured null; literal backslash-u0000 bukan NUL dan envelope sah tetap validated.
- [ ] T02.V3 — Envelope v1 tujuh kind, title/sections/source references tervalidasi. Validation bukan factual approval atau agent provenance; no secret/presigned query/embedded credentials di sources. Data dengan HTML/script dirender escaped plain text, tidak dieksekusi. Kontrak response client cocok penerimaan Go: empty-query separator `https://example.com/?`, U+FEFF nonblank dan U+0085 whitespace tidak boleh membuat import tersimpan menjadi false-failure502; nonempty query/credentials/NUL tetap ditolak dari struktur sah.
- [ ] T02.V4 — Current role+session+CSRF+brand ACL berlaku untuk submit/list/detail; forbidden role/forged author/org/origin/agent/cross-parent record/cross-org dan revoked grant tidak menulis atau bocor. Read-only roles tetap boleh membaca authorized revisions.
- [ ] T02.V5 — Task-row lock menserialisasi revision allocation; concurrent imports memakai versi task sama hanya satu berhasil. New revision+task version+history+audit atomic, rollback tidak parsial; stale retries409 tanpa auto-retry atau lifecycle/approval/dispatch change.
- [ ] T02.V6 — GET latest100 summary revisions descending tidak membawa raw bodies; version+list snapshot coherent, detail old revision tetap identik setelah import baru, error tak disamarkan jadi empty success. Pagination/cursor untuk revisi lebih lama belum tersedia.
- [ ] T02.V7 — Body HTTP1MiB dan decoded raw128KiB diuji termasuk JSON escape overhead, supplementary Unicode dan oversized/invalid UTF8; limit besar tidak memperlebar endpoint lain. Literal invalid UTF8 pada raw HTTP request wajib ditolak sebelum decoder mengganti dengan U+FFFD; U+FFFD yang memang sah dan emoji tetap utuh. Raw diperlakukan data privat, tidak dimasukkan log/error/telemetry. Draft output impor yang belum terkirim tidak hilang saat dependency/brief/parent refresh; import/refresh deliverable tidak menghapus brief dirty.
- [ ] T02.V8 — Origin human_import jelas; import tidak dianggap output agent, completed, approved, revision-request response atau publish-ready. Authenticated agent result ingestion/correlation serta approval bind exact version diuji terpisah setelah workflow tersedia.

### Human revision requests — rincian T02.3

Owner: backend/workflow + frontend + reviewer security. Scope fixtures di disposable DB/browser staging. Domain/application/HTTP short tests lulus; bounded review menemukan collision route deliverable detail dengan suffix revision-requests, fix+regression terintegrasi, parent focused HTTP checks/build/diffcheck lulus; bounded rereview selesai passed=true tanpa security/logic findings; collision tertutup pada source+HTTP fixtures. Persistence migration000006 dan UI sudah terintegrasi; parent259 frontend fixtures/typecheck/lint/build dan short Go/build/vet lulus, Postgres compile-only saja. Independent persistence/UI review berjalan. SQL upgrade000006 dan browser belum dijalankan; deferred fixtures `TestRealRevisionRequestsBindImmutableVersionAndAudit` dan `TestRealRevisionRequestListLatestHundredHasDeterministicTies` tersedia. Tidak ada dispatch/approval/reopen implied. Next action: jalankan deferred SQL fixture setelah implementasi dan prerequisites tersedia, lalu browser race/ACL acceptance. Bukti: exact target ID+revision, immutable raw output, parent version/status, history/audit counts, rejected changes dan cleanup tanpa credential.

- [ ] T02.R1 — Migrasi000001→000006 baru/upgrade mempertahankan data; composite FK menolak target beda org/task/revision dan row update/delete ditolak immutable trigger.
- [ ] T02.R2 — GET latest100 revision-request records+parent version satu snapshot, ordering created_at DESC,id DESC; authorized empty list sah, malformed/error bukan empty fallback, overflow/pagination limitation ditampilkan.
- [ ] T02.R3 — POST wajib current membership/write role/explicit brand grant/opaque session/exact Origin+CSRF; forged org/author/revision/role/unknown fields, revoked grants, cross-parent/cross-org/cross-brand ditolak tanpa write/leak.
- [ ] T02.R4 — Versi task terbaru dan state draft/ready/assigned wajib; stale/concurrent double submit hanya satu commit; lifecycle execution/terminal ditolak dan tidak reopen/dispatch/resume/approval.
- [ ] T02.R5 — Catatan trim GoWhiteSpace1–4000codepoints UTF8/noNUL termasuk4000emoji/48KiB escapedJSON, FEFF nonblank/NEL blank, invalidliteralUTF8/oversize/4001emoji ditolak tanpa perubahan; raw text tidak ada di log/error/audit reason.
- [ ] T02.R6 — Target immutable deliverable valid/raw-invalid/historical dari parent yang sama tersimpan exact ID+revision; create note+task version+event+audit atomic, rollback tidak menyisakan partial rows; output lama tidak berubah.
- [ ] T02.R7 — UI escaped plain body/author/time/target revision; readonly/terminal tidak memiliki form, CSRF dan optimistic snapshot asli dipakai,409 tidak auto-retry dan body/target draft tetap ada sampai explicit action.
- [ ] T02.R8 — Refresh/mutation revision panel tidak menghapus brief/raw-import/revision draft; same-turn edit/refresh dan in-flight completion dijaga, success201 memperbarui versi/history tanpa fake agent result.

### Deadline dan catatan manusia — rincian T02.2

Owner usulan: backend tasks + frontend + reviewer DB. Source `000003_task_collaboration.sql`, `task_collaboration*.go`, `TaskComments.tsx`, dan deadline form tersedia; unit/HTTP/UI fixture pendek bukan bukti acceptance live. Semua kasus tambahan di bawah **belum dijalankan pada DB/browser staging**.

- [ ] T02.D1 — Migrasi 000001→000002→000003 pada database disposable baru dan upgrade dari snapshot task lama menghasilkan nullable deadline tanpa kehilangan row/history; migration checksum dipertahankan.
- [ ] T02.D2 — Create/read/list deadline RFC3339 offset berbeda merepresentasikan instant yang sama setelah persist PostgreSQL; null dan omitted create aman, invalid/zero/year-out-of-range ditolak.
- [ ] T02.D3 — PATCH tanpa field deadline mempertahankan nilai; null menghapus; valid string mengganti; stale version/concurrent patch tidak menghilangkan perubahan lain atau history/audit. Task ready dengan agent_id terisi tidak boleh berubah menjadi assigned hanya karena edit/clear tenggat atau brief; agent_id tidak berubah berarti tidak ada efek assignment baru.
- [ ] T02.D4 — Browser dengan zona waktu berbeda, pergantian hari dan DST diuji; edit brief tanpa mengubah tenggat mempertahankan instant termasuk detik/milidetik. Input lokal ambigu/tidak ada saat DST punya perilaku yang didokumentasikan.
- [ ] T02.D5 — Deadline lampau/future hanya metadata: tidak ada reminder, queued state, worker call, approval, spend atau publish otomatis.
- [ ] T02.C1 — Append comment/clarification persisten, body trim 1–4000 codepoints (termasuk4000 emoji supplementary-plane dan campuran teks;4001 ditolak), bukan UTF16 units atau grapheme clusters; multiline/non-ASCII/invalid UTF8/NUL/oversize dan kind lain tervalidasi; author/org/task/id/time tidak dapat dipalsukan.
- [ ] T02.C2 — No session, expired/revoked session, missing/wrong CSRF, cross-origin serta unsupported methods ditolak sebelum append; error tidak membocorkan SQL/provider/secret.
- [ ] T02.C3 — Current role viewer/analyst/approver read-only; claimed owner lama, membership/brand grant revoked, cross-org dan same-org-other-brand tidak membocorkan catatan atau menulisnya; adapter mengulang query-time authority.
- [ ] T02.C4 — Catatan+audit commit atomik; insert/audit/commit failure tidak menghasilkan partial success. Task status/version/event history tetap sama; catatan task terminal tidak membuka ulang pekerjaan.
- [ ] T02.C5 — UI escaped plain text menahan HTML/script/URL payload; loading/empty/error/refetch/401 dan write-denied states jujur; tidak ada optimistic fixture/auto-retry POST atau assumed agent response.
- [ ] T02.C6 — Ordering timestamp+id konsisten dan batas500 pertama ditampilkan jujur; task dengan >500 catatan belum diterima sebagai complete history sampai pagination benar-benar diimplementasikan.
- [ ] T02.C7 — Clarification hanya human note, bukan agent waiting-input/reply/resume, content approval atau task completion. Perilaku duplicate POST (belum idempotent) serta archive-policy task existing ditinjau sebelum acceptance; jangan menandainya selesai dari unit tests.

### Agent discovery dan route ownership — T03

- [ ] T03.1 — Pin mekanisme/version invocation nyata beserta submit/status/result/error/cancel/retry semantics; endpoint tidak ditebak dari Telegram atau published docs saja.
- [ ] T03.2 — No/wrong credential, wrong peer, root/other route dan cross-tenant access ditolak; correct principal sandbox hanya ke destination yang diizinkan.
- [ ] T03.3 — Get/list/history/cancel/subscribe atau callback features yang didukung menegakkan task ownership, termasuk forged context selector serta pagination/count.
- [ ] T03.4 — Correlation/context collision tidak mencampur org/brand/caller; output tersimpan hanya di destination sandbox yang benar, tidak masuk default/profile lama.
- [ ] T03.5 — Bounded request/result/clarification per identitas sandbox terbukti; kemampuan tidak tersedia tetap limitation/blocked.
- [ ] T03.6 — Timeout/malformed output/connection error terlihat dan tidak memalsukan reachability/completion; cancellation/retry hanya subset yang terverifikasi.

### Outbox dan duplicate/reorder — T04

- [ ] T04.1 — Transaction business state+outbox atomik; rollback tidak mendispatch dan dispatch tidak mendahului persistence.
- [ ] T04.2 — Crash sebelum send dan sesudah accepted sebelum ack dapat dipulihkan tanpa unintended duplicate work; attempt/key/payload hash konsisten.
- [ ] T04.3 — Duplicate consumer admission/callback/event idempotent; out-of-order/delayed event tidak menurunkan atau merusak terminal/version state.
- [ ] T04.4 — Stable retry key dipakai untuk attempt sama, revisi baru punya identity baru; unknown outcome masuk reconciliation sebelum resubmit.
- [ ] T04.5 — Attempt/outcome/dead-letter/operator repair dapat diaudit; klaim dibatasi at-least-once, tidak mengklaim exactly-once end-to-end.

### Temporal — T05

- [ ] T05.1 — Workflow start/activity/retry/timeout/heartbeat diuji pada namespace sandbox dengan policy bounded; permanent/transient failure dibedakan.
- [ ] T05.2 — Human waiting-input/approval pause dan signal resume/cancel tidak melewati authorization atau kehilangan state.
- [ ] T05.3 — Worker kill/restart saat activity berjalan memulihkan workflow dan tidak mengulang side effect yang sudah diterima.
- [ ] T05.4 — API restart, delayed signal/callback dan recovery tetap cocok dengan authoritative PostgreSQL state/history.
- [ ] T05.5 — Replay history/upgrade compatibility diuji untuk versi workflow yang dirilis; nondeterminism/failure terlihat dan ada runbook recovery.
- [ ] T05.6 — Workflow failure/retry/cancel/progress UI jujur dan auditable; local task tests tidak dipakai menutup gate restart Temporal.

## 4. Knowledge, uploads, context, campaign dan version

**Owner usulan:** backend knowledge/storage + frontend + security; DevOps/admin MinIO untuk bucket/credential sandbox. **Inventaris awal:** belum ditemukan suite khusus modul ini; test baru tetap harus dibedakan dari implementasi integrasi nyata. Sumber: Phase 4 §2–9.

### Knowledge/context — K01

- [ ] K01.1 — Authorized CRUD/upload/search/filter/archive knowledge; kategori/owner/source/version/approval/access/timestamps terpelihara.
- [ ] K01.2 — Task-specific retrieval menyelesaikan brand dari server, hanya konteks relevan dan authorized; forged scope/cross-brand context ditolak.
- [ ] K01.3 — Context menyertakan source reference+exact document version dan membedakan approved vs draft; agent tidak diberi confidential unrelated data.
- [ ] K01.4 — Search/cache/index/export setelah grant revoked atau dokumen diarsip tidak membocorkan versi/data yang tidak berhak.
- [ ] K01.5 — Indexed search dievaluasi lebih dahulu; bila vector retrieval dipilih, ada keputusan/model/lifecycle/evaluation dan ACL filtering yang benar, bukan RAG default tanpa bukti.

### Upload/MinIO/serving — K02

- [ ] K02.1 — Image/video/document fixture aman tersimpan di MinIO sandbox dengan metadata/object key PostgreSQL yang sesuai; tidak menyimpan permanent presigned URL.
- [ ] K02.2 — Size/extension/declared MIME/actual signature mismatch, empty/oversize/polyglot/traversal dan filename berbahaya ditolak sesuai policy.
- [ ] K02.3 — Quarantine/malware scan policy disepakati dan diuji; scan unavailable/failed tidak membuat file otomatis tersedia bagi user/publishing.
- [ ] K02.4 — Unsafe preview/XSS/download content disposition aman; unauthorized/cross-org/cross-brand object access ditolak.
- [ ] K02.5 — Presigned URL short-lived, scope/object/version tepat dan expiry ditegakkan; URL lengkap/secret tidak masuk log.
- [ ] K02.6 — Storage failure/interrupted upload tidak meninggalkan metadata published atau object orphan yang tak tertangani; cleanup hanya sandbox.

### Version, campaign dan calendar — K03/K04

- [ ] K03.1 — Material edit content menciptakan versi immutable baru dengan body/metadata/author/change summary/time; approved version tidak di-overwrite.
- [ ] K03.2 — Original dan derived asset versions/hash/reference konsisten; task/deliverable/knowledge/content history dapat dibaca authorized.
- [ ] K03.3 — Exact asset set/version yang akan dipakai downstream berbeda dari alias “latest”; concurrent edit tidak mengganti approved snapshot.
- [ ] K04.1 — Campaign dapat dibuat/dikelola dan dihubungkan ke task/content hanya dalam brand yang diizinkan; cross-brand link ditolak.
- [ ] K04.2 — Objective/audience/channel/date/KPI/owner serta budget/currency tervalidasi; planned budget bukan actual spend.
- [ ] K04.3 — Calendar/list date range/timezone/brand/campaign/channel/status/type filter benar dan tidak leak lintas brand.
- [ ] K04.4 — Planned drafts jelas berbeda dari externally scheduled/published; tanpa Postiz verified tidak ada klaim sudah terschedule eksternal.
- [ ] K04.5 — Dokumentasi object storage, archive/retention dan knowledge/version lifecycle cocok implementasi dan known limitations terlihat.

## 5. Approval exact version dan Postiz uncertainty

**Owner usulan:** backend approval/publishing + frontend + security; admin Postiz dan owner channel untuk sandbox/credential/izin write. **Inventaris awal:** belum ditemukan suite khusus approval/Postiz; [discovery Postiz](../10-postiz-integration.md) bukan live proof. Sumber: Phase 5 §2–8 dan §10.

**Batas keras:** tidak ada publication, scheduling yang dapat berujung publish, account change/spend/destructive external test dalam mandat ini atau otomatis dari full. Gunakan spy/simulator dan sandbox nonpublik yang diverifikasi; bahkan draft/upload eksternal perlu izin tersendiri. Actual network publication acceptance tetap terbuka untuk mandat manusia terpisah, bukan dipalsukan sebagai pass.

### Exact approval dan human gate — P01/P02

- [ ] P01.1 — Approval mengikat exact content version, **seluruh** exact asset versions/derived output, brand, intended platform/channel, intent/schedule, approver dan decision time.
- [ ] P01.2 — Material edit body/metadata/media/destination/schedule menginvalidasi approval atau memerlukan approval baru; tidak transfer otomatis ke versi baru.
- [ ] P01.3 — Race approval/edit/publish diperiksa ulang secara atomik di backend; stale approval/asset alias latest tidak lolos.
- [ ] P01.4 — Submit/reject/revision/comments/comparison/history dan assigned approver/deadline benar serta scoped; keputusan lama tidak ditimpa.
- [ ] P02.1 — Unapproved content tidak dapat melewati application publishing API; uji dengan spy tanpa mengirim ke jaringan social.
- [ ] P02.2 — Unauthorized/cross-brand approve/publish dan forged channel/asset version ditolak backend; UI visibility bukan kontrol utama.
- [ ] P02.3 — Agent task completion tidak memicu publish/approval; explicit authorized human confirmation wajib untuk publication intent.
- [ ] P02.4 — Account changes, advertising spend, brand identity dan destructive operations tetap memerlukan izin manusia serta step-up bila policy menetapkan; tes deny tidak menjalankan operasi itu.

### Contract aktual dan uncertainty — P03/P04/P05

- [ ] P03.1 — Actual deployed Postiz version/image/API/auth/ownership didokumentasi dan diverifikasi di sandbox; jangan menebak endpoint/guarantee dari contoh umum.
- [ ] P03.2 — Workspace/channel brand mapping, disabled/expired channel dan per-network supported format/settings scope tepat; satu key multi-brand bukan isolasi otomatis.
- [ ] P03.3 — Media readiness/processing/transform/version diuji dengan fixture nonpublik berizin; material transform tidak mewarisi approval original diam-diam.
- [ ] P03.4 — UTC/timezone/schedule/partial-channel/error semantics diuji tanpa menjadwalkan publikasi nyata; create 200/QUEUE bukan PUBLISHED.
- [ ] P03.5 — External ID/status/error/sync timestamps disimpan dan ditampilkan sesuai bukti; unknown state/failure actionable dan audit lengkap.
- [ ] P03.6 — Actual network-publication/supported-format acceptance **tetap deferred di luar mandat** sampai ada izin terpisah dan sandbox akun yang aman; tidak ditutup oleh fixture/draft response.
- [ ] P04.1 — Accepted-then-timeout/disconnect/5xx disimulasikan sebagai unknown outcome; tidak blind retry create/publish.
- [ ] P04.2 — Authenticated reconciliation dengan IDs/correlation yang sah mendeteksi external state; ambiguous match diparkir untuk operator, bukan resubmit otomatis.
- [ ] P04.3 — Internal idempotency/outbox duplicate admission mencegah duplicate job; guarantee dedupe eksternal harus dibuktikan, tidak diasumsikan dari internal key.
- [ ] P04.4 — 429/quota/transient/permanent auth/schema/media errors punya bounded retry/backoff yang benar dan failed-operation visibility.
- [ ] P04.5 — Multi-channel partial result dan duplicate/delayed/out-of-order status tidak memalsukan keseluruhan sukses atau merusak approved snapshot.
- [ ] P05.1 — Bila webhook dipakai, actual-version signature/canonical payload/replay/event contract tersedia dan negatifnya diuji; jangan mengarang header signature.
- [ ] P05.2 — Unauthenticated/unverified webhook tidak mengubah authoritative state; bounded authenticated polling fallback diuji bila authenticity belum terbukti.
- [ ] P05.3 — Publication intent/approval/job/outbox/attempt/external status audit dapat ditelusuri tanpa token/payload sensitif; failed/unknown queue punya runbook repair aman.

## 6. Analytics real sources dan Superset RLS

**Owner usulan:** data/analytics + backend + BI/security; owner akun sumber/admin Superset menyediakan izin/scope sandbox. **Inventaris awal:** suite collector/live Superset/RLS belum ditemukan. Sumber: Phase 6 §2–10 dan §11. Superset bukan pengumpul data social.

### Source dan ingestion — N01/N02

- [ ] N01.1 — Setiap platform yang diaktifkan memiliki verified authorized API/source, account type, permissions/app review, retention/history/refresh/rate-limit notes; unsupported platform tidak dijanjikan.
- [ ] N01.2 — Metric definition/source/granularity/unit/timezone/attribution window/organic-paid dan post/account/campaign dimensions terdokumentasi.
- [ ] N01.3 — Missing/unavailable/unauthorized metric ditampilkan null/unavailable dengan alasan, tidak diganti nol atau angka buatan; fixture sintetis diberi label.
- [ ] N01.4 — CTR/conversion/revenue/spend/ROI dihitung hanya dari input sumber sah dan definisi yang benar; planned budget bukan spend aktual.
- [ ] N02.1 — Authorized ingestion mencatat raw source reference+metric names/cursor/window/sync timestamp dan menjaga dimensi platform yang penting.
- [ ] N02.2 — Pagination, rerun, overlap/backfill/correction dan duplicate payload idempotent, tidak double count.
- [ ] N02.3 — 429/token expiry/refresh/transient failure/partial source failure dipulihkan dengan bounded retry dan checkpoint aman.
- [ ] N02.4 — Collector restart/out-of-order data dapat dilanjutkan tanpa kehilangan window atau memalsukan freshness; failure terlihat.
- [ ] N02.5 — Dashboard menunjukkan data freshness/incomplete-source warnings dan filter org/brand/campaign/platform/account/content/period benar.

### Superset access dan report — N03/N04

- [ ] N03.1 — Guest token hanya dibuat backend setelah otorisasi user+dashboard+brand; frontend tidak menerima service/admin credential.
- [ ] N03.2 — Forged org/brand/dashboard ID dan unauthorized embedded request ditolak; token resource scope bukan satu-satunya perlindungan.
- [ ] N03.3 — Actual Superset dataset/query RLS diuji positif/negatif antar org dan brand, termasuk query/export/cache/embedded URL yang tersedia; filter UI tidak dianggap RLS.
- [ ] N03.4 — Token expiry/renewal/revoked membership tidak memperpanjang akses yang dicabut; hasil cache tidak menyeberang tenant.
- [ ] N03.5 — Admin credential tidak ada di bundle/network log/browser storage/telemetry; dashboard hanya memakai dataset yang diizinkan.
- [ ] N04.1 — Competitor record memuat URL sumber, observation/source date, researcher, evidence, confidence/limitation dan report version.
- [ ] N04.2 — Social listening hanya source authorized; observed facts dibedakan dari sentiment/AI interpretation tanpa bypass platform controls.
- [ ] N04.3 — Weekly/monthly report menyatakan period/brand-campaign scope/source references/limitations/missing data; recommendation dibedakan dari measured facts.

## 7. CI/CD, rollback, backup, security, observability dan performance

**Owner usulan:** DevOps/DBA + security + maintainer backend/frontend; owner proyek menerima target/readiness. **Tersedia parsial:** Python repository/release policy/readiness/CLI/VPS boundary fixtures, HTTP/config tests. Workflow yang dibaca bersifat policy/security checks dan release eligibility, **bukan bukti deploy nyata**. Suite rehearsal/restore/failure/load belum ditemukan. Sumber: Phase 7 §2–10 dan §11.

### Security dan CI gates — O01

- [ ] O01.1 — Security review OIDC/MFA/session/CSRF/CORS/cookie/RBAC/API/SQL/XSS/rate-limit/webhook/file upload selesai dan negatifnya diuji, bukan checklist konfigurasi saja.
- [ ] O01.2 — Cross-brand/cross-org privilege escalation kritikal diuji pada API, worker, storage dan analytics; unresolved critical findings memblokir readiness.
- [ ] O01.3 — Runtime secret injection/storage dan telemetry redaction benar; no credential/token/cookie/sensitive payload logging.
- [ ] O01.4 — Container non-root/least privilege, pinned images/resource limits/restart/health/internal networks diverifikasi di staging.
- [ ] O01.5 — DB/cache/MinIO/Temporal/Superset admin tidak terbuka publik; TLS/admin protected path diverifikasi tanpa mengubah jaringan produksi.
- [ ] O01.6 — CI menjalankan formatting/lint/Go unit dan race suite yang sesuai, frontend strict/lint/unit/build, SQL generation drift, OpenAPI dan migration validation nyata.
- [ ] O01.7 — CI integration PostgreSQL disposable dan browser/Auth live staging dibedakan dari unit; skipped mandatory integration bukan hijau.
- [ ] O01.8 — Dependency/container image/security scans dijalankan, findings direview dan exception punya owner/expiry; scanner success bukan bukti tidak ada semua kerentanan.
- [ ] O01.9 — Failing/cancelled/skipped prerequisite, foreign repo/actor, wrong workflow/SHA, direct push dan stale main tidak memperoleh release eligibility.
- [ ] O01.10 — Actual CI/environment/approval enforcement dan exact revision direview terpisah dari synthetic metadata tests; tidak memberi credential produksi ke PR/fork.

### Delivery dan rollback di staging — O02

- [ ] O02.1 — Release artifact/manifest/SHA/digest/provenance diverifikasi pada isolated staging; marker saja tidak memberi readiness.
- [ ] O02.2 — Lock/single promotion/current-main recheck dan untrusted/writable/symlink/stale input denial diuji; policy fixtures bukan bukti hook telah terpasang.
- [ ] O02.3 — Bad candidate atau readiness wrong SHA tidak dipromosikan; HTTP 200 situs lama tidak mengakui aplikasi siap.
- [ ] O02.4 — Kegagalan authenticated/HTTPS smoke memicu rollback staging ke release/config sebelumnya dan health identity diverifikasi kembali.
- [ ] O02.5 — Migration expand/contract compatibility diuji dengan old/new app dan backup checkpoint; irreversible change berhenti/escalate, tidak destructive down otomatis.
- [ ] O02.6 — Runbook deployment/rollback/upgrade menyatakan approval, stop condition, schema limitation dan recovery owner; rehearsal tidak menyentuh situs lama atau Hermes aktif.
- [ ] O02.7 — Tidak ada production deploy/cutover/port/proxy/firewall change dari testing; readiness report dan release approval tetap langkah terpisah milik owner.

### Backup dan restore — O03

- [ ] O03.1 — Target **RPO** PostgreSQL/assets/config disepakati owner+DBA/DevOps; angka belum ada dan tidak boleh diisi perkiraan sebagai target resmi.
- [ ] O03.2 — Target **RTO**, recovery ordering dan definisi awal/akhir pengukuran disepakati owner; angka pending tidak boleh dianggap tercapai.
- [ ] O03.3 — Backup PostgreSQL fixture staging dibuat dengan integrity, encryption/access dan retention yang disetujui; tidak menyalin credential/data produksi ke evidence.
- [ ] O03.4 — Restore PostgreSQL ke database baru terisolasi diverifikasi schema/rows/constraints/org+brand ACL/version consistency, bukan sekadar restore exit 0.
- [ ] O03.5 — Backup dan restore MinIO fixture diverifikasi object count/hash/exact versions/metadata references dan akses bucket, bukan hanya file backup ada.
- [ ] O03.6 — Configuration dan encryption-key recovery procedure diuji melalui secret storage yang disetujui; results hanya reference aman, tanpa key/plaintext backup.
- [ ] O03.7 — Recovery ordering lintas DB/storage/workflow/config diuji; measured recovery/loss dibandingkan target yang telah disetujui.
- [ ] O03.8 — Restore demonstration dan runbook ditinjau DBA/owner; **backup tanpa restore teruji tetap belum valid**.

### Observability dan gangguan — O04/O05

- [ ] O04.1 — Logs/metrics/traces mengikat request→task/workflow/job/attempt dengan correlation aman; sensitive fields disanitasi sebelum export.
- [ ] O04.2 — Request/error/latency, DB/worker/Temporal/Hermes/Postiz/MinIO/ingestion health dan AI usage/cost yang benar-benar tersedia terlihat tanpa fabricated online status.
- [ ] O04.3 — Critical alerts terkirim ke sink uji dan acknowledgement/escalation diverifikasi; dashboard tampil saja bukan alert pass.
- [ ] O04.4 — Freshness/backlog/error/timeout threshold dan penerima alert disepakati; threshold pending bukan SLA yang sudah tercapai.
- [ ] O05.1 — API restart staging mempertahankan session/business state dan memulihkan readiness; tidak menargetkan listener produksi.
- [ ] O05.2 — Worker/Temporal interruption staging memulihkan pekerjaan tanpa duplicate side effect; cocokkan dengan T05 evidence.
- [ ] O05.3 — DB disconnect staging menghasilkan error/readiness jujur lalu recovery terukur tanpa lost write/unsafe retry.
- [ ] O05.4 — Redis interruption diuji **jika digunakan**; jika tidak digunakan, dokumentasikan applicability yang direview, bukan mengklaim test lulus.
- [ ] O05.5 — MinIO outage/interrupted transfer tidak membuat object rusak terlihat siap; reconnect/reconciliation teruji.
- [ ] O05.6 — Hermes timeout/malformed response staging tidak merusak task/schema atau memicu unknown resubmit tanpa reconciliation.
- [ ] O05.7 — Postiz timeout/uncertain result, duplicate webhook dan delayed/out-of-order event diuji dengan simulator/nonpublik sandbox; tidak menerbitkan post.
- [ ] O05.8 — Social sync failure/token expiry/rate limit pulih dari checkpoint dan freshness tidak berbohong.
- [ ] O05.9 — Runbook incident, credential rotation, OIDC outage, integration/storage/DB/Temporal recovery dan rate-limit handling direhearsal di staging oleh operator yang berwenang.

### Performance, documentation dan penerimaan akhir — O06

- [ ] O06.1 — Owner menyetujui workload/dataset/file-size mix/concurrency/durasi/error budget serta latency/throughput/freshness targets; **semua angka masih pending sampai disepakati**.
- [ ] O06.2 — API latency/query/concurrent-user baseline diukur pada isolated staging dengan percentile/error/resource use, bukan kata “cepat”.
- [ ] O06.3 — Upload performance, workflow throughput dan dispatch latency diukur memakai data/sandbox aman tanpa load ke agen/account produksi.
- [ ] O06.4 — Analytics query/ingestion performance diukur per dataset/brand dengan RLS tetap aktif; tidak menonaktifkan security untuk benchmark.
- [ ] O06.5 — Hasil diulang dan dibandingkan target resmi; tanpa target hanya baseline, bukan performance pass atau readiness claim.
- [ ] O06.6 — Audit dokumentasi: README/perintah, arsitektur, OpenAPI, schema/env refs, tested runbooks dan known limitations cocok implementasi; perbaikan file di luar dua dokumen ini adalah pekerjaan terpisah.
- [ ] O06.7 — Semua critical security findings ditutup atau release diblokir; setiap deferred/blocked case tetap terlihat dengan owner/next action.
- [ ] O06.8 — Owner meninjau actual evidence, backup restoration, alert delivery, rollback/failure rehearsal, limitations dan laporan readiness; tidak menganggap local startup/fast/DB/full sebagai persetujuan produksi.

## 8. Register blocker dan pembuka akses

Seluruh baris berikut **pending**, bukan bukti layanan sama sekali tidak ada di host lain. Nama penanggung jawab dan tenggat belum ditetapkan. Jangan menyimpan credential/host internal sensitif dalam kolom ini.

| Blocker | Gate terpengaruh | Owner peran usulan | Next action aman |
|---|---|---|---|
| Runner/long suite belum lengkap | RUN01–RUN11; seluruh full | Maintainer runner + pemilik modul | Implementasi+review command/gate dan redaction, baru pilih eksekusi; jangan ubah status menjadi implemented tanpa suite |
| Docker disposable/tooling belum diverifikasi | A05, PG auth/brand, RUN04–RUN07 | DevOps + maintainer runner | Sediakan Docker lingkungan uji/image pin/capacity; **runner menghasilkan credential**, bukan meminta DSN eksternal |
| Authentik live/client/test users/MFA belum dibuktikan | A01–A03 | Admin Authentik + security | Provision/konfirmasi staging issuer/client/callback/users dan secret ref private, tanpa password/token di chat/results |
| Matriks role/grant/reauth belum lengkap | A02/A04/B01/P02 | Backend + security + owner proyek | Tetapkan policy eksplisit dan fixture allow/deny; unknown policy tetap blocker |
| Hermes route/security/sandbox belum disetujui | T03–T05 | Maintainer integrasi + owner agen | Review discovery, siapkan identitas/storage sandbox terpisah; tidak memakai profile lama atau menebak endpoint |
| Temporal namespace/worker/replay suite belum siap | T05/O05 | Workflow maintainer + DevOps | Sediakan versi+namespace/task queue staging dan restart/fault permit |
| MinIO/scan/knowledge/version lifecycle belum siap | K01–K03/O03 | Backend/storage + admin MinIO + security | Sediakan bucket/policy sandbox, scan policy dan fixture version; credential lewat secret ref |
| Approval/Postiz capability/auth/webhook contract belum terverifikasi | P01–P05 | Backend publishing + admin Postiz + owner channel | Pin versi/API, sandbox nonpublik dan permission; simulator uncertainty dahulu; no publish |
| Actual network publishing dilarang dalam mandat ini | P03.6/acceptance live Phase 5 | Owner proyek + owner channel | Biarkan deferred; memerlukan mandat eksplisit terpisah, bukan bypass full atau post percobaan otomatis |
| Authorized social source/scopes/metric definitions belum lengkap | N01/N02/N04 | Owner data/akun + maintainer analytics | Sediakan akses read-only sandbox/authorized source, scope/review/definition; tidak isi metrik buatan |
| Superset dashboard/dataset/RLS/token policy belum dibuktikan | N03 | Admin BI + backend/security | Siapkan staging dataset lintas brand dan service credential server-side; buktikan query RLS nyata |
| Delivery/recovery/telemetry infra belum direhearsal | O01–O05 | DevOps/DBA/security | Siapkan isolated staging + private evidence/alert sink dan per-fault approval; tanpa perubahan produksi |
| RPO/RTO/performance/freshness/retention target belum disetujui | O03/O04/O06/EV05 | Owner proyek + DBA/DevOps + data/product | Setujui workload/target/retention dan reviewer, lalu bandingkan hasil ukur; jangan mengarang angka |

## 9. Catatan penutupan kasus — diisi sesudah eksekusi

Untuk setiap kasus yang dieksekusi, catat ringkasan aman berikut (tanpa credential atau raw response):

```text
Kasus / suite:
Revisi / build identity:
Mode / label staging sandbox:
Run ID / waktu:
Hasil nyata: pass | fail | blocked | not_run | skipped
Exit code / assertion / durasi:
Evidence teredaksi: artifacts/testing/<run-id>/...
Blocker / owner / next action:
Cleanup resource run:
Reviewer / keputusan / tanggal:
```

Tidak ada baris hasil yang diisi pada tugas dokumentasi ini. Checklist tetap terbuka sampai bukti aktual tersedia; pengembangan dapat lanjut tanpa menyembunyikan utang tes panjang.
