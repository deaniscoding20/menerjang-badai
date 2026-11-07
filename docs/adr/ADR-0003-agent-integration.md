# ADR-0003 — Memetakan enam profil Hermes yang sudah ada

- **Status:** Proposed / menunggu persetujuan owner dan evidence capability per profil.
- **Dasar:** discovery [Phase 0](../../PHASE-00-DISCOVERY-AND-ARCHITECTURE.md), registry [Phase 2](../../PHASE-02-BRAND-AND-AGENT-MANAGEMENT.md), invocation/result gate [Phase 3](../../PHASE-03-HERMES-INTEGRATION-AND-WORKFLOWS.md).

## Konteks

Claudia, Jordan, Kobe, Curry, Sabrina dan Kyrie sudah memiliki profil Hermes native. Aplikasi memerlukan identitas stabil, dispatch, status, hasil, korelasi dan failure visibility tanpa mengubah kepribadian atau mengasumsikan Telegram merupakan task API. Implementasi harus memakai mekanisme yang benar-benar didukung instalasi aktual; kemampuan dokumentasi produk belum membuktikan service lokal aktif atau isolation antar profil.

## Keputusan yang diusulkan

1. Registry aplikasi membuat ID UUID stabil per agen dan mapping terulas ke profil native existing. Display name, role dan Telegram pada [overview](../00-project-overview.md#enam-agen-yang-harus-dipertahankan) mengikuti Phase 2, bukan evidence konektivitas. ID aplikasi tidak berasal dari Telegram username atau identifier internal Hermes.
2. **Tidak membuat enam LLM agent/prompt baru**, tidak memodifikasi `.hermes`, persona, system prompt, credential atau routing existing. Adapter hanya menyusun task brief dan konteks minimum sebagai data pekerjaan; tidak mengganti identitas agen.
3. Invocation memakai `AgentGateway` outbound port. Supported API/private gateway/A2A/invocation native lain hanyalah kandidat sampai interface, authentication, profile routing dan result contract dibuktikan. ADR ini sengaja tidak menetapkan endpoint, CLI flags, listener atau mekanisme callback rekaan.
4. Native private interface yang didukung dan terverifikasi lebih disukai daripada Telegram bot-to-bot. Jika hanya subset capability terbukti, implementasikan subset itu; bila dispatch/result belum terbukti, Phase 3 terkait tetap terblokir atau memakai proses manual berlabel yang disetujui owner, bukan berpura-pura automated/ready.
5. Envelope request/response berversi membawa task/run ID, correlation/operation ID, expected deliverable schema, tujuan, batas tindakan, referensi sumber/versi dan context policy. `native_profile_ref` di-allowlist server-side; tidak menerima shell command/path profil bebas dari user.
6. Task boleh langsung ke spesialis mana pun. Claudia dapat mengoordinasikan workflow kampanye, tetapi bukan mandatory gateway semua tugas.
7. Run/operation ledger PostgreSQL menjaga korelasi, idempotensi, attempt dan observed outcome. Native retry/cancel/query capability tidak diasumsikan. Ambiguous dispatch outcome masuk unknown/reconcile, bukan resend buta.
8. Output raw tersimpan privat dengan retensi/redaction yang disepakati, structured output divalidasi schema. Malformed/unvalidated ditandai jelas dan tidak otomatis menjadi content approved.
9. Context builder menegakkan requester/member permission, brand grant, delegation run dan brand-agent access; query hanya versi relevan yang diizinkan, dengan provenance/classification dan approval status. Secret atau data brand lain tidak dikirim ke agen. Worker bukan superuser tanpa batas.

## Evidence yang harus ada sebelum enable per profil

| Area | Bukti yang diminta | Default bila belum ada |
|---|---|---|
| Identitas/routing | ID native/profil yang benar dan destination yang tidak tercampur | `unknown`, bukan inferensi dari nama/Telegram |
| Invocation/auth | Interface versi aktual, cara autentikasi, reachability terbatas | Dispatch disabled |
| Korelasi/hasil | Receipt, correlation task-run, retrieval/callback/polling yang dibuktikan | Tidak mengklaim running/completed |
| Status | Timestamp dan signal heartbeat/API/worker yang bermakna | Unknown/stale; configured bukan ready |
| Idempotency/retry | Replay key, duplicate behavior, query external execution | Tidak retry side effect ambigu |
| Cancellation/timeout | Bedakan request cancel, acknowledged stop dan late result | Cancel bisnis tidak berarti proses berhenti |
| Security | Scope brand, profil isolation, tool boundary, redaction dan network restriction | Tolak context/akses berisiko |

Capability record memuat verification status, evidence reference, observed/verified timestamp dan contract version. `configured`, `reachable`, `ready`, `busy`, `unavailable`, `error`, `unknown` dibedakan; stale observation tidak ditampilkan online.

## Alternatif dan trade-off

| Alternatif | Alasan ditolak/ditunda |
|---|---|
| Enam client LLM baru dengan prompt tiruan | Tidak mempertahankan profil existing dan menggandakan state/tooling/credential |
| Telegram username sebagai external ID/task API | Tidak membuktikan invocation/results atau reliable bot communication |
| Semua tugas lewat Claudia | Menghambat direct assignment dan menambah bottleneck yang tidak diminta |
| Embed Hermes internal DB/config dalam domain | Mengikat schema/vendor dan mengancam isolasi profil; adapter contract lebih aman |
| Shell invocation bebas dari HTTP | Risiko injection, timeout/proses liar dan profile access; hanya mekanisme native terulas/terbukti dengan allowlist dan batas proses bila kelak dipilih |

Keuntungan: preserve existing investment, profile identity dan testable integration. Biaya: discovery/gate nyata, kemungkinan capability berbeda antar agen, reconciliation/manual escalation dan schema validation. Telegram group privacy/status bukan health check yang cukup.

## Acceptance dan keputusan owner

Owner menyetujui pilihan interface setelah laporan discovery `docs/09-hermes-integration.md` dan matriks capability tersedia; jangan membuka listener atau mengubah profil demi membuktikan proposal tanpa izin. Phase 3 harus menunjukkan dispatch/result pada test task aman, profile isolation, correlation, malformed/duplicate/delayed output, timeout, cancellation limitation dan restart recovery. Dokumen ini tidak menjalankan agen atau live probe; capability belum diverifikasi tidak boleh dinyatakan implemented.

Lihat [task/domain](../05-domain-model.md), [task flow](../architecture/data-flow.md#2-tugas-outbox-temporal-dan-hasil-agen), [ADR workflow](ADR-0004-workflow-orchestration.md).
