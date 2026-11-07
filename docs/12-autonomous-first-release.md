# Otorisasi first release dan credential locality

## Instruksi operator terbaru

Operator terverifikasi **Bos Dean** (`@deaniscoding`, Telegram ID8448343771) memberikan otorisasi saat hendak tidur: development dapat dilanjutkan mandiri, termasuk merge `main` dan deploy pertama ke VPS. Ini menggantikan batas sebelumnya bahwa hanya owner boleh melakukan merge secara manual dan setiap gate harus menunggu balasan chat. Bos Reza (`@jurnalesok`) adalah orang berbeda; username saja bukan grant akses.

Otorisasi berlaku untuk kebutuhan dashboard pada repo/VPS ini, bukan menonaktifkan keamanan global, mengubah persona enam profil, menghapus data, memublikasikan social post, atau membelanjakan iklan. Build/test/security/backup tetap wajib. Jangan menganggap izin ini sebagai bukti integration discovery atau production acceptance yang belum diuji.

## First merge dan first deploy

- Tetap gunakan feature branch, commit author `MB_default_agent`, PR audit trail dan CI. Agent boleh melakukan merge pertama yang lolos pemeriksaan menggunakan otorisasi ini; tidak perlu menunggu user kembali online.
- Tidak menyatakan seluruh Phase1–7 selesai hanya karena dashboard awal runnable. Setiap kemampuan yang belum terkoneksi harus disabled/unavailable dengan alasan.
- Boleh provisioning reversible infrastruktur yang diperlukan di VPS setelah inspection, resource check, dan metadata-preserving backup.
- Sebelum cutover :2023: functional authentication, authorized-data isolation, candidate readiness+commit identity, rollback dan restore test sesuai scope, pemeliharaan dashboard9119/gateway dan TLS. Jika prasyarat keamanan tidak dapat dipenuhi, jangan expose aplikasi publik tanpa auth; tulis blocker.
- Perubahan acceptance readiness untuk rilis pertama terbatas harus berupa ADR eksplisit dan test. Jangan sekadar mengubah phase7/ready=true guna melewati guard padahal acceptance belum dipenuhi.

## Credential tetap pada VPS

Credential produksi disimpan di secret file/store VPS dengan permissions terbatas, tidak ke Git, image layer, browser, Actions artifacts/logs atau GitHub secrets. Penggunaan masked ephemeral GitHub token untuk metadata repo berbeda dari runtime credential produksi.

Workflow SSH dormant yang ada adalah desain lama dan **tidak boleh diaktifkan** dengan menyalin key VPS ke GitHub. Ganti dengan VPS-local deploy puller yang memeriksa main+successful CI+merge proof, exact immutable SHA/image digest, readiness dan lock/idempotency. Gunakan credential GitHub yang memang sudah disetujui dan disimpan VPS-local tanpa mengekstrak nilainya ke output. Jangan menambahkan self-hosted Actions runner untuk mengeksekusi PR arbitrer di host produksi.

Systemd/timer deployment lokal harus restricted, audited, network minimal dan tidak mengakses profile/session Hermes. Hak root hanya untuk narrow installation/cutover yang telah diperiksa; jangan memberi worker generic unrestricted sudo. Runtime secret berbeda untuk app/migration/dependency.

## Kontinuitas pekerjaan

Operator meminta pekerjaan berlanjut selama tidur; jalankan bounded persistent mission dengan single-owner lock, checkpoint dan report. Tidak memakai loop tanpa batas atau restart gateway untuk bertahan. Bila ada blocker credential eksternal atau operasi irreversible, berhenti aman pada scope itu dan catat, bukan fabricate atau meminta persetujuan tidur berulang.

Instruksi terbaru ini harus direkonsiliasi dengan README/CI-CD report yang masih mendeskripsikan desain awal owner-only/manual-approval sebelum merge. Histori instruksi boleh dipertahankan sebagai provenance, tidak dijalankan sebagai constraint yang sudah diganti.
