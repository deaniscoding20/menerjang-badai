# Rencana eksekusi dan acceptance gates

## Aturan pelaksanaan

Dokumen `PHASE-00-*` sampai `PHASE-07-*` di root adalah spesifikasi otoritatif proyek. Phase0 meminta blueprint direview dan disetujui sebelum implementasi besar. PR ini mengajukan approval itu, bukan menyatakan delapan fase selesai. Merge owner atas PR blueprint menyatakan persetujuan desain yang tertulis, bukan bukti provisioning/integrasi yang belum ada. Blocker teknis tetap terbuka setelah merge sampai verifikasi nyata lulus.

Urutan PR: Phase0 blueprint + CI foundation → Phase1 identity foundation → Phase2 brands/agents → Phase3 workflow → Phase4 content → Phase5 publishing → Phase6 analytics → Phase7 hardening/cutover. Pecah fase besar menjadi vertical slices reviewable; setiap slice membawa test RED→GREEN, docs dan evidence. Jangan horizontal membangun semua UI lalu mengganti backend dengan fixtures.

Status awal seluruh phase implementation: **belum diimplementasikan**. Blueprint Phase0: diajukan, discovery external belum lengkap. CI policy/checks adalah satu-satunya kode executable pada PR awal; tidak ada app demo yang menyamar sebagai integrasi aktif.

## Phase 0 — Discovery dan architecture

- Scope: snapshot VPS/repos/network/backup; capability Hermes/Postiz/Auth; desain domain/boundaries/ERD, capacity, risiko dan ADR.
- Evidence tersedia: [lingkungan](01-environment-discovery.md), [arsitektur](02-architecture-overview.md), [Hermes](09-hermes-integration.md), [Postiz](10-postiz-integration.md), [security](08-authentication-security.md).
- Gate masih terbuka: actual Postiz version/API, provision Authentik OIDC, live Hermes security subset, keputusan kapasitas dan owner blueprint approval.
- Acceptance: setiap capability diberi verified/documented/unverified/unavailable, bukan semua ditandai lulus karena docs upstream ada. Phase0 belum complete selama actual integration discovery tidak terpenuhi.

## Phase 1 — Foundation, identity dan security

- Go modular monolith ports/adapters; HTTP/bootstrap/config/log/request-ID/shutdown, health liveness dan readiness berbeda.
- PostgreSQL migrations + sqlc generated queries; org/user/roles/permissions/brand ACL/sessions/audit schema; least privilege migration/app DB roles.
- Authentik Authorization Code+PKCE S256; state/nonce/signature/iss/aud/claims; opaque cookie, CSRF, rotation/idle+absolute expiry/revocation/logout.
- React/TS strict shell: login, navigation/profile, responsive/loading/empty/error/denied/expired states; tidak ada auth bypass demo.
- Tests CI wajib ditambah: Go unit/race/vet/format, frontend lint/typecheck/unit/build, migration clean-DB + rollback/forward compatibility, DB integration ACL/session/audit, OIDC issuer sandbox, Playwright login/logout/denied.
- Exit gate: manusia permitted benar-benar login via Authentik, invalid callback ditolak, CSRF+revoke+cross-brand tests lulus, Compose start teruji. Tanpa deployed provider, jangan klaim acceptance selesai.

## Phase 2 — Brands dan agent registry

- CRUD/archive brand, profile/identity/products/audience/positioning/tone; member assignment; setiap query dan command enforce org+brand.
- Seed idempotent enam agent memakai stable UUID aplikasi, bukan username sebagai ID: Claudia/Jordan/Kobe/Curry/Sabrina/Kyrie. Runtime external mappings diverifikasi; disabled dan unknown bukan online.
- UI list/detail brand/member/agent/org; visual responsibilities bukan mandatory routing chain.
- Tests: unauthorized CRUD/membership, cross-org+brand denial, compound FK, seeding twice, disabled agent, stale heartbeat/error/unknown interpretation, no profile writes.
- Exit gate: names/roles/usernames benar; ACL backend lulus; existing Hermes profiles unchanged.

## Phase 3 — Tasks, Hermes dan Temporal

- Gate: verified invocation/result/correlation/auth/timeouts/cancel/retry per profile; implement hanya capability lulus.
- Task lifecycle mencakup draft/ready/assigned/queued/running/waiting-input/waiting-approval/completed/failed/cancelled; dependencies/comments/revision/history/deliverables.
- Transactional outbox dan idempotent consumers; Temporal sole owner workflow execution, PostgreSQL business truth. Unknown external outcome → reconcile, bukan blind retry.
- Structured versioned task/result schema; raw output sensitif disimpan terkontrol; invalid deliverable tidak otomatis valid.
- UI Kanban/list/detail/dependency/timeline/retry+cancel permissions; direct specialist assignment tidak harus melalui Claudia.
- Tests: live authorized bridge subset, rejected scope, malformed result, duplicate/redelivery, out-of-order callbacks, restart recovery, timeout/cancel caveat, safe retry budget, Temporal replay/worker restart, no source/session DB writes.
- Exit gate: task nyata correlated ke profil benar; result versioned; recovery teruji; semua action audited.

## Phase 4 — Knowledge, assets, campaigns dan content

- Metadata version/owner/source/approval/ACL; approved-context retrieval menyertakan citation/version dan menolak brand lain.
- MinIO private originals + Postgres metadata; short-lived presigned URLs; upload file signature/type/size/ext + quarantine/malware policy; expiry bukan pengganti ACL.
- Indexed search dulu; vector/RAG hanya ADR+evaluation jika perlu.
- Campaign objectives/channels/currency+planned budget/KPIs/tasks; content copy/media/target/version; versioned assets immutable references.
- Calendar filter brand/campaign/channel/status/date; planned bukan scheduled externally.
- Tests: cross-brand document/object retrieval denial, malicious upload/quarantine, presigned TTL, immutable approved version, context source/ACL, campaign/date/currency validation.
- Exit gate: MinIO upload/download nyata, access teruji, version history preserved, calendar labels jujur.

## Phase 5 — Human approval dan Postiz

- Approval exact content version + asset versions + brand + destination + schedule/intent + approver/time/comments.
- Material edits invalidate approval; backend checks at job creation dan immediately before external side effect. Agent identity tidak berhak approve/publish.
- Postiz only verified deployed endpoints/version; media/account mapping/status/rate limit; authenticated webhook hanya jika tersedia, polling fallback resmi.
- Idempotency, ambiguous result reconciliation, no auto-retry uncertain publish; external IDs disimpan, failure visible.
- Tests: stale/replayed approval, wrong role/brand, same-version asset replacement, channel/schedule edit, duplicate/reordered events, Postiz timeout with created post, forged webhook, exact payload and human intent.
- Exit gate: sandbox approved publish teruji melalui API resmi; unapproved publish denied; tidak ada publikasi nyata tanpa explicit authorization owner.

## Phase 6 — Analytics dan intelligence

- Platform-specific approved API source, scopes/account/app-review/rate-limits/definitions/history/retention, idempotent cursors+sync windows.
- Normalize unit/timezone/attribution/organic-paid without losing source dimensions. Missing metrics null+reason; jangan ROI jika spend/revenue tidak ada.
- Superset is BI, not collector; backend guest tokens bounded dashboard+brand, RLS verified; admin credentials never browser.
- Competitor research evidence/date/source/version/limitation; social listening hanya approved API, observed vs AI interpretation jelas.
- Tests: fixture contracts clearly labeled vs actual API validation, cursor redelivery, missing metrics vs zero, source timezone, rate limit/recovery, Superset cross-brand RLS/token expiry, report provenance.
- Exit gate: metric sources nyata, freshness visible, authorized embed teruji. Unavailable platform tampil unavailable, bukan random data.

## Phase 7 — Hardening dan operations

- Review all security/session/ACL/CSRF/CORS/XSS/uploads/SQL/secrets/webhooks/rate-limits/logs/container/network/dependencies.
- Pin images/minimal privileges/internal networks/resource limits/health/restart; admin services not public.
- Restore rehearsal PostgreSQL/MinIO/config/key recovery, offsite retention/RPO/RTO owner-approved.
- Logs/correlation/metrics/traces with redaction; real dependency health and provider token/cost if available.
- Reliability: outages/restarts/Temporal recovery/Hermes malformed+timeout/Postiz uncertain+reconcile/redelivery/social rate limit; perf safe workload targets owner-approved.
- CI add vulnerability/image scan, container smoke, migration validation, restore test, frontend E2E/accessibility, target load tests and deploy/rollback rehearsal before readiness marker.
- Exit gate: no critical unresolved finding, restore demonstrated, runbooks exercised, readiness report owner-approved. Only then `production_ready=true`, protected production gate enabled, non-production rehearsal and owner merge main authorize deployment.

## Target PR review format

Untuk setiap PR: scope + phase, acceptance criterion IDs/checklist, exact commands and real results, fixtures vs live distinguished, migrations/rollback compatibility, security implications, remaining blockers. Owner only merge. Never enable auto-merge on owner's behalf.
