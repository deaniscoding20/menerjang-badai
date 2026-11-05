# Phase 6 — Social Analytics and Marketing Intelligence

**Prerequisite:** Phase 5 acceptance criteria passed  
**Primary outcome:** Reliable marketing performance data and embedded analytics

## 1. Objective

Collect authorized social and campaign metrics, normalize them carefully, and expose meaningful dashboards through the main NEWCONNEXT application.

Apache Superset is the visualization and BI layer. It is not a social media data collector.

## 2. Data source discovery

For each platform, verify:

- Official API availability.
- Required permissions.
- Account type requirements.
- Review or app approval requirements.
- Available metrics.
- Metric definitions.
- Historical data limits.
- Rate limits.
- Data retention rules.
- Token refresh requirements.

Potential platforms include Instagram, Facebook, LinkedIn, TikTok, and YouTube, subject to verified access.

Do not promise unsupported platform coverage.

## 3. Ingestion architecture

Implement a platform-specific adapter for each supported data source.

The ingestion layer must:

- Authenticate securely.
- Fetch only authorized data.
- Respect rate limits.
- Track cursors and sync windows.
- Retry transient failures.
- Avoid duplicate records.
- Record sync timestamps and errors.
- Preserve source metric names and raw payload references where appropriate.
- Normalize metrics into a documented internal model.

Do not discard platform-specific dimensions that are important for correct interpretation.

## 4. Metric model

Potential metrics include:

- Reach.
- Impressions.
- Engagement.
- Likes.
- Comments.
- Shares.
- Saves.
- Views.
- Followers.
- Clicks.
- CTR.
- Conversions.
- Revenue.
- Spend.
- ROI.

Not every platform provides every metric.

For each metric, document:

- Definition.
- Source platform.
- Granularity.
- Unit.
- Time zone.
- Attribution window, if applicable.
- Whether it is organic or paid.
- Known limitations.

Never substitute missing metrics with fabricated values.

## 5. Analytics dimensions

Support filtering and grouping by:

- Organization.
- Brand.
- Campaign.
- Platform.
- Social account.
- Content item.
- Publication date.
- Reporting period.
- Organic/paid classification.

Distinguish post-level, account-level, and campaign-level metrics.

## 6. Superset integration

Embed Apache Superset dashboards securely.

Requirements:

- Generate guest tokens from the backend.
- Scope embedded dashboards to authorized users and brands.
- Never expose Superset admin credentials to the frontend.
- Validate dashboard access and row-level security configuration.
- Document guest-token expiration and renewal.
- Avoid trusting frontend-provided brand IDs without backend authorization.

## 7. NEWCONNEXT analytics pages

Provide:

- Executive marketing overview.
- Brand performance.
- Campaign performance.
- Platform comparison.
- Content performance.
- Audience growth.
- Engagement trends.
- Conversion and ROI views where source data supports them.
- Data freshness and sync health.

Display data freshness and incomplete-source warnings.

## 8. Competitor intelligence

Implement a source-based research record:

- Competitor profile.
- Public reference URL.
- Observation.
- Observation date.
- Source date.
- Researcher/agent.
- Evidence.
- Confidence or limitation.
- Report version.

Do not claim complete competitor monitoring unless a reliable source supports it.

## 9. Social listening

Only implement where authorized and technically supported.

Potential capabilities:

- Public brand mentions.
- Accessible comments.
- Topic classification.
- Sentiment interpretation.
- Trend summaries.

Clearly distinguish observed data from AI interpretation.

Do not bypass platform controls or perform unauthorized scraping.

## 10. Reporting

Support weekly and monthly reports with:

- Reporting period.
- Brand/campaign scope.
- Key metrics.
- Performance comparisons.
- Notable changes.
- Limitations and missing data.
- Recommendations.
- Source references.

AI-generated recommendations must be distinguishable from measured facts.

## 11. Acceptance criteria

- Metrics originate from verified data sources.
- Metric definitions are documented.
- Missing metrics are not fabricated.
- Data sync status and freshness are visible.
- Brand access is enforced in analytics.
- Superset guest tokens are generated server-side.
- Superset administrative credentials are never exposed.
- Reports identify source limitations.
- Ingestion is idempotent and recoverable.

## 12. Handoff to Phase 7

Provide operational dashboards, reliable metric ingestion, and reporting foundations for production hardening.
