# Menerjang Badai

Repository untuk pengembangan aplikasi Menerjang Badai.

## Website

Website publik: [www.menerjangbadai.online](https://www.menerjangbadai.online).

Saat ini website menampilkan halaman HTML sederhana dengan gambar. Layanan website berjalan pada port `2023` dan diakses melalui HTTPS dengan sertifikat gratis Let's Encrypt.

> File website yang sudah berjalan belum dimasukkan ke repository ini. Repository ini masih dalam tahap awal; source code aplikasi, pilihan teknologi, dan perintah build akan ditambahkan sesuai kebutuhan pengembangan.

## Mengambil repository

```bash
git clone git@github.com:deaniscoding20/menerjang-badai.git
cd menerjang-badai
```

Akses SSH ke GitHub diperlukan untuk menggunakan URL tersebut. Untuk pengguna tanpa SSH, gunakan:

```bash
git clone https://github.com/deaniscoding20/menerjang-badai.git
```

## Alur pengembangan

1. Tentukan fitur dan kebutuhan aplikasi.
2. Buat branch untuk perubahan yang dikerjakan.
3. Implementasikan dan uji perubahan.
4. Commit dan push perubahan untuk ditinjau.

Push ke repository tidak otomatis mengubah website produksi. Integrasi deployment otomatis belum disiapkan.

## Keamanan

Jangan commit password, token API, private key SSH, atau file `.env` yang berisi kredensial.
