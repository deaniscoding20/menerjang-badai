# MenerjangBadai — ikhtisar proyek

> **Status: usulan blueprint Phase 0; menunggu persetujuan owner.** Dokumen ini bukan laporan implementasi, bukti integrasi hidup, atau izin mengubah produksi. Baseline repositori yang dibaca: `650134c`, berisi delapan dokumen fase dan belum berisi aplikasi. Seluruh `PHASE-00` sampai `PHASE-07` telah dibaca sebagai sumber persyaratan.

## Tujuan dan batas produk

MenerjangBadai merupakan kantor pemasaran AI multi-brand: manusia mengelola brand dan kampanye, memberi tugas kepada enam profil Hermes yang sudah ada, meninjau hasil, menyetujui versi konten tertentu, dan mengonfirmasi penerbitannya. Data performa hanya berasal dari sumber yang diizinkan dan diverifikasi. AI memberi rekomendasi; tidak memperoleh kewenangan implisit untuk menerbitkan, membelanjakan anggaran, mengganti identitas brand, atau melakukan operasi eksternal destruktif.

Lingkup bertahap: identitas dan otorisasi → brand dan registry agen → tugas dan workflow → pengetahuan, aset, kampanye dan konten → approval dan Postiz → analitik → hardening. Ini bukan penggantian Hermes, pembuat enam agen LLM baru, atau sistem pengiklanan otomatis tanpa persetujuan manusia.

## Prinsip yang mengikat proposal

1. Backend **Go**, frontend **React + TypeScript strict**; bukan FastAPI atau backend Python kedua.
2. Modular monolith dengan Hexagonal Architecture. API dan worker boleh menjadi proses berbeda, tetapi berbagi domain/use case; belum perlu microservices.
3. PostgreSQL adalah sumber kebenaran bisnis. Temporal satu-satunya orkestrator workflow tahan-gagal; history Temporal bukan pengganti catatan bisnis.
4. Transaksi bisnis, audit penting, dan outbox dicatat atomik. Pengiriman minimal sekali membutuhkan konsumen idempoten; tidak ada janji exactly-once lintas layanan.
5. Authentik OIDC Authorization Code + PKCE S256 diproses backend; browser hanya menerima cookie sesi opaque HttpOnly. Provisioning default-deny, RBAC dan brand ACL ditegakkan pada use case serta query.
6. Enam profil Hermes dipetakan ke ID aplikasi yang stabil. Tidak mengubah `.hermes`, prompt, persona, konfigurasi, atau routing mereka.
7. MinIO menyimpan berkas privat; PostgreSQL menyimpan metadata/versi. Approval mengikat konten, media, tujuan, dan maksud publikasi yang tepat.
8. Postiz hanya adapter penerbitan; Superset hanya visualisasi, bukan pengumpul metrik. Redis opsional dan tidak menjadi dependensi correctness.
9. Tidak ada rahasia, cookie, token, payload sensitif, atau URL bertanda tangan dalam log, audit, telemetry, maupun dokumentasi.

## Enam agen yang harus dipertahankan

Nama, peran, dan username berikut adalah persyaratan [Phase 2](../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md), bukan bukti konektivitas.

| Agen | Tanggung jawab | Telegram |
|---|---|---|
| Claudia | Assistant Brand Manager & Coordinator | `@MB_claudia_bot` |
| Jordan | Marketing Strategist | `@MB_jordan_bot` |
| Kobe | Creative Lead | `@MB_kobe_bot` |
| Curry | Market Researcher | `@MB_curry_bot` |
| Sabrina | Social Media & Copywriter | `@MB_sabrina_bot` |
| Kyrie | Creative Designer | `@MB_kyrie_bot` |

ID profil Hermes, jalur invocation, status, pengambilan hasil, timeout, cancellation, retry dan isolation harus diverifikasi per profil. Username Telegram bukan ID Hermes, heartbeat, atau kontrak dispatch. Penugasan langsung ke spesialis didukung; tidak semua pekerjaan wajib melewati Claudia.

## Konteks lingkungan dan batas bukti

Konteks discovery yang diberikan untuk proposal: domain `menerjangbadai.online` melalui HTTPS/proxy menuju port `2023`; Caddy saat ini memakai port tersebut untuk konten `/srv/menerjangbadai/public`; `www` mengarah ke layanan yang sama; rute dashboard `9119 → loopback:9120` harus tetap utuh. Docker belum tersedia. Detail ini **masukan handoff**, bukan inspeksi ulang oleh penulis blueprint. Bukti host, kapasitas, jaringan dan backup dirujuk ke [discovery lingkungan](01-environment-discovery.md); snapshot tersebut melaporkan 2 logical CPU, RAM fisik sekitar 7,45 GiB dan disk root sekitar 78,6 GiB. Angka ini bukan capacity approval untuk seluruh stack; beban puncak, IOPS, growth dan headroom masih harus diukur. Detail integrasi dicatat dalam workstream masing-masing.

Usulan sasaran, hanya setelah izin terpisah: Go mendengarkan `127.0.0.1:2023`, melayani build React dan `/api`; Caddy melayani domain utama melalui 80/443. Jangan berebut bind dengan layanan sekarang. Tidak memasang Docker, mengubah Caddy, membuka port, memindahkan berkas produksi, menjalankan migrasi, atau mengirim konten nyata dalam Phase 0.

## Peta implementasi dan gate

| Fase/sumber persyaratan | Hasil yang direncanakan | Bukti keluar yang harus dihasilkan nanti |
|---|---|---|
| [0 — Discovery](../PHASE-00-DISCOVERY-AND-ARCHITECTURE.md) | Blueprint, discovery, ADR, risiko dan rencana | Review owner; keterbatasan integrasi eksplisit; keputusan infrastruktur belum otomatis disetujui |
| [1 — Foundation](../PHASE-01-APPLICATION-FOUNDATION.md) | Go API, React, migrasi/sqlc, OIDC, sesi, RBAC, audit, lingkungan dev | Callback/CSRF/revocation/isolation negatif; migrasi kosong; build dan startup nyata; izin instalasi dev terpisah bila perlu |
| [2 — Brands/agents](../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md) | Brand, ACL, enam registry agen dan capability evidence | Seed idempoten, mapping stabil, tidak mengubah profil, status bukan rekaan |
| [3 — Workflows](../PHASE-03-HERMES-INTEGRATION-AND-WORKFLOWS.md) | Tugas, run, deliverable, outbox, Temporal, adapter Hermes terverifikasi | Gate invocation/result per profil; timeout, duplicate/out-of-order, restart dan dispatch aman diuji |
| [4 — Knowledge/content](../PHASE-04-KNOWLEDGE-ASSETS-AND-CONTENT.md) | Konteks brand terkontrol, aset/versi, kampanye, konten/versi, kalender rencana | Tidak bocor lintas brand; upload/preview aman; versi immutable; kalender tidak berpura-pura sudah terjadwal eksternal |
| [5 — Approval/Postiz](../PHASE-05-APPROVAL-AND-POSTIZ-PUBLISHING.md) | Approval exact manifest, konfirmasi manusia, publikasi dan rekonsiliasi | API Postiz versi aktual diverifikasi; approval invalidation, permission dan uncertain outcome diuji tanpa posting produksi |
| [6 — Analytics](../PHASE-06-ANALYTICS-AND-MARKETING-INTELLIGENCE.md) | Adapter sumber metrik, ingestion, guest embedding Superset, laporan berbukti | Hak API dan definisi metrik jelas; token server-side/RLS negatif; missing bukan nol; freshness terlihat |
| [7 — Operations](../PHASE-07-PRODUCTION-HARDENING-AND-OPERATIONS.md) | Hardening, load test, restore, observability, runbook, release/rollback | Seluruh acceptance fase sebelumnya lulus; restore nyata, target RPO/RTO disepakati; owner menyetujui readiness dan cutover |

**Approval blueprint Phase 0 tidak sama dengan approval produksi.** Owner memegang merge. Merge owner ke `main` hanya boleh memicu rilis aplikasi yang sudah eligible berdasarkan gate CI, readiness Phase 7 dan izin produksi yang tercatat; dokumen ini maupun merge dokumen tidak mengaktifkan deployment. Switch domain/port memerlukan persetujuan terpisah pada readiness, bukan sekarang.

## Keputusan owner yang diminta

- Menyetujui blueprint dan urutan fase sebelum Phase 1; setujui/ubah ADR-0001 sampai ADR-0004.
- Menetapkan organisasi awal, owner bootstrap, kebijakan invitation, role/brand ACL dan apakah approver wajib berbeda dari penulis/publisher. Default tidak ada akses brand otomatis bahkan untuk Administrator; grant eksplisit tetap diperlukan.
- Memilih co-location atau pemisahan layanan setelah pengukuran CPU/RAM/disk/IOPS, storage growth, beban Hermes/Temporal/Superset dan headroom. Jangan menganggap seluruh stack muat di VPS.
- Menyetujui mekanisme Hermes yang benar-benar dibuktikan, lingkup Postiz/platform, serta sumber analitik yang tersedia; capability yang belum terbukti tetap dinonaktifkan.
- Menetapkan retensi PII, raw output, audit, aset, workflow history, backup; biaya AI, batas upload, RPO/RTO, SLA serta batas beban.
- Menyetujui instalasi/deployment secara terpisah, termasuk proposal port 2023 dan preservasi `www`/dashboard 9119. Tidak ada izin implisit dari dokumen ini.

## Risiko utama dan mitigasi desain

| Risiko | Konsekuensi | Mitigasi/gate |
|---|---|---|
| Invocation Hermes atau hasil tidak mendukung korelasi/idempotensi | Pekerjaan duplikat atau hasil salah task | Bukti kontrak per profil, operation ledger, state `unknown`, jangan retry side effect secara buta |
| Kebocoran brand lewat worker, file, analitik atau konteks LLM | Data rahasia lintas pelanggan | Compound FK, query terikat org/brand, grant agen, context snapshot, use-case ACL dan tes negatif |
| Timeout Postiz setelah berhasil di luar | Posting ganda | Ledger idempoten, klasifikasi ambiguous, rekonsiliasi sebelum retry, human escalation bila tak dapat dibuktikan |
| Approval stale | Versi/media/tujuan salah dipublikasikan | Intent immutable + fingerprint, approval exact snapshot, revalidasi atomik sebelum submit |
| Kekurangan kapasitas atau bentrok listener | Gangguan layanan yang sudah ada | Ukur dulu, segregasi layanan bila perlu, staging, rollback dan izin cutover Phase 7 |
| Superset/Temporal berbagi credential bisnis berlebih | Bypass ACL atau kerusakan data | Database/role terpisah, view analytics read-only, guest RLS dan jaringan internal |
| Dokumen disalahartikan sebagai verifikasi live | Gate keamanan dilewati | Semua diagram berlabel usulan; hasil uji hanya dicatat setelah eksekusi nyata |

## Indeks blueprint

- [Ikhtisar arsitektur](02-architecture-overview.md), [batas hexagonal](03-hexagonal-architecture.md), [struktur rencana](04-project-structure.md).
- [Model domain](05-domain-model.md), [desain database dan ERD](06-database-design.md).
- [Context](architecture/system-context.md), [container](architecture/container-diagram.md), [component](architecture/component-diagram.md), [deployment](architecture/deployment-diagram.md), [data flow](architecture/data-flow.md).
- [ADR awal](adr/ADR-0001-initial-architecture.md), [autentikasi](adr/ADR-0002-human-authentication.md), [agen](adr/ADR-0003-agent-integration.md), [workflow](adr/ADR-0004-workflow-orchestration.md).
- Detail lintas workstream: [discovery](01-environment-discovery.md), [implementation plan](07-implementation-plan.md), [security](08-authentication-security.md), [Hermes](09-hermes-integration.md), [Postiz](10-postiz-integration.md), [CI/CD](11-ci-cd-and-deployment.md) serta laporan readiness harus direview bersama; blueprint ini tidak menyatakan acceptance integrasi/produksi sudah terpenuhi.
