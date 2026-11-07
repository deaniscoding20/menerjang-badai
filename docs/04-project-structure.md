# Struktur proyek yang diusulkan

> **Usulan Phase 0; bukan inventaris file aplikasi yang sudah ada.** Baseline `650134c` hanya memuat delapan `PHASE-*.md`. Tree berikut adalah arah implementasi bertahap setelah approval owner, bukan instruksi membuat scaffold kosong sekarang.

```text
menerjang-badai-dashboard/
├── PHASE-00-...md sampai PHASE-07-...md
├── cmd/
│   ├── api/main.go                 # composition root HTTP + static React
│   └── worker/main.go              # Temporal worker + relay outbox terbatas
├── internal/
│   ├── identity/                  # domain/, application/ sesuai kebutuhan nyata
│   ├── brands/
│   ├── agents/                    # registry/mapping existing Hermes; bukan persona baru
│   ├── work/                      # task/run/dependency/deliverable
│   ├── knowledge/
│   ├── media/
│   ├── campaigns/
│   ├── content/
│   ├── approvals/
│   ├── publishing/
│   ├── analytics/
│   ├── adapters/
│   │   ├── inbound/http/          # route, middleware, DTO, handlers tipis
│   │   ├── inbound/activities/    # delegasi ke application use cases
│   │   ├── outbound/postgres/     # implementasi repo, transaksi, mapping row
│   │   │   └── sqlc/              # generated; tidak diekspos ke domain
│   │   ├── outbound/oidc/         # Authentik client dan provider verification
│   │   ├── outbound/hermes/       # adapter kontrak native yang diverifikasi
│   │   ├── outbound/minio/
│   │   ├── outbound/postiz/
│   │   ├── outbound/superset/
│   │   └── outbound/metrics/      # satu adapter per sumber yang diizinkan
│   ├── workflows/                # Temporal deterministic orchestration
│   └── platform/                 # config, log redaction, DB wiring, clock bila perlu
├── db/
│   ├── migrations/               # schema increment sesuai fase
│   └── queries/                  # SQL terikat org/brand; sumber sqlc
├── api/openapi.yaml              # kontrak API yang dipelihara dan diperiksa
├── web/
│   ├── src/
│   │   ├── app/                  # router, shell, session handling
│   │   ├── features/             # identity, brands, work, content, approvals, analytics
│   │   ├── components/           # komponen aksesibel bersama
│   │   └── lib/api/              # API client dan tipe kontrak, bukan secret
│   ├── public/
│   └── dist/                     # hasil build; bukan source yang diedit
├── deploy/
│   ├── compose/                  # dev dahulu; production setelah gate
│   └── caddy/                    # contoh terulas; tidak menimpa config host otomatis
├── tests/
│   ├── integration/
│   ├── contracts/
│   ├── workflows/
│   └── e2e/
├── scripts/                      # task dev/CI terulas, tanpa rahasia atau auto-cutover
├── docs/
│   ├── 00-project-overview.md
│   ├── 02-architecture-overview.md
│   ├── 03-hexagonal-architecture.md
│   ├── 04-project-structure.md
│   ├── 05-domain-model.md
│   ├── 06-database-design.md
│   ├── architecture/*.md
│   └── adr/ADR-*.md
├── .env.example                  # nama konfigurasi dan placeholder saja
├── sqlc.yaml
├── go.mod
└── go.sum
```

## Aturan penempatan

- Setiap modul memiliki `domain` dan `application` hanya saat tanggung jawabnya muncul. Port dimiliki application; implementasi persistence/vendor berada di adapter, tidak di domain.
- `cmd` hanya bootstrap, dependency injection constructor, lifecycle, signal handling dan shutdown. Tidak memuat SQL atau aturan approval.
- `platform` hanya fasilitas teknis, bukan tempat menaruh use case bersama yang mengaburkan ownership. Hindari circular imports dan service locator global.
- DTO HTTP, format provider dan row sqlc tidak menjadi entity domain. Mapping eksplisit menjaga perubahan API/vendor tidak merembet ke inti.
- Repository queries ditata per modul. Semua akses brand wajib org/brand scope; jangan menyembunyikan SQL global di helper umum.
- `workflows` merupakan adapter runtime Temporal walaupun letaknya top-level internal. Activity handlers tipis memanggil application; workflow tidak mengimpor concrete DB client.
- Build frontend dilayani Go di origin sama dengan `/api`, dengan fallback SPA hanya untuk route UI. `/api`, `/auth`, health, file route dan error tidak boleh jatuh ke HTML fallback. Jangan cache respons sesi/API pribadi sebagai aset publik; hashed assets boleh immutable cache.
- Penggabungan `web/dist` ke binary atau release directory dipilih Phase 1; release harus memasangkan versi UI dan API, bukan membaca source frontend di produksi.
- Unit test ditempatkan dekat package Go/komponen frontend; folder `tests` untuk lintas modul, integration, contract, workflow dan E2E.

## Urutan membuat struktur, bukan scaffold massal

| Fase | Path yang baru layak dibuat |
|---|---|
| 0 | Dokumentasi blueprint/discovery/ADR saja |
| 1 | `cmd/api`, bootstrap/config, identity dan minimal brand foundation, HTTP/Postgres/OIDC adapters, migration/sqlc, web shell, kontrak API, dev environment; worker entry point minimal bila ada tanggung jawab terdefinisi |
| 2 | Brand domain lengkap, agent registry/capability/mapping, UI dan tes isolation |
| 3 | Work, workflows, Activity, relay outbox, Hermes adapter verified dan test restart/duplicate |
| 4 | Knowledge/media/campaign/content, safe upload dan versioned context |
| 5 | Approval/publishing dan Postiz verified subset |
| 6 | Analytics ingestion dan Superset secure embedding |
| 7 | Config production, CI/release gated, observability/runbook/backup/restore/performance |

## Kontrak tooling dan CI yang harus diputuskan di Phase 1

Usulan: `go test`, formatter/linter Go, TypeScript strict/typecheck, frontend unit/E2E, OpenAPI contract check, sqlc generation diff, migrasi kosong/upgrade, dependency/secret scan. Pilih satu package manager frontend dan commit lockfile; jangan campur npm/pnpm/yarn. Pilih migrator beserta strategi expand/contract; down migration destruktif bukan rollback produksi otomatis. Semua command sebenarnya, versi dan dependency didokumentasikan setelah dijalankan, bukan dipalsukan di Phase 0.

Docker Compose adalah target lingkungan dev dan kandidat produksi dari dokumen fase. Karena Docker belum tersedia menurut handoff discovery, jangan memasang atau menjalankan Compose sekarang. Kapasitas, lokasi layanan, service/DB role dan volume baru dirancang rinci setelah pengukuran dan approval. Tidak menyimpan konfigurasi Hermes aktif di repository, tidak membuat enam file prompt sebagai pengganti profil native, dan tidak mengubah direktori `.hermes`.
