# ADR-0001 — Arsitektur awal Go dan React

- **Status:** Proposed / menunggu persetujuan owner Phase 0.
- **Ruang lingkup:** bentuk aplikasi, data utama, batas integrasi dan deployment bertahap.
- **Dasar:** [Phase 0](../../PHASE-00-DISCOVERY-AND-ARCHITECTURE.md), [Phase 1](../../PHASE-01-APPLICATION-FOUNDATION.md), seluruh urutan fase pada [overview](../00-project-overview.md). Keputusan ini belum diimplementasikan.

## Konteks

Baseline repositori `650134c` berisi rencana fase, bukan aplikasi. Produk multi-brand membutuhkan identitas manusia, enam agen existing, kerja tahan-gagal, versioned content, approval manusia, publishing dan analytics. Infrastruktur serta kemampuan API harus diverifikasi; tidak ada alasan awal untuk memisahkan setiap modul sebagai microservice. User memilih Go + React/TypeScript; proposal backend FastAPI tidak dipakai.

## Keputusan yang diusulkan

1. Gunakan **Go modular monolith, Hexagonal Architecture**, thin HTTP handlers, domain bebas framework/SQL/vendor dan application use cases dengan ports eksplisit. Composition root terpisah API dan worker.
2. HTTP memakai `net/http` dan router standar yang sesuai versi Go terpilih. Persistence memakai PostgreSQL + `pgx` + `sqlc` + migrasi versioned; transaksi, scoped SQL dan outbox eksplisit.
3. Frontend **React, Vite, TypeScript strict, Tailwind** sebagai pilihan awal, komponen aksesibel dan validasi konsisten. Build UI dilayani Go bersama `/api` pada origin yang sama; rahasia dan OIDC token tidak masuk frontend.
4. PostgreSQL adalah authoritative business truth; MinIO untuk binary privat. Temporal adalah satu-satunya orkestrator durable; vendor persistence/database roles terpisah. Redis opsional bila kebutuhan cache/transient terbukti, bukan critical queue/session truth.
5. Authentik, Hermes, Postiz, Superset dan sumber metrik berada di balik adapter; aplikasi tidak mengandalkan tabel internal vendor. Tidak ada enam agen LLM baru atau perubahan profil `.hermes`.
6. Desain tenant/brand mencakup query scope, composite FK, permission + brand grant dan worker context. Provider callbacks/output/dokumen adalah input tidak tepercaya.
7. Docker Compose dievaluasi untuk dev dan produksi setelah izin serta capacity gate. Tidak memasang Docker atau mengubah host saat ini. Target domain `menerjangbadai.online`, alias `www`, dan preservasi rute dashboard 9119→loopback9120 harus diuji sebelum cutover.

## Alternatif yang dipertimbangkan

| Alternatif | Alasan tidak dipilih sekarang |
|---|---|
| FastAPI backend | Tidak sesuai stack yang dipilih user dan fase; menambah runtime utama kedua tanpa kebutuhan |
| Microservice per capability | Biaya operasi, transaksi dan isolation lebih besar sebelum scale/ownership membenarkan |
| ORM luas sebagai boundary domain | Berisiko menyembunyikan scoping/transaksi; SQL eksplisit dengan sqlc lebih cocok proposal ini |
| Event sourcing atau Temporal sebagai business database | Bertentangan dengan PostgreSQL business truth; audit/events tetap ada tanpa menjadikan replay satu-satunya state |
| Redis queue dan scheduler sendiri | Menggandakan orkestrasi/recovery yang ditangani Temporal dan outbox |
| SPA mengelola token OIDC | Memperluas exposure token; dipilih backend session cookie dan CSRF |
| RAG/vector database sejak awal | Ordinary indexed search belum dievaluasi; belum ada bukti kebutuhan/ACL evaluation |

## Konsekuensi

Positif: satu model bisnis, transaksi lokal lebih jelas, domain/adapters dapat dites, API dan worker dapat diskalakan terpisah. Biaya: disiplin import/interface, pemeliharaan query dan kontrak, compatibility migration API-worker, serta kapasitas stack eksternal perlu dikelola. Process worker terpisah bukan alasan membuat akses DB atau otorisasi paralel yang melewati use case.

Monolith tidak berarti seluruh vendor harus ditempatkan pada VPS sama. Authentik/Temporal/Postiz/Superset memiliki persistence dan operational overhead sendiri; co-location hanya bila capacity assessment nyata membuktikan aman. Kurang kapasitas adalah alasan memindahkan layanan, bukan menghapus gate keamanan.

## Syarat penerimaan dan peninjauan ulang

- Owner menyetujui blueprint dan implementasi Phase 1. Versi Go/Node, library, migrator, tooling frontend dan package manager dipilih/pin melalui validasi aktual.
- Phase 1 membuktikan build/run nyata, migrasi kosong, authentication/authorization/CSRF dan secret scan. Belum ada hasilnya dalam ADR ini.
- Perubahan ke microservice/ORM/router lain/RAG/Redis critical memerlukan alasan terukur, ADR pengganti dan review isolation/recovery.
- Semua acceptance Phase 1–6 dan hardening/restore/load Phase 7 lulus sebelum production readiness disetujui.

## Izin yang tidak diberikan ADR ini

Tidak ada izin deploy, memasang Docker, membuat database, mengubah credential, proxy, port atau profile Hermes. Proposal memindahkan Go ke loopback:2023 dan Caddy ke 80/443 memerlukan cutover approval owner terpisah setelah readiness Phase 7. Merge owner ke `main` hanya trigger rilis aplikasi eligible, bukan auto-deploy dokumen atau izin produksi sekarang.

## Artefak terkait

[Architecture overview](../02-architecture-overview.md), [Hexagonal](../03-hexagonal-architecture.md), [struktur](../04-project-structure.md), [deployment](../architecture/deployment-diagram.md), [database](../06-database-design.md), [ADR workflow](ADR-0004-workflow-orchestration.md).
