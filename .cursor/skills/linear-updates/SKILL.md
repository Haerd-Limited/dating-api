---
name: linear-updates
description: Format release notes and status updates for Linear in a consistent structure. Use when the user asks for a Linear update, release note, shipped summary, or status update for Linear (or similar issue trackers).
---

# Linear updates format

When writing a Linear update, release note, or "shipped" summary, use this structure so all updates look consistent.

## Structure

1. **Title (h1):** `## [Author name] Shipped: [Short descriptive title]`
2. **Summary:** One short paragraph: what was achieved and why it matters.
3. **What shipped:** Bulleted list of user-facing or functional outcomes (endpoints, behavior, limits, audit, data scope, etc.).
4. **Technical:** Bulleted list of implementation details (new domains, migrations, new repo methods, services, flow, file paths).

## Template

```markdown
## [Author] Shipped: [Concise title]

**Summary**

[One paragraph: what was implemented and the impact.]

**What shipped**

- **[Topic]:** [Detail.]
- **[Topic]:** [Detail.]
- **Data / scope:** [What's included, limits, audit.]

**Technical**

- **New / changed:** [Domain, migration, table, or component] — [brief detail].
- **Repo/service:** [New or changed methods, flow.]
- **Apply / run:** [How to deploy or run, if relevant.]
```

## Rules

- Use bold for section headings (**Summary**, **What shipped**, **Technical**) and for the first word of each bullet when it's a label (e.g. **Endpoint:**, **Rate limiting:**).
- In **What shipped**, focus on behavior and value (what users or the system get). Include endpoint paths, status codes, and messages when relevant.
- In **Technical**, focus on code/schema (paths, migrations, new methods, which handler/service owns the flow).
- Keep bullets short; use sub-bullets only when necessary.
- If there is no "shipped" yet, use a title like "In progress: [title]" or "Update: [title]" and keep the same sections.
