# Alur data dan jaminan kegagalan

> **Usulan Phase 0.** Alur menggambarkan kontrak yang akan diuji, bukan hasil integrasi hidup. PostgreSQL adalah business truth; Temporal adalah satu-satunya orkestrator. Nama operasi bersifat konseptual, bukan endpoint vendor yang diklaim ada.

## 1. Login backend dan sesi opaque

```mermaid
sequenceDiagram
    actor Human as Manusia
    participant Browser
    participant API as Go API
    participant DB as PostgreSQL bisnis
    participant IdP as Authentik
    Human->>Browser: Login
    Browser->>API: Mulai login
    API->>DB: Simpan transaksi login single-use, state/nonce/PKCE dan TTL
    API-->>Browser: Redirect authorization request, PKCE S256
    Browser->>IdP: Autentikasi dan MFA sesuai policy provider
    IdP-->>Browser: Redirect callback dengan code dan state
    Browser->>API: Callback
    API->>DB: Consume state secara atomik, cek expiry dan browser binding
    API->>IdP: Tukar code dengan PKCE verifier server-side
    IdP-->>API: Token response backend
    API->>API: Validasi issuer, audience, signature, nonce, exp dan claims
    API->>DB: Resolve issuer+subject, invitation dan membership default-deny
    alt User berizin
        API->>DB: Buat/rotasi hash session dan catat audit aman
        API-->>Browser: Secure HttpOnly opaque cookie dan redirect aman
    else Identitas atau policy tidak valid
        API-->>Browser: Tolak tanpa membuat akses aplikasi
    end
```

Redirect/return URL harus allowlist lokal; state browser binding mencegah login CSRF. Cookie produksi host-only Secure/HttpOnly, SameSite Lax atau lebih ketat jika kompatibel; CSRF token + Origin check untuk request mutasi, expiry idle/absolute, revocation server-side, rotasi setelah login/privilege change. OIDC token tidak dikirim ke JavaScript/localStorage. Logout aplikasi mencabut sesi lebih dahulu; logout provider bergantung konfigurasi yang diverifikasi dan tidak dijanjikan global single logout.

## 2. Tugas, outbox, Temporal dan hasil agen

```mermaid
sequenceDiagram
    actor User as User berizin
    participant API as Go API
    participant DB as PostgreSQL bisnis
    participant Relay as Outbox relay
    participant TM as Temporal
    participant W as Go worker Activities
    participant HG as Native Hermes adapter
    User->>API: Create/assign task dan idempotency key
    API->>API: Membership, permission, brand grant dan input policy
    API->>DB: TX task/run queued + audit/event + outbox
    DB-->>API: Commit
    API-->>User: Resource dan operation ID, belum tentu running
    Relay->>DB: Claim committed event dengan lease
    Relay->>TM: Start workflow ID stabil atau signal event ID
    TM-->>Relay: Ack atau workflow already exists
    Relay->>DB: Mark delivery, duplikasi tetap mungkin
    TM->>W: Activity build context dan dispatch
    W->>DB: Revalidasi task, grant user/agent dan simpan context exact versions
    W->>DB: Reserve external operation unik dan request hash
    W->>HG: Invoke profil terpetakan, task/run/correlation dan minimum context
    HG-->>W: Receipt atau outcome unknown sesuai kontrak terbukti
    W->>DB: Simpan receipt/state/event, bukan asumsi sukses
    Note over HG,W: Callback autentik atau polling hanya jika didukung dan diverifikasi
    HG-->>W: Hasil, status, correlation dan event identity
    W->>DB: TX inbox dedupe, lock run, validate schema, versioned deliverable, audit dan outbox
    User->>API: Baca status task
    API->>DB: Query task dengan org, brand dan permission yang diotorisasi
    DB-->>API: State bisnis dan versi deliverable yang boleh dibaca
    API-->>User: Status teramati dan hasil tervalidasi
```

User membaca melalui API scoped, bukan akses DB langsung. Callback HTTP sebenarnya melalui inbound adapter yang mengautentikasi dan memanggil use case sama; polling dilakukan Activity dengan policy timeout/rate limit. Tidak mengklaim mekanisme itu sudah tersedia untuk keenam profil. Jika hanya subset terbukti, capability lainnya disabled dan tidak tampil ready.

Activity timeout tidak berarti agen gagal atau belum bekerja. Bila native operation tidak mendukung idempotency/query hasil, jangan melakukan resend otomatis setelah ambiguous outcome. Rekonsiliasi atau human escalation dibutuhkan. Cancellation bisnis dicatat terpisah dari konfirmasi agen berhenti. Out-of-order hasil run superseded/cancelled disimpan sebagai late event tanpa menghidupkan task kembali. Workflow history membawa ID/ref, bukan raw context/token.

## 3. Upload dan authorized context

```mermaid
sequenceDiagram
    actor User
    participant API as Go API dan use cases
    participant DB as PostgreSQL metadata
    participant Obj as MinIO private
    participant Check as Inspector atau scan adapter
    participant Context as Context builder Activity
    User->>API: Minta upload untuk brand yang berizin
    API->>DB: Create upload session dan quarantine object key server-side
    API-->>User: Mekanisme upload berukuran/TTL terbatas
    User->>Obj: Upload hanya melalui grant sementara atau proxy API
    User->>API: Finalize upload dengan upload ID
    API->>Obj: Verifikasi existence, checksum, size dan file signature
    API->>Check: Validasi MIME/extension/actual type dan scan sesuai policy
    alt Berkas lolos kebijakan distribusi
        API->>DB: Simpan asset version immutable dan clean/accepted
    else Gagal atau belum terbukti aman
        API->>DB: Rejected atau tetap quarantine, tidak dapat dipublish
    end
    Context->>DB: Resolve task scope dan grant, query approved relevant versions
    Context->>DB: Persist manifest snapshot source/version dan policy
    Context->>Obj: Ambil exact version hanya setelah authorization
    Context-->>Context: Batasi context, label provenance/draft dan data tidak tepercaya
```

Presigned upload bukan public bucket; finalize tidak mempercayai metadata browser. Batas size/type/checksum diverifikasi server, URL singkat tidak disimpan permanen, private downloads mengecek grant baru. Pilihan proxy upload vs presigned bergantung jaringan browser, CORS dan capability MinIO terverifikasi. Aset belum clean tidak masuk preview bersama, context agen, atau publishing. Search awal ordinary indexed search; vector retrieval bukan default.

## 4. Exact approval, konfirmasi manusia dan Postiz

```mermaid
sequenceDiagram
    actor Editor
    actor Approver as Approver manusia
    actor Publisher as Publisher manusia
    participant API as Go use cases
    participant DB as PostgreSQL bisnis
    participant TM as Relay dan Temporal
    participant W as Publishing Activity
    participant P as Postiz adapter terverifikasi
    Editor->>API: Submit content version dan exact asset/channel/schedule intent
    API->>DB: TX immutable intent/manifest + approval request + audit
    Approver->>API: Review dan approve exact intent
    API->>DB: TX permission/brand checks, decision dan snapshot binding
    Publisher->>API: Confirm publish exact approved intent, idempotency key
    API->>DB: TX lock intent/approval, revalidate, job+confirmation+audit+outbox
    DB-->>API: Commit, bukan berarti externally scheduled
    TM->>W: Start job workflow ID stabil
    W->>DB: Revalidasi manusia/grant, approval aktif, clean assets dan intent hash
    W->>DB: Reserve unique external operation
    W->>P: Submit approved confirmed immutable intent
    alt External receipt dapat dibuktikan
        P-->>W: External ID dan observed status
        W->>DB: Simpan receipt/status dan audit
    else Timeout atau hasil ambigu
        P-->>W: Outcome unknown
        W->>DB: Mark unknown, jangan blind retry
        W->>P: Reconcile status hanya dengan capability lookup yang terbukti
    end
```

Approval tidak otomatis dispatch. Material edit membuat content/asset/intent baru; decision lama tidak berpindah. Revoke/privilege change diperiksa sebelum submit, tetapi race setelah penerimaan provider membutuhkan cancel/update capability atau eskalasi, bukan klaim transaksi lintas sistem. Postiz auth, endpoint, webhook signature, retry/idempotency, platform format dan status API harus dibuktikan terhadap versi aktual; diagram tidak mengarang endpoint. Webhook/polling hasil melewati authenticated inbox/dedupe dan policy seperti alur tugas. Tidak mengulangi submit ketika outcome unknown kecuali bukti rekonsiliasi mengizinkan.

## 5. Ingestion metrik dan embed Superset

```mermaid
sequenceDiagram
    participant TM as Temporal workflow
    participant W as Metrics Activity
    participant S as API platform berizin
    participant DB as PostgreSQL bisnis dan analytics views
    actor User
    participant API as Go API
    participant BI as Superset
    TM->>W: Sync window dan source scope
    W->>DB: Resolve connection/capability/cursor dan authorization
    W->>S: Fetch window/cursor, hormati rate limit
    S-->>W: Source observations, dimensions dan pagination
    W->>DB: TX idempotent upsert series/period, provenance, cursor dan sync status
    User->>API: Minta akses dashboard dan brand
    API->>DB: Membership, permission, brand grant dan dashboard allowlist
    API->>BI: Issue guest token singkat dengan resource dan RLS buatan server
    BI-->>API: Guest token scope terbatas
    API-->>User: Embed configuration dan token sementara
    User->>BI: Embed sesuai domain allowlist dan expiry
    BI->>DB: Read curated views dengan read-only role dan RLS yang diuji
    BI-->>User: Visualisasi dan freshness/limitation
```

Browser tidak menentukan SQL RLS, token admin atau arbitrary dashboard ID. Token guest hanya resource dan brand yang dihitung server, expiry singkat, renewal kembali melalui authorization. Source ingestion bukan fungsi Superset. Metrik unavailable tetap missing, bukan nol; definisi/grain/organic-paid/attribution/timezone dipertahankan. Reports membedakan measured fact dari interpretasi AI.

## Kontrak recovery lintas alur

| Titik kegagalan | Respons desain |
|---|---|
| DB rollback sebelum outbox commit | Tidak dispatch; client dapat retry command dengan key sama |
| Relay crash sesudah start sebelum ack | Redelivery dengan workflow/event ID sama; tidak create run/job baru |
| Activity crash setelah side effect | Ledger pending/unknown; query/reconcile, tidak mengasumsikan belum dieksekusi |
| Duplicate/out-of-order callback | Inbox unique key, correlation mapping, lock/version transition; terminal state tidak mundur |
| Approval atau grant dicabut | Tolak side effect baru; yang sudah diterima provider harus direkonsiliasi/cancel sesuai capability |
| Worker/Temporal restart | History/replay dan outbox memulihkan proses; PostgreSQL tetap business truth |
| Restore dari backup | Dispatch ditahan, pulihkan DB/object/key konsisten, reconcile provider sebelum resume |

Semua jalur wajib safe error/log redaction, correlation ID, scoped ACL dan audit. Reliability tests Phase 3/5/7 harus menghasilkan bukti nyata; diagram ini bukan pengganti uji.
