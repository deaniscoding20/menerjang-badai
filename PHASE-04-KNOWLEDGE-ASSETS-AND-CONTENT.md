# Phase 4 — Brand Knowledge Base, Asset Management, Campaigns, and Content

**Prerequisite:** Phase 3 acceptance criteria passed  
**Primary outcome:** A centralized, brand-aware marketing production workspace

## 1. Objective

Implement brand knowledge, secure media storage, campaign management, and content production workflows. Ensure AI agents receive only relevant and authorized brand context.

## 2. Brand Knowledge Base

Support documents and structured knowledge for:

- Brand guidelines.
- Business profile.
- Products and services.
- Target audiences.
- Positioning and value proposition.
- Tone of voice.
- FAQs.
- Competitor references.
- Approved campaigns.
- Approved marketing examples.
- Research reports.

For every knowledge item, track:

- Brand.
- Type/category.
- Owner.
- Source.
- Version.
- Approval status.
- Access permissions.
- Created and updated timestamps.

Support document upload, preview where safe, search, filtering, version history, and archive.

## 3. Agent context retrieval

Implement a controlled context service.

It must:

- Resolve the task's brand.
- Retrieve only relevant knowledge.
- Enforce brand and user authorization.
- Include source references and document versions.
- Distinguish approved material from drafts.
- Avoid sending unrelated or confidential brand data.

Do not implement RAG by default if ordinary indexed search is sufficient.

If vector retrieval is justified, document the decision, evaluation method, embedding model, data lifecycle, and access-control filtering.

## 4. MinIO asset management

Use MinIO for original files and media.

Support:

- Image upload.
- Video upload.
- Document upload.
- Metadata.
- Asset preview.
- Brand and campaign association.
- Asset versioning.
- Access permissions.
- Archive.

Store object keys and metadata in PostgreSQL.

Generate short-lived presigned URLs when needed. Do not store permanent presigned URLs.

Validate file size, content type, file extension, and actual file signature where possible.

Consider malware scanning before making uploads available to other users or external publishing services.

## 5. Campaign management

Implement:

- Campaign name and description.
- Brand.
- Objectives.
- Audience.
- Channels.
- Budget fields and currency.
- Start/end dates.
- Strategy references.
- Creative direction.
- KPIs.
- Owner.
- Status.
- Linked tasks and content.

Do not treat planned budget as actual advertising spend.

## 6. Content management

Support:

- Content drafts.
- Content types.
- Captions.
- Hooks.
- Hashtags.
- Calls to action.
- Platform-specific copy.
- Media attachments.
- Campaign association.
- Target channel.
- Content status.
- Version history.

Content types may include single image, carousel, short video, story, reel, text post, and blog post, subject to platform support.

## 7. Content versioning

Every material edit creates a new version.

Track:

- Version number.
- Content body.
- Metadata.
- Asset references and versions.
- Author.
- Change summary.
- Timestamp.

Do not overwrite approved content in place.

A later approval phase must bind approval to an exact version and asset set.

## 8. Content calendar

Implement calendar and list views with:

- Date range.
- Brand filter.
- Campaign filter.
- Channel filter.
- Status filter.
- Content type filter.

At this phase, calendar entries may be planned drafts. Do not imply they are scheduled externally until Postiz integration is complete.

## 9. Acceptance criteria

- Authorized users can create and manage brand knowledge.
- Documents are stored in MinIO and metadata in PostgreSQL.
- Access controls prevent cross-brand leakage.
- Agents can receive task-specific authorized context.
- Assets are validated and securely served.
- Campaigns can be created and linked to tasks/content.
- Content versions are preserved.
- Calendar clearly distinguishes planned from externally scheduled content.
- Documentation covers object storage and knowledge lifecycle.

## 10. Handoff to Phase 5

Provide versioned content, assets, campaigns, and brand knowledge for approval and publishing.
