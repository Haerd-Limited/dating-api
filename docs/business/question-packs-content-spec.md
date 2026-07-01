# Compatibility Question Packs — Content Spec

Authoritative content specification for the refreshed compatibility question bank. Phase 1 deliverable for [question-packs-content-refresh.md](question-packs-content-refresh.md). Evidence basis: [COMPATIBILITY_QUESTIONS_REVIEW.md](COMPATIBILITY_QUESTIONS_REVIEW.md).

**Scope:** content-only. No scoring-engine changes, no schema changes, no API field renames. Phase 2 transcribes this spec 1:1 into a re-seed migration.

> **Revision (2026-07-01):** Now **40 questions** after signal-focused edits on the original 43.
> - Pass 1 cut four low/non-matchable items: pace-to-exclusivity, care/love-languages, travel priority, tidiness (fitness kept by product direction).
> - Pass 2 cut two weak Faith items (decision-basis, core-values-alignment); "Faith & Values" renamed "Faith".
> - Pass 3 added three high-signal congruence items: relationship/household roles (Kids & Family), where you'll build your life (Relationship Intent), and long-term financial goals (Money & Finances).
> Global question IDs are renumbered contiguously (Q1–Q40). See §7 for the full rationale.

---

## 1. How to read this document

### Schema mapping (Phase 2 must follow exactly)

Source schema: [migrations/20251022211336_create_matching_schema_and_seed.sql](../migrations/20251022211336_create_matching_schema_and_seed.sql) (lines 5–30). `questions.sort_order` added NOT NULL by [migrations/20251202003413_add_sort_order_to_questions.sql](../migrations/20251202003413_add_sort_order_to_questions.sql).

| Spec field | DB column | Notes |
|------------|-----------|-------|
| Category `key` | `question_categories.key` | Stable; exactly one new key (`temperament_emotional_health`) |
| Category display name | `question_categories.name` | Renamed where noted; keys unchanged |
| Category order (1–11) | Insert order | No `sort_order` on categories; `GetQuestionCategories()` has no `ORDER BY` |
| Question text | `questions.text` | UNIQUE per category |
| Question order within pack | `questions.sort_order` | 1-based within each category |
| Question type | `questions.type` | Always `'structured'` |
| Active | `questions.is_active` | Always `true` |
| Answer label | `question_answers.label` | UNIQUE per question |
| Answer order | `question_answers.sort` | 1-based, left-to-right in spec |

**Metadata not seeded:** `tier` (A/B/C) and `dealbreaker_eligible` (★) are recorded for a future editorial-weighting project. They do not map to columns today.

### Content-only invariants

- No new `importance` levels beyond `irrelevant`, `a_little`, `somewhat`, `very`, `mandatory`.
- Every question `type` is `'structured'`.
- No API field renames.

### Answer convention (D6)

- Single-select, 3–5 ordered options per question.
- Genuine variance — no zero-variance / social-desirability traps (e.g. old "integrity" question).
- Concise labels (~≤60 characters where possible).
- No "prefer not to say" option.

### Decisions recorded

| ID | Decision |
|----|----------|
| D1 | Substances stays its own pack — **11 categories total** (overrides review §5 which dropped it). |
| D2 | Variable pack sizes — **40 questions total**, not uniform per pack. |
| D3 | Emotional-regulation lives in **Temperament & Emotional Health**, not Conflict. |
| D4 | **Pets dropped** from packs; deferred to profile attribute. |
| D5 | Category **keys stable**; display names renamed where noted; exactly one new key. |
| D6 | Answer convention (above). |
| D7 | Tier and ★ are forward-compatible metadata only — not wired into scoring now. |
| D8 | Category display order = migration insert order. |
| D9 | Order categories by **predictive strength** (highest → least); Phase 2 re-seeds all 11 category rows in that order. |
| D10 | Lifestyle reduced to a **single question** (fitness/health) after tidiness was cut. |
| D11 | **Signal trim pass 1 (2026-07-01):** cut pace-to-exclusivity, care/love-languages, travel priority, tidiness. Fitness/health kept by product direction. |
| D12 | **Signal trim pass 2 (2026-07-01):** cut the two weak Faith & Values items (decision-basis, core-values-alignment), leaving Faith as a 2-item dealbreaker pack. |
| D13 | **Renamed display name "Faith & Values" → "Faith"** (2026-07-01), since the values items were cut. Key `faith_values` unchanged (D5). |
| D14 | **Added 3 high-signal congruence items (2026-07-01):** relationship/household roles (Kids & Family), where you'll build your life (Relationship Intent), long-term financial goals (Money & Finances). Concrete, matchable positions with real variance. |

---

## 2. Categories (11 packs, predictive-strength order)

Insert categories in this exact order (D9). Faith grouped with value/structural dealbreakers per product direction.

| # | Key | Display name | Q count | Change | Rationale |
|---|-----|--------------|--------:|--------|-----------|
| 1 | `kids_family` | Kids & Family | 6 | — | Tier A; wanting/having children is the most reliable structural dealbreaker (review §3.3). |
| 2 | `relationship_intent` | Relationship Intent | 4 | — | Tier A; intent/marriage congruence is the primary go/no-go (§3.2). |
| 3 | `monogamy_boundaries` | Monogamy & Intimacy | 4 | Renamed | Tier A; structure is a true go/no-go; sexual satisfaction is a top-5 predictor (§3.4, §1.1). |
| 4 | `faith_values` | Faith | 2 | Renamed | Tier A; faith-importance and need-for-shared-faith are the two matchable dealbreakers (§3.1). Display name shortened to "Faith" (D13); key `faith_values` unchanged. |
| 5 | `money_work` | Money & Finances | 6 | Renamed | Tier A; financial conflict is the #1 content predictor of divorce (§1.5). |
| 6 | `conflict_communication` | Conflict & Communication | 5 | — | Tier B; Gottman conflict/repair process predicts dissolution (§1.4). |
| 7 | `temperament_emotional_health` | Temperament & Emotional Health | 3 | **New key** | Tier B; neuroticism + attachment (§1.6–§1.7). |
| 8 | `substances` | Substances | 4 | — | Tier A/B; smoking and hard drugs are genuine dealbreakers for a subset (§3.6). |
| 9 | `time_ambition` | Openness & Ambition | 3 | Renamed | Tier B; openness/autonomy/ambition congruence (§1.8). |
| 10 | `politics_tolerance` | Politics & Worldview | 2 | Renamed | Tier B; small-but-real dealbreaker (§3.8). |
| 11 | `lifestyle_cleanliness` | Lifestyle | 1 | Renamed | Tier C; single low-weight lifestyle item (fitness). |

**Total: 40 questions.**

Phase 3 note: `temperament_emotional_health` needs an app icon. Renamed display names are API-driven — no app copy change.

---

## 3. Questions and answers

Global IDs (Q1–Q40) are stable internal references. Within-pack `sort_order` matches the `order` column below.

★ = dealbreaker-eligible (Tier A). Tier is forward-compatible metadata (D7).

---

### Pack 1 — Kids & Family (`kids_family`)

#### Q1 · sort 1 · ★ · Tier A

**Text:** Do you want children?

| sort | label |
|-----:|-------|
| 1 | Yes, I want children |
| 2 | Maybe / unsure |
| 3 | No, I don't want children |

**Provenance:** KEEP — old Q11.

---

#### Q2 · sort 2 · ★ · Tier A

**Text:** Do you have children, and how do you feel about dating someone who does?

| sort | label |
|-----:|-------|
| 1 | I have children — open to a partner with or without |
| 2 | I have children — prefer a partner without children |
| 3 | No children — open to a partner with children |
| 4 | No children — prefer a partner without children |

**Provenance:** NEW — review §6 item 7 (existing children).

---

#### Q3 · sort 3 · Tier A

**Text:** If you want children, what's your rough timeline?

| sort | label |
|-----:|-------|
| 1 | Within 1–2 years |
| 2 | In 3–5 years |
| 3 | 5+ years |
| 4 | Not sure / no timeline |

**Provenance:** KEEP — old Q12.

---

#### Q4 · sort 4 · Tier B

**Text:** How important is closeness with extended family?

| sort | label |
|-----:|-------|
| 1 | Very important |
| 2 | Somewhat important |
| 3 | Not very important |
| 4 | Not important |

**Provenance:** KEEP — old Q14.

---

#### Q5 · sort 5 · Tier B

**Text:** How should parenting roles be divided?

| sort | label |
|-----:|-------|
| 1 | Traditional roles |
| 2 | Flexible — mix of both |
| 3 | Fully shared / equal split |

**Provenance:** KEEP — old Q15.

---

#### Q6 · sort 6 · Tier B

**Text:** In a committed relationship, how do you see roles and responsibilities at home?

| sort | label |
|-----:|-------|
| 1 | Traditional — clearly separate roles |
| 2 | Mostly traditional |
| 3 | Mostly shared / equal |
| 4 | Fully shared — we split everything equally |

**Provenance:** NEW (D14) — household division-of-labour congruence (§1.4/§1.5). Distinct from Q5 (parenting-specific); this covers the whole household, including childless couples.

---

### Pack 2 — Relationship Intent (`relationship_intent`)

#### Q7 · sort 1 · ★ · Tier A

**Text:** What are you looking for in a relationship right now?

| sort | label |
|-----:|-------|
| 1 | Marriage-minded / long-term commitment |
| 2 | Serious relationship, open to where it leads |
| 3 | Casual / short-term |
| 4 | Still figuring it out |

**Provenance:** KEEP — old Q6.

---

#### Q8 · sort 2 · ★ · Tier A

**Text:** How do you feel about marriage?

| sort | label |
|-----:|-------|
| 1 | I want to get married |
| 2 | Open to marriage |
| 3 | Unsure |
| 4 | Not interested in marriage |

**Provenance:** NEW — review §6 item 7 (explicit marriage views).

---

#### Q9 · sort 3 · Tier B

**Text:** Would you relocate for the right relationship?

| sort | label |
|-----:|-------|
| 1 | Yes |
| 2 | Maybe — depends on circumstances |
| 3 | No |

**Provenance:** KEEP — old Q9.

---

#### Q10 · sort 4 · Tier B

**Text:** Where do you see yourself building a life long-term?

| sort | label |
|-----:|-------|
| 1 | Big city |
| 2 | Suburb or small town |
| 3 | Rural / countryside |
| 4 | No strong preference |

**Provenance:** NEW (D14) — location/lifestyle-setting congruence (a concrete logistics/goal axis beyond the one-off relocate question).

---

### Pack 3 — Monogamy & Intimacy (`monogamy_boundaries`)

#### Q11 · sort 1 · ★ · Tier A

**Text:** What relationship structure do you want?

| sort | label |
|-----:|-------|
| 1 | Strictly monogamous |
| 2 | Monogamous with some flexibility |
| 3 | Open / ethically non-monogamous |
| 4 | Polyamorous / multi-partner |

**Provenance:** KEEP + MERGE — old Q16 + Q20 (exclusivity folded in).

---

#### Q12 · sort 2 · Tier B

**Text:** Where are your boundaries around flirting or outside attention?

| sort | label |
|-----:|-------|
| 1 | Not acceptable |
| 2 | Light flirting is fine |
| 3 | Depends on context |
| 4 | Generally fine |

**Provenance:** KEEP (reframe) — old Q17.

---

#### Q13 · sort 3 · Tier B

**Text:** How important is sexual compatibility to you?

| sort | label |
|-----:|-------|
| 1 | Very important |
| 2 | Somewhat important |
| 3 | Nice but not essential |
| 4 | Not a priority |

**Provenance:** NEW — review §6 item 3 (sexual values).

---

#### Q14 · sort 4 · Tier B

**Text:** How comfortable are you openly discussing sexual needs with a partner?

| sort | label |
|-----:|-------|
| 1 | Very comfortable |
| 2 | Somewhat comfortable |
| 3 | Uncomfortable but willing |
| 4 | Prefer not to discuss |

**Provenance:** NEW — review §6 item 3 (sexual values).

---

### Pack 4 — Faith (`faith_values`)

#### Q15 · sort 1 · ★ · Tier A

**Text:** How important is faith or religion in your life?

| sort | label |
|-----:|-------|
| 1 | Central to my life |
| 2 | Important but balanced |
| 3 | Cultural / occasional |
| 4 | Not important |

**Provenance:** KEEP — old Q1.

---

#### Q16 · sort 2 · ★ · Tier A

**Text:** Do you need a partner who shares your faith or practices?

| sort | label |
|-----:|-------|
| 1 | Yes — must share my faith |
| 2 | Prefer shared faith but flexible |
| 3 | Faith doesn't need to match |
| 4 | Not applicable — faith isn't important to me |

**Provenance:** KEEP + MERGE — old Q3 + Q4.

---

### Pack 5 — Money & Finances (`money_work`)

#### Q17 · sort 1 · ★ · Tier A

**Text:** Saver or spender — how do you relate to money?

| sort | label |
|-----:|-------|
| 1 | I'm a saver |
| 2 | More saver than spender |
| 3 | More spender than saver |
| 4 | I'm a spender |

**Provenance:** KEEP — old Q31.

---

#### Q18 · sort 2 · Tier B

**Text:** How do you feel about debt and financial risk?

| sort | label |
|-----:|-------|
| 1 | Avoid debt — very cautious |
| 2 | Some debt is OK if managed |
| 3 | Comfortable with moderate risk |
| 4 | Comfortable taking financial risks |

**Provenance:** NEW — review §6 item 4 (debt attitude).

---

#### Q19 · sort 3 · Tier B

**Text:** How would you prefer to handle finances in a relationship?

| sort | label |
|-----:|-------|
| 1 | Fully joint accounts |
| 2 | Mostly joint with some separate |
| 3 | Mostly separate with shared expenses |
| 4 | Fully separate |

**Provenance:** NEW — review §6 item 4 (financial transparency).

---

#### Q20 · sort 4 · Tier B

**Text:** When you and a partner disagree about money, how do you handle it?

| sort | label |
|-----:|-------|
| 1 | Talk it through calmly until resolved |
| 2 | Take space, then discuss |
| 3 | Avoid the topic when possible |
| 4 | Argue or shut down |

**Provenance:** NEW — review §6 item 4 (money-conflict handling).

---

#### Q21 · sort 5 · Tier A

**Text:** How do you feel about splitting expenses in a relationship?

| sort | label |
|-----:|-------|
| 1 | 50/50 split |
| 2 | Proportional to income |
| 3 | One partner covers more |
| 4 | Flexible — case by case |

**Provenance:** KEEP — old Q34.

---

#### Q22 · sort 6 · Tier B

**Text:** How ambitious are your long-term financial goals (e.g. owning a home, building wealth)?

| sort | label |
|-----:|-------|
| 1 | Very — ownership / building wealth is a priority |
| 2 | Moderately — steady progress matters |
| 3 | Not very — comfort matters more than growth |
| 4 | I don't set long-term financial goals |

**Provenance:** NEW (D14) — financial-goal congruence (§1.5). Distinct from Q17 (day-to-day money personality); this is long-term aspiration.

---

### Pack 6 — Conflict & Communication (`conflict_communication`)

#### Q23 · sort 1 · Tier B

**Text:** What's your default conflict style?

| sort | label |
|-----:|-------|
| 1 | Address it directly |
| 2 | Collaborate and find compromise |
| 3 | Need time before discussing |
| 4 | Avoid or withdraw |

**Provenance:** KEEP (elevate) — old Q41.

---

#### Q24 · sort 2 · Tier B

**Text:** When you're upset, do you need space or to talk it out right away?

| sort | label |
|-----:|-------|
| 1 | Need space first |
| 2 | Mix — depends on the situation |
| 3 | Want to talk it out right away |
| 4 | Prefer not to discuss it |

**Provenance:** KEEP — old Q42.

---

#### Q25 · sort 3 · Tier B

**Text:** When criticised by a partner, how do you usually react?

| sort | label |
|-----:|-------|
| 1 | Listen and reflect on the feedback |
| 2 | Get defensive but try to hear them |
| 3 | Push back or argue |
| 4 | Shut down or withdraw |

**Provenance:** NEW — review §6 item 5 (anti-defensiveness).

---

#### Q26 · sort 4 · Tier B

**Text:** After a fight, how easily do you take responsibility and repair?

| sort | label |
|-----:|-------|
| 1 | I apologise and repair quickly |
| 2 | I need time but will repair eventually |
| 3 | I struggle to apologise first |
| 4 | I rarely initiate repair |

**Provenance:** REFRAME — old Q43 (behavioural repair, not "is apologising important").

---

#### Q27 · sort 5 · Tier B

**Text:** How open are you to couples therapy or working on a relationship?

| sort | label |
|-----:|-------|
| 1 | Very open |
| 2 | Open if needed |
| 3 | Unlikely but not closed off |
| 4 | Not interested |

**Provenance:** KEEP (elevate) — old Q45.

---

### Pack 7 — Temperament & Emotional Health (`temperament_emotional_health`)

#### Q28 · sort 1 · Tier B

**Text:** How easily are you thrown off by stress or strong emotions?

| sort | label |
|-----:|-------|
| 1 | I stay calm under pressure |
| 2 | I recover fairly quickly |
| 3 | It affects me for a while |
| 4 | I get overwhelmed easily |

**Provenance:** NEW — review §6 item 1 (neuroticism proxy).

---

#### Q29 · sort 2 · Tier B

**Text:** How comfortable are you with closeness and depending on a partner?

| sort | label |
|-----:|-------|
| 1 | Very comfortable with closeness |
| 2 | Comfortable with some independence |
| 3 | I value a lot of independence |
| 4 | Closeness makes me uneasy |

**Provenance:** NEW — review §6 item 2 (attachment).

---

#### Q30 · sort 3 · Tier B

**Text:** How do you react when a partner needs space or independence?

| sort | label |
|-----:|-------|
| 1 | I give space easily |
| 2 | A little uneasy but I manage |
| 3 | I find it hard — it makes me anxious |
| 4 | I struggle and need reassurance |

**Provenance:** NEW — review §6 item 2 (attachment).

---

### Pack 8 — Substances (`substances`)

#### Q31 · sort 1 · Tier B

**Text:** Do you drink alcohol?

| sort | label |
|-----:|-------|
| 1 | Never |
| 2 | Occasionally / socially |
| 3 | Regularly |
| 4 | Frequently |

**Provenance:** KEEP — old Q26.

---

#### Q32 · sort 2 · ★ · Tier A

**Text:** Do you smoke or vape?

| sort | label |
|-----:|-------|
| 1 | Yes |
| 2 | Occasionally |
| 3 | Used to — not anymore |
| 4 | No |

**Provenance:** KEEP — old Q27.

---

#### Q33 · sort 3 · Tier B

**Text:** What is your view on recreational cannabis?

| sort | label |
|-----:|-------|
| 1 | I use it / fine with it |
| 2 | OK in moderation |
| 3 | Prefer partner doesn't use |
| 4 | Not OK with it |

**Provenance:** KEEP — old Q28.

---

#### Q34 · sort 4 · ★ · Tier A

**Text:** What is your view on other recreational drugs?

| sort | label |
|-----:|-------|
| 1 | OK with occasional use |
| 2 | Depends on the substance |
| 3 | Not OK with use |
| 4 | Hard no — dealbreaker |

**Provenance:** KEEP — old Q29.

---

### Pack 9 — Openness & Ambition (`time_ambition`)

#### Q35 · sort 1 · Tier B

**Text:** How drawn are you to novelty and adventure vs routine and familiarity?

| sort | label |
|-----:|-------|
| 1 | Strongly prefer novelty and adventure |
| 2 | Lean toward adventure |
| 3 | Lean toward routine and familiarity |
| 4 | Strongly prefer routine and familiarity |

**Provenance:** NEW — review §6 item 6 (openness).

---

#### Q36 · sort 2 · Tier B

**Text:** How important are your own goals and hobbies outside the relationship?

| sort | label |
|-----:|-------|
| 1 | Very important |
| 2 | Somewhat important |
| 3 | Nice to have |
| 4 | Not very important |

**Provenance:** KEEP — old Q48.

---

#### Q37 · sort 3 · Tier B

**Text:** How important is career ambition to you?

| sort | label |
|-----:|-------|
| 1 | Very important |
| 2 | Somewhat important |
| 3 | Moderate |
| 4 | Not a priority |

**Provenance:** KEEP — old Q33.

---

### Pack 10 — Politics & Worldview (`politics_tolerance`)

#### Q38 · sort 1 · ★ · Tier A

**Text:** Could you be with someone with very different political views?

| sort | label |
|-----:|-------|
| 1 | Yes |
| 2 | Maybe — depends on the issues |
| 3 | Probably not |
| 4 | No |

**Provenance:** KEEP — old Q37.

---

#### Q39 · sort 2 · Tier B

**Text:** How central are politics and worldview to your day-to-day life?

| sort | label |
|-----:|-------|
| 1 | Very central |
| 2 | Somewhat central |
| 3 | Rarely think about it |
| 4 | Not important to me |

**Provenance:** KEEP (de-weight) — old Q36.

---

### Pack 11 — Lifestyle (`lifestyle_cleanliness`)

#### Q40 · sort 1 · Tier C

**Text:** How important is fitness and health in your lifestyle?

| sort | label |
|-----:|-------|
| 1 | Very important — daily focus |
| 2 | Somewhat important |
| 3 | Casual — when I can |
| 4 | Not a focus |

**Provenance:** DE-WEIGHT — old Q25. Kept by product direction despite weak long-term signal (§7).

---

## 4. Missing-construct coverage (review §6)

| # | Construct | Covered by |
|---|-----------|------------|
| 1 | Emotional stability / stress reactivity | Q28 |
| 2 | Attachment / closeness & autonomy | Q29, Q30 |
| 3 | Sexual-values alignment | Q13, Q14 |
| 4 | Financial transparency, debt, money-conflict | Q17, Q18, Q19, Q20, Q21, Q22 |
| 5 | Conflict repair & response to criticism | Q25, Q26 (+ Q23 conflict style) |
| 6 | Openness / adventure | Q35 |
| 7 | Explicit marriage views & existing children | Q2, Q8 (+ Q7 intent) |

All seven constructs from review §6 remain covered.

**Additional congruence axes added in D14 (beyond review §6):** household division of labour (Q6), long-term location/setting (Q10), long-term financial goals (Q22) — concrete, matchable positions chosen for strong compatibility signal (see §7).

---

## 5. Variance check

One-line assertion per question that the answer set has genuine spread (not a social-desirability trap).

| Q | Variance OK | Notes |
|---|:-----------:|-------|
| Q1 | ✓ | Yes / maybe / no on children — core structural split |
| Q2 | ✓ | Four distinct partner-with-kids preferences |
| Q3 | ✓ | Timeline spread; "not sure" avoids forcing false precision |
| Q4 | ✓ | Importance gradient, not everyone picks "very" |
| Q5 | ✓ | Three distinct role models |
| Q6 | ✓ | Traditional ↔ fully shared — genuinely divergent in the dating pool |
| Q7 | ✓ | Casual ↔ marriage-minded spread |
| Q8 | ✓ | Want / open / unsure / not interested — not everyone wants marriage |
| Q9 | ✓ | Yes / maybe / no on relocating |
| Q10 | ✓ | City ↔ rural ↔ no preference |
| Q11 | ✓ | Monogamy ↔ ENM spectrum |
| Q12 | ✓ | Strict ↔ permissive boundaries |
| Q13 | ✓ | Importance gradient |
| Q14 | ✓ | Comfort gradient — not everyone "very comfortable" |
| Q15 | ✓ | Central ↔ not important |
| Q16 | ✓ | Must-share ↔ not applicable |
| Q17 | ✓ | Saver ↔ spender continuum |
| Q18 | ✓ | Risk tolerance gradient |
| Q19 | ✓ | Joint ↔ separate finance models |
| Q20 | ✓ | Healthy ↔ avoidant conflict styles (not all "talk calmly") |
| Q21 | ✓ | Split models vary |
| Q22 | ✓ | Ambitious ↔ no-goals spread; distinct from saver/spender |
| Q23 | ✓ | Engage ↔ withdraw spectrum |
| Q24 | ✓ | Space ↔ talk-now spectrum |
| Q25 | ✓ | Reflect ↔ shut-down — behavioural, not "I'm always great" |
| Q26 | ✓ | Repair ease gradient — replaces "apologising is important" trap |
| Q27 | ✓ | Openness to therapy varies |
| Q28 | ✓ | Calm ↔ overwhelmed — neuroticism spread |
| Q29 | ✓ | Closeness ↔ uneasy spectrum |
| Q30 | ✓ | Secure ↔ anxious reactions to partner space |
| Q31 | ✓ | Never ↔ frequently |
| Q32 | ✓ | Yes / occasional / former / no |
| Q33 | ✓ | Use ↔ not OK spectrum |
| Q34 | ✓ | OK ↔ hard-no spectrum |
| Q35 | ✓ | Adventure ↔ routine spectrum |
| Q36 | ✓ | Very ↔ not very important |
| Q37 | ✓ | Ambition gradient |
| Q38 | ✓ | Yes ↔ no on cross-party dating |
| Q39 | ✓ | Central ↔ not important |
| Q40 | ✓ | Daily focus ↔ not a focus |

---

## 6. Self-consistency validation

| Check | Result |
|-------|--------|
| Category count | 11 ✓ |
| Question count | 40 ✓ (6+4+4+2+6+5+3+4+3+2+1) |
| Per-pack counts match §2 | ✓ |
| Every ★ is Tier A | ✓ (★: Q1, Q2, Q7, Q8, Q11, Q15, Q16, Q32, Q34, Q38 — all Tier A; Q3/Q21 are Tier A but not ★) |
| Exactly one new key | `temperament_emotional_health` ✓ |
| No keys removed | ✓ (all 11 original keys retained) |
| All question text unique within category | ✓ |
| All answer labels unique within question | ✓ |
| Every question has 3–5 answers | ✓ (Q1, Q5, Q9 have 3; all others 4) |
| Content-only invariants | ✓ |
| §6 constructs covered | ✓ |
| Added in D14 | household roles (Q6), location (Q10), long-term financial goals (Q22) |
| Cut across trims | pace-to-exclusivity, care/love-languages, travel priority, tidiness (D11); decision-basis, core-values-alignment (D12) |

---

## 7. Compatibility-signal rationale (why each item earns its place)

The scoring engine ([internal/compatibility/storage/repository.go](../internal/compatibility/storage/repository.go)) compares a viewer's answer against a target's answer via `acceptable_answer_ids`, weighted by importance. An item only produces real signal if it is (a) a **congruence** position where a clash matters, (b) a **screening** filter for a trait that predicts success individually, or (c) a genuine dealbreaker. Self-rated meta-preferences and emergent constructs match weakly.

**Added for strong matchable signal (D14):**
- **Household roles (Q6)** — concrete traditional ↔ fully-shared position (plain-language labels); division-of-labour mismatch is a documented conflict driver and views are genuinely divergent. Backfills the concrete "values" item removed in D12. Distinct from parenting-specific Q5.
- **Where you'll build your life (Q10)** — city/rural/setting congruence; a real logistics/goal clash beyond the one-off relocate question (Q9).
- **Long-term financial goals (Q22)** — aspiration-level money congruence, distinct from day-to-day saver/spender (Q17).

**Cut across the trims (little/no matchable signal):**
- **Pace to exclusivity** — low-stakes, easily negotiated.
- **Give/receive care / love languages** — weak evidence; complementarity, not similarity.
- **Travel/adventure priority** — overlapped the openness item (Q35); lifestyle noise.
- **Tidiness** — weakest predictive tier (review §3.5).
- **Decision-basis** and **core-values-alignment** — abstract / self-rated meta-preference; don't match well.

**Kept despite weak signal, with rationale:**
- **Fitness/health (Q40)** — retained by product direction; low-weight lifestyle item, not a dealbreaker.
- **Temperament/attachment (Q28–Q30) and conflict items (Q23–Q27)** — weak as *similarity* signal but target the strongest *individual* predictors (§1.6–§1.7, §1.4). They function as partner **screens**, not congruence.
- **Politics/substances** — not similarity-based, but clean, high-variance **dealbreakers**.

**Structural ceiling (honest caveat):** the strongest predictors overall — sexual satisfaction, actual conflict patterns, responsiveness — are *emergent* and cannot be captured pre-match (review §1.1). This bank maximises screening + congruence signal; it cannot predict chemistry.

**Open items (flagged, not yet actioned):**
- **Q25/Q26 (criticism / repair)** and **Q29/Q30 (attachment)** — possible redundancy; could be merged for a leaner bank.
- **Q14 (discuss sexual needs)**, **Q27 (counselling openness)** — borderline keep-or-cut.

---

## 8. Sign-off

**Status:** Ready for stakeholder review before Phase 2 (re-seed migration).

Review this spec for question wording, answer options, and pack order. Once approved, Phase 2 transcribes it verbatim into SQL.
