# 13 — Strategi pengujian bertahap

## 1. Keputusan dan batas bukti

**Prioritas saat ini: pengembangan. Pengujian panjang ditunda, bukan dianggap lulus.** Dokumen ini menyiapkan urutan, data, penanggung jawab, dan kriteria bukti agar utang pengujian tidak hilang. Pekerjaan dokumentasi ini hanya membaca spesifikasi/inventaris; **tidak menjalankan tes, build, Docker, jaringan, login, deployment, atau perubahan produksi**.

Sumber persyaratan lokal:

| Sumber | Bagian yang menjadi acuan |
|---|---|
| [Phase 1](../PHASE-01-APPLICATION-FOUNDATION.md) | §2.2–2.8 dan §5: bootstrap, OIDC, session, CSRF, RBAC, audit, migrasi |
| [Phase 2](../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md) | §2–8 dan §9: brand isolation, registry enam agen, status berbasis bukti |
| [Phase 3](../PHASE-03-HERMES-INTEGRATION-AND-WORKFLOWS.md) | §2–7 dan §10: integration discovery, task, outbox, Temporal, recovery |
| [Phase 4](../PHASE-04-KNOWLEDGE-ASSETS-AND-CONTENT.md) | §2–9: knowledge/context, MinIO, campaign, version, calendar |
| [Phase 5](../PHASE-05-APPROVAL-AND-POSTIZ-PUBLISHING.md) | §2–8 dan §10: exact approval snapshot, authorization, Postiz uncertainty |
| [Phase 6](../PHASE-06-ANALYTICS-AND-MARKETING-INTELLIGENCE.md) | §2–10 dan §11: sumber metrik, ingestion, Superset, batas interpretasi |
| [Phase 7](../PHASE-07-PRODUCTION-HARDENING-AND-OPERATIONS.md) | §2–10 dan §11: security, failure, backup/restore, performance, CI/CD |

[Checklist tertunda](testing/DEFERRED-CHECKLIST.md) adalah register acceptance per kasus. Penanggung jawab di sana merupakan **usulan peran**, bukan klaim seseorang sudah menerima tugas. Owner proyek perlu menetapkan nama dan jadwal; tenggat belum disepakati.

### Makna status

- **Berkas tersedia:** isi tes dibaca secara statis; bukan bukti tes dapat dikompilasi atau sudah lulus.
- **Parsial/fixture:** hanya boundary yang disebut yang dicakup. Mock, signed local OIDC, dan synthetic API response tidak membuktikan layanan nyata.
- **Rencana/belum ditemukan:** belum ditemukan suite khusus dalam inventaris yang dibaca; implementasi paralel dapat menambah berkas setelah inspeksi. Cocokkan ulang sebelum eksekusi.
- **Deferred / not_run:** sengaja belum dijalankan. **Blocked:** prasyarat, implementasi, credential, atau izin belum ada. Keduanya bukan pass; `skip` juga bukan pass.
- **Pass:** hanya untuk kasus yang benar-benar dieksekusi pada revisi/lingkungan tercatat, dengan assertion, exit code dan bukti teredaksi yang direview. Pass suatu unit test bukan persetujuan acceptance satu fase.

Inventaris ini snapshot pembacaan 6 November 2026 pada workspace yang sedang dikembangkan. Tidak membekukan jumlah tes atau hasil run terdahulu. Runtime wiring, frontend/task backend, dan runner dapat berubah bersamaan. Authentik, Temporal, MinIO, Postiz, dan Superset live **belum dibuktikan oleh dokumentasi ini**; jangan menyatakan integrasi provisioned/operasional berdasarkan keberadaan adapter atau dokumentasi discovery.

## 2. Inventaris tes yang benar-benar dibaca

| Berkas yang tersedia saat inspeksi | Cakupan assertion yang terlihat | Batas bukti |
|---|---|---|
| `internal/application/auth_test.go` | Opaque token berbeda, hash, pengikatan CSRF | Unit helper; bukan lifecycle session lengkap |
| `internal/adapters/oidc/provider_test.go` | Signed ID token RS256 dari `httptest`, discovery/JWKS lokal, PKCE S256, issuer/audience/nonce/signature/expiry negatif | **OIDC fixture bertanda tangan, bukan Authentik live**, bukan browser login end-to-end |
| `internal/adapters/httpapi/server_test.go` | Session 401/200, logout dengan CSRF/Origin negatif, revocation pada memory store; login state/browser cookie dan bad callback; readiness/request ID; API JSON bukan HTML fallback; static serving | AuthStore/provider di-mock; belum matriks semua mutasi, kegagalan dependency, atau runtime nyata |
| `internal/platform/config_test.go` | Secret file private, raw secret environment ditolak, origin tidak aman ditolak | Unit konfigurasi, bukan inspeksi permission host/deployment |
| `internal/domain/brand_test.go` | Validasi brand/URL/status; registry enam ID stabil, disabled/unknown | Tidak menghubungi enam agen; hanya sebagian identitas diassert eksplisit |
| `internal/application/brands_test.go` | Membership default-deny dan role create default-deny | Bukan matriks seluruh role × operasi × organisasi/brand |
| `internal/adapters/postgres/store_test.go` | Migrasi melalui `testStore`; brand create/list/archive; cross-org dan same-org tanpa explicit brand grant ditolak | Integrasi PostgreSQL **opt-in**, belum bukti semua endpoint/role/SQL isolation |
| `internal/adapters/postgres/auth_test.go` | Session revoke/idle/absolute expiry, identity belum diprovision ditolak, reuse hash ditolak; login request single-use/browser binding | Menggunakan PostgreSQL opt-in; bukan Authentik live atau uji browser |
| `internal/domain/tasks_test.go` | Kontrak draft → ready → assignment langsung ke spesialis → cancel; status eksekusi ditolak saat disabled; terminal tidak dibuka ulang | Tes pengembangan yang baru ditambahkan; pasangan source/kompilasi belum diverifikasi; **bukan dispatch, Temporal, atau full lifecycle aktif** |
| `web/src/lib/api.test.ts`, `web/src/app/App.test.tsx` | API same-origin/session validation/CSRF/error; login link dan shell | `fetch` fixture; tidak melakukan autentikasi backend |
| `web/src/features/Brands.test.tsx`, `web/src/features/Agents.test.tsx` | Form/CRUD/konfirmasi arsip, directory enam agen dan status jujur | Transport frontend fixture dari `web/src/test/fixtures.ts`; bukan ACL backend atau eksekusi Hermes |
| `tests/test_release_policy.py`, `tests/test_release_cli.py`, `tests/test_vps_delivery.py` | Synthetic GitHub metadata, wrong actor/SHA/repo/CI dan stale main; kontrak trusted JSON/filesystem | GitHub boundary mocked; bukan release/deploy nyata. Ekspektasi tes tidak membuktikan helper yang sedang dikerjakan telah lengkap |
| `tests/test_release_readiness.py`, `tests/test_deploy_boundary.py` | Readiness marker dan batas workflow tanpa transport credential/eksekusi produksi | Local policy/static workflow assertion; bukan rollback rehearsal |
| `tests/test_repository_checks.py`, `tests/test_phase_deliverables.py` | Link/deliverable/scanner fixtures, diagnostik tidak membocorkan nilai | Bukan audit keamanan aplikasi menyeluruh |
| `tests/test_test_setup.py` | Kontrak redaction, private report, fast terpisah, full fail-closed | Runner sedang dikerjakan; keberadaan tes runner bukan hasil verifikasi runner |

`testStore` membaca `MB_TEST_DATABASE_URL`: tanpa DSN ia dapat `Skip`; dengan `MB_INTEGRATION=1` tetapi DSN kosong ia gagal. DSN yang ada dapat mengaktifkan integrasi meskipun caller tidak bermaksud demikian. Karena itu **jangan menjalankan Go tests dengan environment warisan sembarang**. Hanya runner mode `database` yang boleh menyuntikkan DSN disposable; skip default tidak boleh dilaporkan sebagai PostgreSQL pass.

`web/package.json` menyediakan unit/lint/typecheck/build dan script Playwright. Pada inspeksi awal belum ditemukan berkas browser `.spec.*`; script/dependency Playwright saja bukan suite E2E terimplementasi. `.github/workflows/ci.yml` yang dibaca masih berfokus pada Python policy/repository/security lint; `deploy.yml` menyebut **Release eligibility (no deployment)**. Bedakan isi workflow aktual dari rancangan fase berikutnya; dokumen lama dapat mendahului perubahan source baru.

### Pembaruan development setelah snapshot inventaris

Runner kini tersedia dengan profil/perintah detail di [setup](../testing/README.md). Enam unit test runner lulus, sedangkan full ditolak karena seluruh long suite masih planned. Ini tidak menutup acceptance live. Task domain/application/PostgreSQL adapter/API kini tersambung dalam source; detail dan bukti short checks di [task development](12-task-development.md). Entry point runtime/admin ada di [local development](14-local-development.md). Frontend259 unit/transport-fixture tests, typecheck, lint dan production build lulus setelah UI revision-request terintegrasi; dua warning anotasi Zod/Rollup nonfatal. Ini bukan browser/backend/agent E2E. Dirty brief/raw draft coordination fix terintegrasi dengan8 regresi baru; review ulang terbatas bersih (passed=true). Decoder invalid UTF8 sudah diperbaiki RED→GREEN dan parent short Go checks/build/vet lulus; re-review decoder terbatas bersih (passed=true). Dua mismatch source empty-query/Unicode whitespace Go-vs-JS sudah diperbaiki dan review ulang terbatas bersih. Fixtures pass tidak menutup blocker lain. Backend dependency graph dan human-imported immutable deliverable revisions sudah ditambahkan dengan short checks Go; SQL/migrations000004/000005 dan database acceptance belum dijalankan. Frontend deliverable import/viewer terintegrasi; empat blocker review telah diperbaiki dan bounded independent rereview bersih. Checklist aktual dihitung ulang dari seluruh checkbox:238 unchecked/0 checked, termasuk delapan tambahan revision-request cases T02.R1–R8 setelah230 kasus sebelumnya; belum ada acceptance dicentang, termasuk7 graph cases dan8 manual deliverable cases. Domain/application/HTTP revision-request short tests+build+vet lulus, tetapi bounded review menemukan collision suffix route deliverable detail vs revision-requests; narrow fix+regression terintegrasi dan parent HTTP tests/build/diffcheck lulus; bounded rereview selesai passed=true tanpa security/logic findings; collision tertutup pada source+HTTP fixtures. Persistence/UI revision-request terintegrasi, parent short Go checks/build/vet dan Postgres compile-only lulus; independent persistence/UI review berjalan. Migration000006/realSQL/browser belum dijalankan. Cocokkan kembali inventaris/revisi sebelum full testing.

## 3. Jalur eksekusi yang disiapkan

Kontrak runner yang tersedia di `scripts/test_setup.py` (full tetap fail-closed):

```text
python3 scripts/test_setup.py plan
python3 scripts/test_setup.py fast
python3 scripts/test_setup.py database
python3 scripts/test_setup.py full
```

Ini petunjuk penggunaan **setelah implementasi ditinjau**, bukan perintah yang telah dijalankan atau janji bahwa stub sudah berfungsi. Jangan mengarang flag tambahan. Cocokkan bantuan/plan aktual, daftar long gate, dan checklist sebelum menjalankan mode berikutnya.

| Mode | Tujuan dan batas wajib | Arti hasil |
|---|---|---|
| `plan` | Menampilkan suite pendek/panjang, prasyarat dan blocker. Tidak mulai service, container, browser login, install dependency, atau test | Rencana saja, bukan hasil tes |
| `fast` | Unit Python/Go/frontend, lint/typecheck dan build pendek sesuai implementasi runner. Tidak Docker/DB/live OIDC/browser E2E/failure injection/load test | Feedback pengembangan untuk suite yang benar-benar dijalankan, bukan readiness |
| `database` | **Eksklusif PostgreSQL Docker disposable** dengan credential acak buatan runner, database/volume/network khusus run, bind loopback, timeout dan cleanup terarah | Bukti integrasi DB saja; bukan Authentik atau seluruh acceptance |
| `full` | Permintaan eksplisit setelah preflight; hanya isolated staging/sandbox. Tolak jika inventaris long suite kosong, suite belum implemented/tanpa command, prasyarat tidak valid, atau target/izin tidak jelas | **Fail-closed hingga seluruh long gate terimplementasi**; tidak fallback ke `fast`, tidak pass karena skip |

Aturan implementasi/operasional runner yang harus direview, bukan klaim sudah terverifikasi:

1. `fast` membersihkan/menolak integration environment warisan, terutama `MB_TEST_DATABASE_URL` dan `MB_INTEGRATION`, serta akses credential/service live. Unit OIDC `httptest` tetap fixture lokal. Tidak memakai fake success ketika dependency/build gagal.
2. `database` menolak DSN eksternal, termasuk DSN staging yang disuplai manual. Credential/DSN ephemeral hanya diberikan ke child test process; jangan dicetak, disimpan dalam results, command line, `.env`, atau Git. Identifikasi container milik run sendiri sebelum migrasi maupun cleanup.
3. Jangan menggunakan host PostgreSQL, database bersama, volume produksi, atau credential runtime aplikasi untuk `database`. Container gagal dibuat berarti blocked/fail, bukan skip. Cleanup hanya resource berlabel/ID run itu; tidak `prune`, tidak menghapus volume lain.
4. `full` tidak diaktifkan oleh environment flag tidak sengaja. Status `planned`, `missing`, `deferred`, `blocked`, nilai tidak dikenal, atau sekadar tulisan `passed` dalam config tidak membuktikan suite implemented. Setelah preflight, semua command wajib benar-benar berjalan dan hasilnya diperiksa.
5. Jangan menandai long suite implemented untuk membuka gate saja. Hubungkan setiap suite dengan ID kasus checklist, command nyata, fixture aman, batas waktu, evidence dan owner. Satu suite yang hilang tetap memblokir klaim full selesai.

## 4. Preflight manual sebelum eksekusi di kemudian hari

### Lingkungan dan alat

- Tetapkan revisi/build identity, scope kasus, operator, reviewer, jadwal, izin failure injection, dan cara penghentian. Jika workspace masih berubah, jangan menganggap output berlaku untuk source sesudahnya.
- Go mengikuti `go.mod` yang dibaca (`1.27.3`); path lokal yang diberikan konteks pengembangan `/home/ubuntu/.local/lib/go/bin/go` bukan jaminan tersedia di mesin lain. Node mengikuti `web/package.json` (`>=22.12.0`), npm/lockfile proyek, Python kompatibel runner; CI yang dibaca memakai Python 3.12/3.14. Cek versi nyata nanti; jangan menulis hasil versi tanpa pemeriksaan.
- Dependency Go/npm/Python harus sudah tersedia dari sumber yang disetujui. Instalasi/download bukan bagian tugas dokumentasi atau otomatis dianggap aman. Build dan unit frontend tidak membutuhkan credential OIDC.
- Untuk `database`: Docker daemon yang diizinkan, kapasitas disk/memori/port, image PostgreSQL yang dipin/disetujui, dan izin membuat resource disposable. Akses Docker setara hak tinggi; jangan memberi akses daemon produksi demi tes.
- Untuk `full`: deployment staging terpisah, browser Playwright dan binary yang cocok, TLS/CA/redirect URI staging, sink notifikasi uji, clock/timezone terkendali, dan dependency sandbox. Tool migrasi/query sesuai repo; tool Temporal/replay, load test, vulnerability/image scan, `actionlint`/Gitleaks sesuai gate yang kelak dibuat. Belum mengklaim semua tool tersebut terpasang.
- Pastikan tidak ada auto-publish, scheduler ke akun nyata, email/Telegram ke pengguna nyata, iklan/spend, atau webhook keluar yang mengarah ke produksi. Hentikan jika allowlist target belum bisa dibuktikan.

### Environment dan akses

| Kebutuhan | Nilai/ref yang disiapkan operator, tanpa menyalin secret | Owner pembuka blocker yang diusulkan |
|---|---|---|
| Runtime foundation staging | `MB_PUBLIC_URL`, `MB_HTTP_ADDR`, `MB_COMMIT_SHA`, `MB_WEB_DIR`; origin HTTPS dan bind loopback sesuai validator aktual | Backend + DevOps |
| Database runtime | `MB_DATABASE_URL_FILE` private; role aplikasi/migrasi least privilege terpisah | DevOps/DBA |
| OIDC live | `MB_OIDC_ISSUER`, `MB_OIDC_CLIENT_ID`, `MB_OIDC_CLIENT_SECRET_FILE`; provider/client/test users/MFA dan callback staging terverifikasi | Admin Authentik + owner keamanan |
| Database test disposable | Runner menghasilkan dan menyuntikkan `MB_TEST_DATABASE_URL` + `MB_INTEGRATION=1` hanya untuk child integrasi | Maintainer runner + DevOps; **bukan meminta DSN user** |
| Hermes | Route/principal/credential referensi sandbox yang terisolasi dan integration discovery disetujui; jangan menebak endpoint | Maintainer integrasi + owner agen |
| Temporal / MinIO | Namespace/task queue serta bucket/key prefix/credential reference khusus staging; versi dan policy diketahui | DevOps + maintainer workflow/storage |
| Postiz / social data | Versi/API/workspace/channel sandbox, scope dan izin operasi yang jelas; collector hanya akun sumber yang diotorisasi | Admin Postiz + owner akun/sumber |
| Superset | Dashboard/dataset staging, service credential server-side, guest-token policy dan RLS | Admin BI + backend/security |
| Backup/observability | Storage restore terisolasi, key recovery ref, sink alert uji, collector/log access private | DevOps/DBA/security |

Nama environment di atas hanya yang sudah terlihat pada foundation/DB tests. Konfigurasi integrasi mendatang harus mengikuti implementasi yang direview; jangan menciptakan nama env/endpoint untuk fitur yang belum tersedia. Credential disediakan lewat secret manager atau file private di luar repository. Jangan meminta password/token dalam chat. Kekurangan credential adalah **blocker milik admin/owner terkait**, bukan alasan menggunakan akun admin produksi atau melemahkan validasi.

### Isolasi wajib

Gunakan fixture dua organisasi dan minimal dua brand dalam organisasi pertama, dengan user berizin/no-grant, seluruh role, session revoked/expired, dan channel/bucket/dashboard terpisah. Data harus sintetis dan ditandai sebagai fixture, tanpa PII atau konten brand nyata. Jangan mengimpor backup produksi untuk kemudahan.

Seluruh full testing hanya di **isolated staging dan sandbox**: database, storage, Temporal namespace, cache jika digunakan, domain, account/channel dan telemetry sink terpisah. Jangan memakai situs lama, port/listener situs aktif, deployment produksi, dashboard Hermes aktif, atau profile/home/session/storage enam Hermes yang sudah ada sebagai fixture. Enam identitas registry dapat divalidasi tanpa mengirim pekerjaan; eksekusi memerlukan sandbox yang owner setujui, bukan mengubah prompt/persona/profile lama. Tidak membuka port publik/tunnel atau mengubah proxy/firewall untuk mempercepat tes.

**Tidak ada publication dalam mandat ini, termasuk otomatis dari `full`; tidak ada destructive test produksi.** Postiz uncertainty dapat diuji dengan simulator fault dan, setelah izin tersendiri, operasi nonpublik sandbox yang diverifikasi aman. Draft/upload eksternal pun merupakan write dan perlu izin. Bukti network publishing nyata tetap deferred/di luar cakupan sampai mandat terpisah diberikan; jangan menyamakan draft/create HTTP 200 dengan publish. Backup/restore, restart, network interruption, rollback dan load hanya menyasar resource staging/disposable yang diizinkan.

## 5. Matriks cakupan per modul

Semua hasil eksekusi matriks berikut saat dokumen dibuat: **not_run/deferred**. Kolom berkas merujuk inventaris §2, bukan pass. `U` = unit/fixture, `D` = PostgreSQL disposable, `S` = staging/sandbox, `M` = review/manual dengan bukti. ID dirinci lagi dalam checklist.

| ID / modul | Kasus positif, negatif, dan invariant | Lapisan | Berkas tersedia / gap yang harus ditutup |
|---|---|---|---|
| A01 auth OIDC | Login permitted user; state/browser binding/replay; PKCE S256; nonce/issuer/audience/signature/expiry/claims salah; issuer+subject identity bukan email | U, D, S | `auth_test.go`, `provider_test.go`, `server_test.go`, PG `auth_test.go` parsial; live Authentik/MFA/browser belum dibuktikan |
| A02 session | Opaque HttpOnly/Secure/SameSite cookie; idle+absolute expiry; revoke/logout; rotasi sesudah login/perubahan privilege; sensitive-action reauth; provider logout limitation | U, D, S | Helper/HTTP/PG tersedia parsial; rotasi privilege/reauth dan provider logout live perlu suite |
| A03 CSRF / API / frontend | Tolak tanpa session; setiap mutasi menolak missing/wrong CSRF, origin asing/kosong; CORS aman; error 401/403/404 konsisten tanpa bocor; tidak ada bearer/localStorage; session-expired UI | U, S | HTTP logout dan frontend fixture ada; perlu seluruh mutasi dan browser cookie/CORS/XSS |
| A04 org+brand / RBAC | Role × read/create/update/archive/member/admin; cross-org, same-org tanpa grant, forged brand ID, list/count/search, worker context; privilege escalation ditolak | U, D, S | Brand use-case/PG isolation parsial; membership/role inheritance/all endpoint matrix belum lengkap |
| A05 migrasi / SQL | Empty DB; apply ulang; checksum drift; transaksi/concurrent migrator; constraints/index; SQL injection; role migrasi terpisah dari role app | D, M | `db/migrate.go` dipanggil `testStore`; suite khusus migrasi/drift/least privilege belum ditemukan |
| A06 runtime / audit | Start/shutdown terdokumentasi, health vs readiness saat dependency gagal, commit identity, OpenAPI sesuai API; audit action/security tanpa secret; static bukan API fallback | U, S, M | HTTP/config tests parsial; startup wiring, failure probes, audit completeness dan contract suite perlu verifikasi |
| B01 brands / registry | CRUD/member authorized; enam nama/role/Telegram/ID tepat dan seed idempotent; profile lama utuh; disabled tidak dispatch; configured ≠ reachable ≠ ready; unknown tidak disulap online | U, D, S, M | Domain/brand/frontend tersedia parsial; membership, status signal, seluruh identitas/seed perlu diperluas |
| T01 task / transitions | Draft/ready/assigned/cancelled subset aman; semua transisi sah/ilegal; optimistic version conflict, terminal, permission, audit; execution-disabled tidak tampak running | U, D, S | `internal/domain/tasks_test.go` kontrak subset; full lifecycle/persistence/API masih pekerjaan pengembangan |
| T02 assignment / deliverable | Direct specialist tanpa Claudia; dependency/cycle, reassignment, comments/revision/cancel; correlation; request/response schema version; malformed/raw output berlabel, version immutable | U, D, S | Subset direct assignment di domain; suite use-case/API/deliverable version belum ditemukan pada inspeksi |
| T03 agent dispatch | Mekanisme terverifikasi; no/wrong credential, wrong peer/route/tenant ditolak; ownership get/list/cancel/history; output hanya ke sandbox destination; capability disabled fail-closed | U, S, M | Registry fixture bukan A2A proof; discovery/security gate dan sandbox per identitas diperlukan |
| T04 outbox / idempotency | Business state+outbox atomik; crash sebelum/sesudah send; duplicate attempts/callbacks dan out-of-order tidak double work; stable key; unknown reconcile; audit outcomes | U, D, S | Suite outbox/dedupe/fault belum ditemukan; tidak ada klaim exactly-once |
| T05 Temporal | Retry/timeouts/heartbeat, waiting input/approval, cancellation; kill/restart worker/API; replay versi workflow, delayed signal; business DB konsisten, side effect tidak diulang | U, S | Live Temporal/restart/replay suite belum ditemukan; helper/domain task bukan substitusi |
| K01 knowledge / context | CRUD/search/filter/archive authorized; task resolves brand; relevant approved vs draft; source+exact document version; revoked grant tidak leak melalui cache/search/agent context | U, D, S | Suite khusus belum ditemukan; retrieval biasa dahulu, vector/RAG hanya bila keputusan/evaluasinya ada |
| K02 upload / serving | Ukuran/extension/MIME/signature cocok; empty/oversize/polyglot/traversal dan unsafe preview ditolak; quarantine/malware policy; presigned TTL, unauthorized/cross-brand object ditolak | U, S | MinIO/upload/scan suite belum ditemukan; scanner policy dan sandbox storage blocker |
| K03 asset / content version | Original/derived asset references immutable, PostgreSQL metadata sesuai object hash; edit membuat versi baru; approved tidak overwrite; task/deliverable/content history terpelihara | U, D, S | Suite version storage belum ditemukan; pilih exact asset version saat approval |
| K04 campaign / calendar | Campaign↔task↔content same brand; date/timezone/filter, budget/currency planned bukan actual spend; planned draft bukan externally scheduled | U, D, S | Suite campaign/calendar belum ditemukan; UI fixture tidak membuktikan scheduling |
| P01 exact approval | Content version + seluruh asset versions + brand/channel/platform + intent/schedule + approver/time bound; material edit/in-flight race invalidates; stale approval tidak dipakai | U, D, S | Suite approval backend/concurrency belum ditemukan |
| P02 human gate | Unauthorized/unapproved/cross-brand publish ditolak; selesai task tidak publish; explicit human permission/confirmation untuk publish/account/spend/identity/destructive action | U, D, S, M | Tes deny dapat memakai publishing spy tanpa side effect; live publish tetap dilarang |
| P03 Postiz contract | Pin actual version/API; workspace/channel ownership, media readiness/transform exact version, UTC schedule semantics, external IDs/status/error mapping, per-channel partial result | U, S, M | [Discovery Postiz](10-postiz-integration.md) bukan adapter live; endpoint/capability/webhook/idempotency guarantee tidak diasumsikan |
| P04 Postiz uncertain outcome | Accepted-then-timeout/disconnect/5xx → unknown → authenticated reconciliation, ambiguous parkir operator; no blind retry; 429 bounded backoff; duplicate/delayed event tidak merusak state | U, D, S | Simulator/fault+reconcile suite belum ditemukan; internal key bukan jaminan dedupe eksternal |
| P05 webhook / polling | Version-specific signature/replay/body limit; untrusted event tidak mengubah authoritative state; jika authenticity belum terbukti, authenticated bounded polling; status/audit/failed queue terlihat | U, S | Kontrak webhook aktual belum dibuktikan; jangan mengarang nama header signature |
| N01 sumber metrik | Verified authorized source per platform; scope/review/account requirements; definition/unit/granularity/timezone/attribution/organic-paid; null/unavailable bukan angka buatan; CTR/ROI hanya jika input sah | U, S, M | Suite collector nyata belum ditemukan; fixture sintetis harus berlabel, Superset bukan collector |
| N02 ingestion | Cursor/windows/pagination, dedupe/backfill/correction, 429/expiry/refresh, partial failure/restart recovery; preserve source/raw ref; freshness dan missing-source warning | U, D, S | Suite ingestion/idempotency/recovery belum ditemukan |
| N03 Superset / RLS | Guest token server-side; forged dashboard/brand/org ditolak; tenant RLS tetap berlaku pada query/export/cache/embedded URL; expiry/renewal/revoked membership; admin secret tidak ke browser | U, S, M | Superset live dan RLS suite belum ditemukan; token claims saja tidak membuktikan datasource isolation |
| N04 reports / research | Competitor source URL/tanggal observation/evidence/confidence/report version; social listening authorized; AI interpretation/recommendation berbeda dari fakta; report period/scope/source limitations | U, S, M | Suite source lineage/report belum ditemukan; tidak menjanjikan cakupan platform tanpa akses |
| O01 security / CI | Formatting/lint/unit/build, sqlc/OpenAPI/migration checks, dependency/container scan; CORS/XSS/rate-limit/SQL/webhook/privilege exposure; failing/cancelled/skipped gate tidak release | U, D, S, M | Python policy/workflow tests tersedia; Go/frontend/DB/image/security gate tambahan perlu wiring dan bukti aktual |
| O02 delivery / rollback | Exact current-main SHA/owner/CI/digest; staging lock/stale input denied; bad candidate/wrong SHA/HTTPS smoke rollback ke release sebelumnya; schema expand/contract, no destructive down migration | U, S, M | Policy fixtures tersedia; staged deployment+rollback rehearsal belum dibuktikan; attestation bukan CD |
| O03 backup / restore | PostgreSQL+MinIO+metadata/config/key recovery; restore baru terisolasi, integrity/hash/row+version+ACL check, recovery ordering, retention dan akses backup | S, M | Suite restore belum ditemukan; backup tanpa demonstrated restore bukan valid |
| O04 observability | Redacted logs/metrics/traces, request/task/job correlation; DB/worker/Temporal/Hermes/Postiz/MinIO/ingestion/cost health; critical alert masuk sink uji, bukan sekadar dashboard ada | U, S, M | Request ID assertion tersedia; telemetry coverage/alert delivery belum dibuktikan |
| O05 reliability | API/worker restart, DB disconnect, Redis bila dipakai, MinIO/Temporal outage, Hermes timeout/malformed, Postiz unknown, duplicate/delayed/out-of-order callback, sync failure → recovery/reconciliation | S | Suite gangguan lintas sistem belum ditemukan; izin per fault diperlukan |
| O06 performance | Workload realistis: user/concurrency, latency API/query/upload, workflow throughput, dispatch, analytics; ukur baseline dan percentile/error/resource use | S, M | Suite/target belum disepakati; tidak load test production atau upstream nyata |

## 6. Urutan development-first dan gate penerimaan

1. Sekarang: tambah fitur beserta unit/contract test pendek, jaga boundary disabled/missing integration jujur; perbarui inventaris/checklist tanpa menandai deferred sebagai pass.
2. Setelah runner siap dan operator memilihnya: `plan`, lalu `fast` untuk feedback pendek. Kegagalan nyata tetap dilaporkan; tidak mengganti output dengan angka perkiraan atau hasil run lama.
3. Saat persistence siap: `database` dengan Docker disposable; laporkan explicit executed vs skipped. Jangan menyimpulkan isolasi seluruh modul dari brand tests saja.
4. Setelah seluruh long suite implemented dan preflight disetujui: full staging dengan urutan auth/isolation → task/reliability → storage/version → approval/contract → analytics/RLS → ops/restore/performance. Integration discovery yang gagal memblokir langkah dependen; tidak mengaktifkan fitur berdasarkan mock.
5. Reviewer memeriksa bukti per acceptance, blocker, known limitations dan runbook. Owner menyetujui readiness tersendiri. `fast`, `database`, `full`, marker release, startup lokal, atau halaman lama HTTP 200 **tidak otomatis memberi izin produksi**.

### Target yang harus disepakati, belum ada angka

| Target pending | Keputusan yang dibutuhkan | Owner usulan |
|---|---|---|
| RPO | Batas kehilangan data per PostgreSQL/asset/config, frekuensi backup dan retention | Owner proyek + DBA/DevOps |
| RTO | Recovery ordering, titik mulai/akhir pengukuran, waktu pulih yang disepakati | Owner proyek + DevOps |
| Performance | Workload, dataset/file-size mix, concurrency, durasi, p95/p99 API/query, error budget, throughput/dispatch/ingestion/upload budgets | Product/owner + backend/DevOps |
| Freshness/alert | Sync freshness, timeout/backlog thresholds, penerima dan escalation path | Owner data + DevOps |

Catat hasil ukur aktual dahulu; tanpa target yang disetujui jangan menyebut “memenuhi RPO/RTO”, “cepat”, “scalable”, atau performance pass. Jangan mengarang angka target agar tabel lengkap.

## 7. Bukti, keamanan hasil, dan cara menutup checklist

Semua bukti lokal di **`artifacts/testing/<run-id>/`**. `.gitignore` yang dibaca sudah mengabaikan `artifacts/`; jangan menambahkan hasil ke Git atau memindahkannya ke docs agar ikut commit. `artifacts/testing` dan setiap direktori run harus private **0700**, file **0600**, dibuat dengan umask 077 dan diperiksa ownership/symlink; jangan melonggarkan permission di luar folder yang dimiliki run.

Isi laporan aman: ID run, waktu mulai/selesai, revisi/build identity, versi alat/image, label target nonsecret, scope/mode, ID acceptance/suite, status `pass/fail/blocked/not_run/skipped`, exit code, durasi, ringkasan assertion, pointer bukti teredaksi, blocker+owner, resource cleanup dan keputusan reviewer. Simpan hanya command yang sudah disanitasi. Tidak ada hasil eksekusi yang dibuat oleh tugas dokumentasi ini.

**Tidak ada secret dalam Git maupun hasil, meskipun folder private.** Redact sebelum stdout/stderr ditulis atau ditampilkan: DSN/password, Authorization headers, client/API keys, cookie/session/CSRF, OIDC code/token/state, presigned URL query, webhook secret, sensitive payload/PII. Jangan membuat environment dump, shell tracing, raw HAR, screenshot/token response, browser storage state, unredacted traces atau backup plaintext di folder hasil. Credential sementara tetap di memory/secret storage private terpisah dan dihapus setelah run. Backup sensitif disimpan terenkripsi di storage yang disetujui; evidence hanya reference+integrity summary teredaksi.

Jika raw output telanjur bocor: hentikan distribusi, batasi akses, informasikan owner keamanan, revoke/rotate melalui admin yang berwenang, lalu perbaiki redaction; menghapus teks saja bukan pemulihan credential. Upload CI artifact/public report tidak otomatis diizinkan. Retention hasil, akses reviewer, dan penghapusan aman masih perlu keputusan owner.

Satu checkbox hanya ditutup setelah kasusnya dijalankan sesuai scope, bukti teredaksi tersedia, dan reviewer menyetujui hasil pada revisi tersebut. Simpan detail run pada evidence private; checklist tracked hanya boleh menyimpan ringkasan aman/referensi run, bukan credential/host internal sensitif. Kasus blocked tetap `[ ]` dengan owner dan next action. Implementasi test file atau hasil fixture tidak boleh menutup acceptance live. Perubahan source, policy, versi provider, schema, atau target berarti evaluasi ulang bukti yang relevan.
