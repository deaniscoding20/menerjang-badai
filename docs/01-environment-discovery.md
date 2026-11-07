# Discovery lingkungan — observasi 2026-11-05 UTC / 2026-11-06 UTC+8

## Scope dan provenance

Inspeksi read-only pada VPS ini, fetch Git, REST GitHub dan HTTP health; tidak ada install service sistem, restart gateway, perubahan firewall, routing Caddy, persona, atau deploy aplikasi. Pemasangan CLI `gh` dilakukan user-local dengan checksum upstream, login disetujui owner. Nilai credential tidak dicatat. Ini snapshot, bukan monitoring terus-menerus atau audit cloud control plane.

## Source dan akses

- Repo: `deaniscoding20/menerjang-badai`, public; default `main`.
- Baseline diperiksa: `650134c405cf9492b3e369d1072e4e802921c5b0` (`edit phase 01`).
- Delapan dokumen ditemukan: Phase 00 prerequisite + Phase 01–07 implementasi. Tidak ada source aplikasi di baseline, workflow Actions, issue atau PR terbuka saat inspeksi.
- Worktree baru branch `feat/mb-default-agent-dashboard`; worktree lama dan untracked files-nya dipertahankan. Tidak ada reset/clean terhadap pekerjaan lama.
- Git SSH deploy key terkonfigurasi dan fetch sukses. Deploy key Git bukan akses PR/secrets API.
- Setelah OAuth user, `gh` memverifikasi akun owner dengan permission ADMIN dan scopes repo/workflow. Credential store headless plaintext di luar repo, mode harus terbatas; jangan disalin ke Actions atau dashboard.
- Identitas author commit yang diminta: `MB_default_agent`; akun autentikasi GitHub tetap owner yang mengotorisasi. Mengubah `user.name` bukan membuat akun GitHub baru.
- Tidak ada Actions secrets atau variables saat pemeriksaan terautentikasi awal.

## Kapasitas dan perangkat lunak

- Ubuntu 24.04.4 LTS; 2 logical CPU.
- RAM fisik sekitar 7,45 GiB; MemAvailable sekitar 6,03 GiB pada snapshot.
- Disk root sekitar 78,6 GiB; ruang bebas sekitar 65,3 GiB.
- Caddy v2.11.8; Git 2.43.0.
- Node v26.12.1/npm12.0.2 tersedia lewat Hermes tools, bukan pin aplikasi baru. Phase 1 perlu pin Node LTS dan Go supported release yang dipilih saat implementasi.
- Go, Docker, Docker Compose, Podman, psql dan redis-server tidak ditemukan pada PATH. Tidak ditemukan unit service/listener aplikasi PostgreSQL, Redis, MinIO, Temporal, Authentik, Postiz atau Superset pada host yang diperiksa. Ini bukan bukti ketiadaan instance remote/cloud; endpoint remote belum diberikan.
- Jangan menginstal seluruh stack sekaligus hanya karena MemAvailable saat idle cukup. Authentik/Postiz/Superset/Temporal plus database dan Hermes pada 2 CPU/~8GB memiliki risiko contention. Tetapkan limit, uji beban, pertimbangkan layanan terpisah/remote atau peningkatan VPS dengan persetujuan biaya.

## Routing dan jaringan yang diamati

Konfigurasi aktif `/etc/caddy/Caddyfile` dibaca, tidak ditulis:

- TLS `menerjangbadai.online` dan `www.menerjangbadai.online` → `127.0.0.1:2023`.
- `http://:2023` saat ini Caddy file server `/srv/menerjangbadai/public`, bind publik. Ini konflik dengan bind aplikasi baru di port yang sama; jangan jalankan keduanya sekaligus.
- Dashboard Hermes terpisah `https://menerjangbadai.online:9119/hermes-dashboard/` → `127.0.0.1:9120`.
- A2A loopback `127.0.0.1:9900`; jangan expose melalui domain publik.
- Listener lain termasuk SSH22, Caddy80/443/2019, DNS lokal dan IDE lokal. Listener temporary milik Hermes tidak otomatis port aplikasi.
- HTTP 200 teramati untuk domain root HTTPS, backend2023, serta dashboard9119. Ini hanya baseline reachability; bukan bukti autentikasi/isolasi app baru.
- UFW inactive dan `nft list ruleset` tidak menampilkan rule. Firewall/cloud security-group eksternal tidak diakses. Tidak ada perubahan firewall, maupun pengulangan tindakan containment historis yang ditolak.

## Backup dan recovery

Backup terlokalisasi pekerjaan Hermes sebelumnya ada tetapi tidak dibaca karena dapat mengandung secret. Tidak ditemukan bukti restore teruji untuk platform baru karena platform/database belum diinstal. Status backup cloud, offsite, encryption-key escrow dan retention belum diverifikasi. Phase 7 membutuhkan restore rehearsal dan owner menyetujui RPO/RTO; keberadaan file backup saja bukan lulus.

## Baseline integrasi

Enam persona/profile Hermes sudah ada dan tidak diubah. Riwayat local A2A regressions bukan sertifikasi runtime terpasang; live isolation/retry/history gate masih terbuka. Telegram connection bukan task execution API.

Authentik issuer/client/redirect belum diprovision; Postiz deployed version/API credential belum tersedia; social scopes dan Superset deployment belum tersedia. Catat kemampuan dokumentasi dan capability deployed secara terpisah pada dokumen integrasi.

## Perubahan infrastruktur yang meminta approval

1. Install Docker/Compose dan pin image/service versions; tidak dilakukan PR ini.
2. Pilihan capacity: VPS upgrade, remote services, atau deployment bertahap; biaya belum disetujui.
3. Authentik issuer/domain/client, MFA dan bootstrap owner; jangan auto-admin akun dari claim email saja.
4. Cutover :2023 dari website statis ke container loopback; backup metadata-preserving Caddy + static files dan rollback teruji terlebih dahulu. Pertahankan :9119 dan routing www yang disetujui.
5. User deploy non-root, narrow forced-command SSH, known-hosts fingerprint terverifikasi, registry credentials dan secret store.
6. Deployment gate production tetap disabled hingga readiness Phase7 dan persetujuan owner; merge Phase0 tidak boleh menimpa website.
