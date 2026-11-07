# ADR-0002 — Autentikasi manusia dan sesi backend

- **Status:** Proposed / menunggu persetujuan owner; konfigurasi dan capability Authentik aktual belum dibuktikan di sini.
- **Dasar:** [Phase 1](../../PHASE-01-APPLICATION-FOUNDATION.md), brand isolation [Phase 2](../../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md), security review [Phase 7](../../PHASE-07-PRODUCTION-HARDENING-AND-OPERATIONS.md).

## Konteks

Autentikasi provider tidak cukup untuk mengotorisasi organisasi/brand. Aplikasi perlu menolak identitas yang tidak diundang, merevokasi sesi, melindungi request mutasi terhadap CSRF dan membatasi credential dari JavaScript, log dan workflow history. Kemampuan MFA/passkey/logout serta reverse proxy Authentik harus diverifikasi sesuai instance/client aktual, bukan diasumsikan dari nama produk.

## Keputusan yang diusulkan

1. Authentik sebagai OIDC provider; **Authorization Code Flow + PKCE S256**, backend menjadi pemilik callback/code exchange dan validasi token. State, nonce, browser binding dan authorization transaction single-use dengan expiry wajib ada.
2. Backend memvalidasi issuer, signature/alg allowlist/JWKS, audience dan authorized party bila relevan, nonce, expiry/time skew serta claims yang diperlukan. Redirect URI exact match dan return URL allowlist; tidak memakai email sebagai primary identity key.
3. Mapping memakai `(issuer,subject)` ke user aplikasi, lalu membership organisasi aktif. **Default deny** user baru. Invitation/provisioning terbatas harus disetujui, berumur pendek, single-use dan terikat identity/email verified sesuai policy; claim group tidak otomatis menjadikan Owner/Administrator.
4. Browser menerima **opaque random session cookie HttpOnly**, host-only, Secure di produksi, SameSite=Lax atau lebih ketat jika callback kompatibel. Database menyimpan hash/HMAC token, bukan cookie bearer asli. OIDC client secret/refresh token tidak pernah dikirim ke browser atau localStorage.
5. Session idle/absolute expiration, rotation saat login/privilege/org change dan revocation server-side wajib ada. Authz revision atau validasi DB memastikan grant dicabut tidak menunggu cookie kadaluarsa. Cookie expiration saja tidak dianggap logout aman.
6. Request mutasi perlu CSRF token yang terikat sesi, pemeriksaan Origin/Referer sesuai policy dan CORS same-origin/allowlist sempit. SameSite bukan satu-satunya proteksi CSRF. GET tidak menimbulkan perubahan bisnis.
7. Logout aplikasi mencabut sesi dan membersihkan cookie; provider end-session/front/back-channel logout hanya diaktifkan jika perilakunya terbukti. Tidak mengklaim logout aplikasi mengakhiri seluruh sesi IdP.
8. Role awal: Owner, Administrator, Marketing Manager, Editor, Approver, Analyst, Viewer. Use case dan query memerlukan org membership + permission + brand grant. Frontend nav visibility bukan kontrol keamanan. Bootstrap Owner lewat proses terulas/terbatas dengan audit, bukan first-login-admin.
9. Sensitif seperti publish, social account change, spending, brand identity dan destructive external operation memerlukan konfirmasi manusia; step-up/reauth policy dan batas self-approval ditetapkan owner sebelum fitur aktif.

## Alternatif

| Alternatif | Pertimbangan |
|---|---|
| Password lokal aplikasi | Menambah penyimpanan credential dan recovery padahal Authentik dipilih |
| Frontend OIDC token/localStorage | Menambah exposure bearer/refresh token; backend cookie dipilih |
| JWT sesi self-contained tanpa server revocation | Revocation/privilege change lebih rumit; PostgreSQL session authoritative lebih sederhana awal |
| Auto-provision semua user provider | Melewati invitation dan brand ACL; bertentangan default deny |
| Group provider langsung menentukan semua brand | Tidak merepresentasikan grant bisnis yang granular dan diaudit |
| Cukup middleware atau filter UI | Worker/query/file/analytics bisa melewati otorisasi; policy wajib pada use case dan repository |

## Konsekuensi dan risiko

Cookie membutuhkan CSRF dan same-origin policy yang benar; OIDC callback membutuhkan transient state store yang terlindungi dan anti-replay. PostgreSQL session lookup menambah read path; cache opsional harus fail-safe dan revocation-aware. Logout provider mungkin terbatas; UI/runbook harus membedakan sesi aplikasi dan IdP. OIDC outage tidak boleh membuka bypass login; sesi existing mengikuti expiry/revocation policy yang disetujui.

Token guest Superset adalah pengecualian terbatas untuk SDK embedding: singkat, hanya dashboard/row scope yang disetujui backend, tidak long-lived OIDC/provider token. Presigned object URL juga bearer terbatas dan tidak boleh dicatat/log/share sebagai link permanen.

## Gate verifikasi dan keputusan owner

- Buktikan issuer/discovery/JWKS/client type, PKCE S256, redirect HTTPS, claims, clock, logout behavior dan proxy header trust pada Authentik aktual sebelum menyebut integrasi siap.
- Owner menyetujui organisasi awal/identity bootstrap, invitation binding, role/permission matrix, semua-brand access eksplisit bila perlu, idle/absolute timeout, step-up/MFA/passkey dan four-eyes approval. Sampai ada keputusan, operasi sensitif default deny.
- Phase 1 menguji state/nonce/code replay, issuer/audience/signature salah, CSRF/Origin, fixation/rotation/revocation, invitation misuse, open redirect, cross-org/brand dan privilege escalation; Phase 7 mengulang pada deployment HTTPS sesungguhnya.
- Jangan log cookie, code, token, verifier, password atau raw user claims. Audit hanya event keamanan dan metadata teredaksi; retensi PII disetujui owner.

Lihat [alur login](../architecture/data-flow.md#1-login-backend-dan-sesi-opaque), [database sessions/RBAC](../06-database-design.md), dan detail keamanan workstream `docs/08-authentication-security.md` saat tersedia. ADR ini tidak membuat client/provider atau mengubah konfigurasi produksi.
