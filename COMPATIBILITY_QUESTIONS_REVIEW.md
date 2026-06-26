# Compatibility Question Pack Review — Does the Science Back It Up?

A research-grounded audit of the 50 compatibility questions seeded in
`migrations/20251022211336_create_matching_schema_and_seed.sql`, measured against
what relationship science actually says predicts long-term romantic success.

> **Scope.** This reviews the **50 structured compatibility questions** across the
> **10 question packs** that feed the weighted match percentage. It does *not* cover
> profile voice prompts, GIF prompts, or filters. The current packs were inspired by
> old-school OkCupid; this document checks each item against peer-reviewed research and
> recommends keep / reframe / de-weight / replace, plus a redesigned "best-questions" set.

---

## TL;DR

1. **The good news:** Several packs are genuinely science-aligned — specifically the
   **dealbreaker / value-congruence** items (kids, relationship intent, monogamy,
   faith importance, smoking/drugs, money values). These are the parts that matter most.
2. **The bad news:** Roughly **a third of the questions are low-signal "lifestyle
   preference" noise** (tidiness, early-bird/night-owl, cook vs eat out, PDA, weekend
   style, daily texting frequency). Research shows similarity on these barely predicts
   long-term satisfaction — they inflate the % without making it mean more.
3. **A blunt truth about the % itself:** The largest study ever done on this
   (Joel et al., 2020, 11,196 couples) found that **who-you-pair-with similarity is a
   weak predictor** of relationship quality, and that relationship success is *largely
   unpredictable* from pre-relationship questionnaires. The constructs that genuinely
   predict success (conflict patterns, responsiveness, sexual satisfaction, commitment,
   emotional stability) are **mostly emergent** — they show up *inside* a relationship,
   not in two static profiles.
4. **What that means for Haerd:** An OkCupid-style "similarity %" should be honestly
   reframed as a **"Values & Dealbreaker Alignment" score** — it's excellent at
   *screening out* incompatibility and signalling value/goal congruence, but it can't
   promise chemistry. Lean into the dealbreakers, drop the noise, and add the
   high-signal constructs that are currently missing entirely.
5. **Missing entirely (high value):** emotional stability / stress reactivity
   (the single strongest individual predictor), attachment/closeness comfort,
   sexual-values alignment, financial transparency & debt attitudes, conflict-repair &
   taking responsibility, openness/adventure, and explicit views on marriage.
6. **Recommended action:** keep ~32 of the 50, reframe ~8, cut/replace ~10, and add
   ~10 new high-signal items. Re-tier the importance weighting so the score is driven
   by what science says matters.

This also directly answers the skeptical DM ("How do you know your compatibility %
translates to a relationship that works? I've seen incompatible people work and vice
versa — is there any science?"). Short version: **yes, but only for the right
questions.** See [What This Means For Haerd's Match %](#what-this-means-for-haerds-match-).

---

## 1. What Actually Predicts Long-Term Relationship Success

Seven findings from the strongest available research. Effect sizes are included so we
can weight questions honestly rather than by gut feel.

### 1.1 Relationship quality is mostly *emergent*, and matching is a weak predictor
**Joel, Eastwick et al. (2020), PNAS** ([paper](https://www.pnas.org/doi/10.1073/pnas.1917036117)
· [PubMed](https://pubmed.ncbi.nlm.nih.gov/32719123/)
· [full PDF](https://sites.lsa.umich.edu/whirl/wp-content/uploads/sites/792/2020/08/Joel-et-al-2020-PNAS.pdf)
· [plain-English summary](https://www.psychologytoday.com/us/blog/dating-and-mating/202007/the-strongest-predictors-of-romantic-relationship-quality))
— machine learning across **43 longitudinal datasets, 11,196 couples, ~2,000 variables.** The strongest predictors of relationship
quality were **relationship-specific** experiences: *perceived partner commitment,
appreciation, sexual satisfaction, perceived partner satisfaction, and conflict.*
The strongest *individual* predictors were *life satisfaction, negative affect,
depression, attachment avoidance, and attachment anxiety.*

Three implications matter enormously for a compatibility %:
- **Your own relationship-specific experience explains 2–4× more than your partner's
  traits.** "Combining two profiles" (the core of a similarity %) added essentially
  nothing beyond a person's own experience.
- Relationship-specific variables explained **up to 45% of variance at baseline but only
  ~18% later**; individual differences ~21% / ~12%.
- **Change in relationship quality over time was largely unpredictable** from any
  combination of self-reports.

> Takeaway: a questionnaire taken *before* two people interact has a hard ceiling on how
> well it can predict success. The honest job of the % is to **screen out doomed pairings
> and surface value/goal congruence** — not to promise a soulmate.

### 1.2 *Actual* similarity barely predicts satisfaction; *perceived* similarity does
**Montoya, Horton & Kirchner (2008)** meta-analysis
([paper](https://journals.sagepub.com/doi/10.1177/0265407508096700)
· [via ResearchGate](https://www.researchgate.net/publication/249719130_Is_actual_similarity_necessary_for_attraction_A_meta-analysis_of_actual_and_perceived_similarity)),
460 effects, 313 studies: actual similarity predicts attraction at
**zero/short acquaintance only**; in **existing relationships its effect is ~null.**
Perceived similarity predicts attraction throughout. **Tidwell, Eastwick & Finkel
(2013)** ([PDF](https://faculty.wcas.northwestern.edu/eli-finkel/documents/InPress_TidwellEastwickFinkel_PersonalRelationships_000.pdf))
replicated this in speed-dating: perceived, not actual, similarity drove attraction.
**Dyrenforth et al. (2010, JPSP)** ([DOI](https://doi.org/10.1037/a0020385)):
across three nationally representative samples, **personality similarity was unrelated
to satisfaction** — your *own* personality mattered far more. A 2025 scoping review of
**339 studies** ([Sage](https://journals.sagepub.com/doi/10.1177/02654075251349720))
reaches the same conclusion: similarity is *not* universally linked to better outcomes.

> Takeaway: matching people because they both "like pets," "cook at home," or are both
> "early birds" is mostly statistical noise. Similarity scoring on lifestyle preferences
> is the weakest possible use of the engine.

### 1.3 …but similarity in *values, life goals, and dealbreakers* does matter (a bit)
**Arránz Becker (2013)**, Personal Relationships
([DOI](https://doi.org/10.1111/j.1475-6811.2012.01417.x)) — German Family Panel,
3,674 couples: similarity in **life goals and values** has *small but real* positive
effects on satisfaction and stability. A **goal-interdependence meta-analysis (2022)**
([Sage](https://journals.sagepub.com/doi/abs/10.1177/02654075221128994)) found
**goal congruence correlates r = .43** with satisfaction — one of the larger congruence
effects in the literature. And the classic structural dealbreakers — **wanting children,
religiosity, monogamy** — are where mismatch reliably ends relationships.

> Takeaway: congruence on **values, goals, and a small set of structural dealbreakers**
> is the scientifically defensible core of a compatibility score. Everything else should
> be de-weighted.

### 1.4 It's not *what* you fight about — it's *how* (Gottman)
**Gottman & Levenson's** "Four Horsemen"
([Gottman Institute](https://www.gottman.com/blog/the-four-horsemen-recognizing-criticism-contempt-defensiveness-and-stonewalling/)
· [antidotes](https://www.gottman.com/blog/the-four-horsemen-the-antidotes/)) —
**criticism, contempt, defensiveness, stonewalling** — predicted divorce with
**~90% accuracy**, with **contempt the single strongest predictor.** Stable couples
maintained a **≥5:1 positive-to-negative** interaction ratio. **Repair attempts** and
"soft start-ups" distinguish masters from disasters.

> Takeaway: the Conflict & Communication pack is targeting the right *domain* but the
> wrong *variables* — it asks about logistics (how many daily check-ins) instead of the
> conflict *patterns* that actually predict divorce (contempt, defensiveness, repair).

### 1.5 Money *conflict* is the #1 divorce predictor among everyday disagreements
**Dew, Britt & Huston (2012), Family Relations**
([DOI](https://doi.org/10.1111/j.1741-3729.2012.00715.x)
· [Britt & Huston working paper PDF](https://acci.memberclicks.net/assets/docs/CIA/CIA2011/2011_britthuston.pdf)
· [NYT write-up](https://archive.nytimes.com/economix.blogs.nytimes.com/2009/12/07/money-fights-predict-divorce-rates/)),
N = 4,574 couples: **financial disagreements were the strongest disagreement type
predicting divorce** — stronger than fights about chores, in-laws, sex, or time
together — **even controlling for income, debt, and net worth.** Couples who disagreed
about money weekly were **~30%+ more likely to divorce.** Money fights are harsher,
last longer, and take longer to recover from.

> Takeaway: financial *values, transparency, and how disagreements are handled* are
> high-signal. "Do you keep a budget?" (a habit) is not what the research points to.

### 1.6 Emotional stability (low neuroticism) is the strongest individual trait
Of the Big Five, **neuroticism is the strongest and most consistent predictor** of
dissatisfaction and divorce (meta-analytic r ≈ −.24 to −.26, some reviews up to −.44).
See the **Five-Factor / relationship-satisfaction meta-analysis**
([ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S0092656609002001)),
the **neuroticism × relationship-quality meta-analytic review**
([BYU, 148 studies](https://scholarsarchive.byu.edu/cgi/viewcontent.cgi?article=11155&context=etd)),
and **Solomon & Jackson (2014)**, *Why do personality traits predict divorce?*
([PDF](https://gwern.net/doc/psychology/personality/2014-solomon.pdf)), which finds
low conscientiousness and low agreeableness also predict dissolution, beyond even SES
and IQ.

> Takeaway: a self-reported **emotional-stability / stress-reactivity** item would be one
> of the highest-signal questions in the whole bank — and it's currently absent.

### 1.7 Attachment security and perceived responsiveness drive satisfaction
**Li & Chan (2012)** meta-analysis ([DOI](https://doi.org/10.1002/ejsp.1842)) —
73 studies, 21,602 people: **both attachment anxiety and avoidance** are negatively
associated with relationship quality; **avoidance is the most corrosive to closeness
and support**, anxiety drives conflict. A later **actor–partner meta-analysis (132
studies)** ([ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S0191886919302673))
confirms the same. Secure attachment → higher satisfaction, trust, longevity (see
[Fraley's adult-attachment overview](https://labs.psychology.illinois.edu/~rcfraley/attachment.htm)).
**Perceived partner responsiveness** (feeling understood, validated, cared for) is a
core mechanism of satisfaction, **r ≈ .56**
([Canevello & Crocker, 2010, PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC2891543/)).

> Takeaway: light-touch **attachment/closeness-comfort** and **how-you-give-and-receive-
> care** items proxy genuinely predictive constructs better than half the lifestyle bank.

### 1.8 What OkCupid's own data said (since the packs were inspired by it)
Christian Rudder's OkTrends analysis
([original OkTrends post](https://gwern.net/doc/psychology/okcupid/thebestquestionsforafirstdate.html)
· [BBC News](https://www.bbc.co.uk/news/business-26613909)
· [Mic summary](https://www.mic.com/articles/85297/these-3-simple-questions-can-predict-if-an-okcupid-date-will-succeed))
of ~35,000 real couples found the **three questions that best predicted forming a real
relationship** were *"Do you like horror movies?",
"Have you ever travelled around another country alone?",* and *"Wouldn't it be fun to
chuck it all and go live on a sailboat?"* — **agreement** (either both yes or both no)
appeared in ~32% of successful couples, **3.7× chance**, out-performing OkCupid's own
top user-rated match questions. The reason: these are **proxies for openness to
experience and sensation-seeking**, not their literal content.

Two cautions from the same source: (1) OkCupid's match % was validated mainly against
**mutual messaging**, not long-term outcomes; and (2) Rudder's own conclusion —
*"two people may have exactly the same iTunes history, but if one doesn't like the
other's clothes or the way they look, there simply won't be any future."* Attraction
gates everything.

> Takeaway: an **openness/adventure** item is a cheap, validated signal — and none of the
> current 50 questions captures it.

---

## What This Means For Haerd's Match %

Putting the science together produces a clear verdict on the **model**, not just the
questions:

- **The dealbreaker gate is the scientifically strongest part of the system.** Hard
  "mandatory" mismatches on kids, monogamy, faith-if-important, smoking/hard drugs, and
  relationship intent reliably doom relationships. Screening these *out* is exactly what
  the % should do well, and it already can (`HasMandatoryMismatch`).
- **The similarity score on lifestyle preferences is the weakest part.** Rewarding two
  people for both being tidy night-owls who eat out is noise dressed up as insight, and
  it's precisely what makes a skeptic say "this percentage means nothing."
- **The biggest real predictors can't be matched from two profiles** (conflict dynamics,
  responsiveness, sexual satisfaction, commitment). The best a pre-match questionnaire
  can do is proxy **individual** risk factors (emotional stability, attachment, conflict-
  repair attitude) and **value/goal congruence.**

**Honest reframe (answers the DM directly):** rename and re-explain the number as a
**"Values & Dealbreaker Alignment"** score, e.g. *"This reflects how aligned you are on
the things that most often make or break long-term relationships — not a prediction of
chemistry."* This is both more truthful and more defensible, and it's a differentiator:
most apps oversell their algorithm; Haerd can be the one that's honest about what a
questionnaire can and can't do, while quietly making the questionnaire much better.

---

## 2. Scoring Framework: Tier the Questions by Evidence

Recommend classifying every question into a tier and weighting accordingly.

| Tier | Meaning | Evidence basis | Weighting role |
|------|---------|----------------|----------------|
| **A — Dealbreakers & core value/goal congruence** | Mismatch reliably ends relationships | §1.3, §1.5; kids/faith/monogamy/intent/finances | Eligible to be **mandatory**; high default weight |
| **B — Predictive individual & dyadic-risk constructs** | Proxies for real predictors | §1.4, §1.6, §1.7, §1.8; emotional stability, attachment, conflict-repair, ambition, openness | Medium weight; rarely mandatory |
| **C — Lifestyle preferences / low signal** | Weak long-term predictive value | §1.2; tidiness, chronotype, PDA, etc. | **Low weight or move to profile filters**, not the headline % |

The current `importance_weights` (irrelevant 0 → mandatory 30) are fine as a *user*
lever. The recommendation is to add an **editorial/base weight per question by tier**, so
a Tier-C "very important" can't outweigh a Tier-A "somewhat," and to **default the Tier-A
structural items toward higher importance** during onboarding.

---

## 3. Category-by-Category Review

Verdict key: **KEEP** (high signal) · **REFRAME** (right idea, fix wording/answers) ·
**DE-WEIGHT** (keep but low weight / profile filter) · **REPLACE** (cut, low signal).

### 3.1 Faith & Values  —  *Strong category, but 3 of 5 are redundant*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 1 | How important is faith or religion in your life? | A | **KEEP** | Religiosity congruence is a real stability factor; "importance" is the key axis. |
| 2 | How often do you participate in faith activities? | A/B | **KEEP (de-weight)** | Lived-practice fit, but partly redundant with #1/#4. |
| 3 | Would you date someone with very different beliefs? | A | **REFRAME / MERGE** | Dealbreaker-elicitation, but overlaps heavily with #4. |
| 4 | Do you prefer a partner who shares your faith practices? | A | **KEEP** | Cleanest dealbreaker-elicitation of the three. |
| 5 | Is personal integrity a non-negotiable? | — | **REPLACE** | **Near-zero variance** (everyone says "yes") → no discriminating power, and it isn't really "faith." Social-desirability trap. |

**Category fix:** #3, #4 (and partly #2) all ask "do you need a same-faith partner?" in
different words. Consolidate to **one** importance item + **one** dealbreaker item, and
reclaim two slots for:
- *"What most guides your big life decisions — faith/tradition, logic/evidence, or
  intuition/feelings?"* (values axis with real variance), and
- *"How aligned do your core values need to be with a partner's?"* (value-congruence,
  §1.3) — keeping integrity's intent but with variance.

### 3.2 Relationship Intent  —  *High-signal category; keep almost all*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 6 | What relationship horizon are you seeking? | A | **KEEP (make default-mandatory)** | Intent congruence (casual vs marriage-minded) is a primary go/no-go. |
| 7 | How quickly should it become exclusive? | B | **KEEP (de-weight)** | Pace preference; moderate signal. |
| 8 | How important is regular quality time? | B | **KEEP** | Light goal/lifestyle congruence. |
| 9 | Would you relocate in 2 years for the right person? | A/B | **KEEP** | Genuine logistics/goal congruence for serious daters. |
| 10 | How do you feel about long-distance? | B | **KEEP (de-weight)** | Logistics; fine. |

**Category fix:** add an explicit **"How do you feel about marriage?"** item — "horizon"
≠ marriage, and views on marriage itself are a classic structural alignment for the
"serious, marriage-minded" audience Haerd targets.

### 3.3 Kids & Family  —  *The strongest category. Keep all.*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 11 | Do you want children? | A | **KEEP (default-mandatory)** | The most reliable structural dealbreaker in dating. |
| 12 | If so, when would you start? | A | **KEEP** | Timeline congruence matters for serious daters. |
| 13 | Open to adoption/fostering? | B | **KEEP (de-weight)** | Lower base rate; relevant to a subset. |
| 14 | Importance of closeness with extended family? | B | **KEEP** | In-law/family-enmeshment expectations are a real conflict source. |
| 15 | Parenting roles: traditional / flexible / shared? | B | **KEEP** | Division-of-labour expectations predict conflict (§1.4/§1.5). |

**Category fix (add a slot):** *"Do you have children already, and how do you feel about
dating someone who does?"* — a genuine, common dealbreaker dimension currently missing.

### 3.4 Monogamy & Boundaries  —  *Keep the structural items; trim the preferences*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 16 | What relationship structure do you prefer? | A | **KEEP (default-mandatory)** | Monogamy vs ENM is a true go/no-go. |
| 17 | Is flirting with others acceptable? | B | **KEEP (reframe)** | Reasonable boundary-clarity item. |
| 18 | Comfortable with PDA? | C | **REPLACE** | Low long-term predictive value; easily negotiated preference. |
| 19 | Is watching adult content acceptable? | B/C | **REFRAME** | Some boundary signal but high judgment/social-desirability risk; reframe toward openness about sexuality. |
| 20 | How important is sexual exclusivity? | A | **MERGE into #16** | Redundant with relationship structure. |

**Category fix — add the missing high-signal item:** sexual *satisfaction* is a top-5
predictor (§1.1), yet the pack never asks about **sexual-values alignment.** Replace the
freed slots (#18, #20) with:
- *"How important is sexual compatibility / a satisfying sex life to you?"*, and
- *"How comfortable are you openly discussing sexual needs with a partner?"*
(libido/importance + openness are the matchable proxies for what later becomes sexual
satisfaction).

### 3.5 Lifestyle & Cleanliness  —  *Weakest category. Collapse it.*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 21 | How tidy do you prefer your home? | C | **DE-WEIGHT** | Mild cohabitation relevance; keep one lifestyle item at most. |
| 22 | How do you feel about pets? | C | **MOVE TO PROFILE FILTER** | A factual dealbreaker for some (allergies) — better as a profile field/filter than a weighted % item. |
| 23 | Early mornings or late nights? | C | **REPLACE** | Chronotype similarity has negligible long-term effect (§1.2). |
| 24 | Cook at home vs eat out? | C | **REPLACE** | Noise. |
| 25 | Importance of regular exercise/fitness? | C | **DE-WEIGHT** | Weak shared-lifestyle signal. |

**Category fix:** this is the single biggest opportunity. Keep tidiness + pets (as filter),
**reclaim 3 slots** and redeploy them to the highest-value missing constructs — e.g.
**emotional stability**, **attachment/closeness comfort**, and **openness/adventure**
(§1.6–§1.8). Consider renaming the pack to **"Lifestyle & Temperament."**

### 3.6 Substances  —  *Mostly solid (these are real dealbreakers).*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 26 | Do you drink alcohol? | A/B | **KEEP** | Drinking-pattern congruence has real effect; heavy use predicts dissolution. |
| 27 | Do you smoke / vape? | A | **KEEP** | Common, legitimate dealbreaker. |
| 28 | View on recreational cannabis? | B | **KEEP** | Reasonable. |
| 29 | View on other recreational drugs? | A | **KEEP** | Hard-drug use is a genuine red flag/dealbreaker. |
| 30 | Comfortable with partner drinking socially without you? | C | **REFRAME** | Lower signal; really a trust/autonomy/jealousy proxy — reframe toward trust & independence generally. |

**Category fix:** minor. Optionally reframe #30 into a broader **trust/autonomy** item
(independence in a relationship), which has more reach than the drinking-specific framing.

### 3.7 Money & Work  —  *Right domain; swap the habit item for values items*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 31 | How would you describe your money habits (saver/spender)? | A | **KEEP** | Spender/saver mismatch is a classic conflict driver (§1.5). |
| 32 | Do you keep a monthly budget? | C | **REPLACE** | A *habit*, weakly tied to values; replace with debt/transparency. |
| 33 | How important is career ambition? | B | **KEEP** | Ambition/goal congruence. |
| 34 | How do you feel about splitting expenses? | A | **KEEP** | Equity expectations + money = strong predictor. |
| 35 | One partner pausing career for family? | B | **KEEP** | Family-finance/role congruence (overlaps parenting roles). |

**Category fix — these would be among the highest-signal questions in the whole bank
(§1.5):**
- *"How do you feel about debt and financial risk?"* (cautious ↔ comfortable),
- *"Separate finances, fully joint, or a mix?"* (transparency/structure), and
- *"When you and a partner disagree about money, how do you usually handle it?"*
  (this one is doubly powerful — it hits the #1 *content* predictor **and** the #1
  *process* predictor at once).

### 3.8 Politics & Tolerance  —  *Two redundant items + one weak one*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 36 | How important are politics day-to-day? | B | **KEEP (de-weight)** | Salience matters more than position. |
| 37 | Could you date someone with very different politics? | A/B | **KEEP** | Political congruence is a small-but-real, increasingly common dealbreaker. |
| 38 | Comfortable discussing sensitive topics? | B | **REFRAME → move to Communication** | Misfiled; it's a communication-openness item. |
| 39 | Importance of community service / volunteering? | C | **REPLACE** | Weak long-term predictor; niche value. |
| 40 | Should partners share moral/ethical views? | B | **KEEP (reframe)** | Value-congruence, but watch low variance toward "yes." |

**Category fix:** consolidate #36/#37 framing, move #38 to Conflict & Communication,
replace #39. Reclaimed slots → values/goal-congruence items (§1.3) or moved to other
packs that need them.

### 3.9 Conflict & Communication  —  *Right domain, wrong variables (fix to match Gottman)*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 41 | Default conflict style (direct/collaborative/avoidant)? | B | **KEEP (elevate)** | Good; refine answers to capture *withdrawal/stonewalling* vs engagement. |
| 42 | When upset: space or immediate discussion? | B | **KEEP** | Captures demand–withdraw risk; one of the more science-aligned items. |
| 43 | Importance of apologizing / making amends? | B | **REFRAME (elevate)** | Repair attempts are critical (§1.4), but low variance — reframe to behaviour: *"After a fight, how easily do you take responsibility / say sorry?"* |
| 44 | How often do you communicate during the day? | C | **REPLACE** | Texting-frequency preference; weak predictor (and an attachment proxy at best). |
| 45 | How do you feel about couples therapy/counselling? | B | **KEEP (elevate)** | Openness to working on the relationship is a real growth-mindset signal. |

**Category fix — add the actual divorce predictors (§1.4, §1.6):**
- *"When a partner criticises you, how do you usually react?"* (defensiveness vs
  taking it on — anti-Four-Horsemen), and
- an **emotional-regulation** item: *"How easily are you thrown off by stress or strong
  emotions?"* (neuroticism proxy — the strongest individual predictor in §1.6).

### 3.10 Time & Ambition  —  *Two noise items; swap for openness & temperament*

| # | Question | Tier | Verdict | Why / Action |
|---|----------|------|---------|--------------|
| 46 | Structured vs spontaneous week? | C | **REPLACE** | Weak. |
| 47 | How do you spend weekends? | C | **REPLACE** | Weak; overlaps social energy. |
| 48 | Importance of personal goals/hobbies outside the relationship? | B | **KEEP** | Healthy autonomy/interdependence expectations. |
| 49 | Prioritise travel/adventure next 3 years? | B | **KEEP (reframe)** | Maps to the OkCupid openness finding (§1.8) — strengthen as an openness/adventure item. |
| 50 | Comfortable with different social energy (intro/extrovert)? | B | **KEEP (de-weight)** | Tolerance item; fine. |

**Category fix:** replace #46/#47 with a clean **openness-to-experience / sensation-
seeking** item (validated proxy, §1.8) and an **attachment/closeness-comfort** item
(§1.7) if not already added in 3.5.

---

## 4. Summary of Verdicts

| Action | Count | Questions |
|--------|------:|-----------|
| **KEEP** (high signal, as-is or elevate) | ~22 | 1, 2, 4, 6–17, 26–29, 31, 33–35, 41, 42, 45, 48, 50 |
| **REFRAME** (right idea, fix wording/answers) | ~7 | 3, 19, 30, 38, 40, 43, 49 |
| **DE-WEIGHT / move to profile filter** | ~5 | 13, 21, 22, 25, 36 |
| **REPLACE** (low signal) | ~10 | 5, 18, 20, 23, 24, 32, 39, 44, 46, 47 |

*(Counts overlap slightly where an item is both reframed and merged.)*

---

## 5. The Redesigned "Best Questions" Pack

Keeping the 10-pack structure (good for UX/progress), here is a science-weighted target
set. **★ = recommend default-mandatory eligible (Tier A dealbreaker).**

**Faith & Values**
1. ★ How important is faith/religion in your life?
2. ★ Do you need a partner who shares your faith/practices? *(merges old 3+4)*
3. What most guides your big decisions — faith/tradition, logic/evidence, or intuition? *(new, §1.3)*
4. How aligned do your core values need to be with a partner's? *(new, replaces "integrity")*

**Relationship Intent**
5. ★ What are you looking for right now? (casual ↔ marriage-minded)
6. ★ How do you feel about marriage? *(new)*
7. How quickly do you like things to become exclusive?
8. Would you relocate for the right relationship?

**Kids & Family**
9. ★ Do you want children?
10. ★ Do you have children / how do you feel about dating someone who does? *(new)*
11. If you want kids, what's your rough timeline?
12. How important is closeness with extended family?
13. Parenting roles: traditional, flexible, or fully shared?

**Monogamy & Intimacy** *(renamed)*
14. ★ What relationship structure do you want? (monogamy ↔ ENM; folds in exclusivity)
15. Where are your boundaries around flirting/outside attention?
16. How important is sexual compatibility to you? *(new, §1.1)*
17. How comfortable are you openly discussing sexual needs? *(new, §1.1)*

**Money & Finances** *(focus on values, §1.5)*
18. ★ Saver or spender — how do you relate to money?
19. How do you feel about debt and financial risk? *(new)*
20. Separate, joint, or mixed finances? *(new)*
21. When you and a partner disagree about money, how do you handle it? *(new — content × process)*
22. How do you feel about splitting expenses (50/50 vs proportional)?

**Conflict & Communication** *(re-anchored on Gottman, §1.4)*
23. What's your conflict style — engage, collaborate, or withdraw?
24. When upset, do you need space or to talk it out right away?
25. When criticised, how do you usually react? *(new — anti-defensiveness/contempt)*
26. After a fight, how easily do you take responsibility / repair? *(reframed from "apologise")*
27. How open are you to counselling / working on a relationship?

**Temperament & Emotional Health** *(new pack — highest individual signal, §1.6/§1.7)*
28. How easily are you thrown off by stress or strong emotions? *(neuroticism proxy)*
29. How comfortable are you with closeness and depending on a partner? *(attachment)*
30. How do you react when a partner needs space/independence? *(attachment)*
31. How do you prefer to give and receive care/support? *(responsiveness, §1.7)*

**Lifestyle** *(slimmed, low weight)*
32. How tidy do you like your home?
33. How important is fitness/health to you?
34. Pets — *(profile filter, not weighted)*

**Openness & Ambition** *(§1.8)*
35. How drawn are you to novelty/adventure vs routine/familiarity? *(openness proxy)*
36. Would you prioritise travel/adventure in the next few years?
37. How important are your own goals/hobbies outside the relationship?
38. How important is career ambition to you?

**Politics & Worldview**
39. ★ Could you be with someone with very different politics?
40. How central are politics/worldview to your day-to-day life?

> This is ~40 weighted items (plus pets as a filter): denser on Tiers A/B, with the
> lifestyle noise cut and four genuinely predictive constructs (emotional stability,
> attachment, sexual values, openness, financial transparency) added.

---

## 6. Constructs Currently Missing (ranked by evidence)

1. **Emotional stability / stress reactivity** — strongest individual trait predictor (§1.6).
2. **Attachment / comfort with closeness & autonomy** — §1.7.
3. **Sexual-values alignment** (importance + openness to discuss) — proxies a top-5
   emergent predictor (§1.1).
4. **Financial transparency, debt attitude, and money-conflict handling** — §1.5.
5. **Conflict repair & response to criticism** (anti-Four-Horsemen) — §1.4.
6. **Openness / adventure** — cheap, OkCupid-validated proxy (§1.8).
7. **Explicit views on marriage** and **existing children** — structural dealbreakers
   the current bank skips.

---

## 7. Implementation Notes (engineering)

- **Per-question base weight by tier.** Add an editorial weight column so Tier-C items
  can't dominate the % even when a user marks them "very important." This is the single
  most impactful change to make the number *mean* something.
- **Default Tier-A items toward higher importance** during onboarding (kids, intent,
  monogamy, faith-if-important, smoking/drugs) so the mandatory-gate (`HasMandatoryMismatch`)
  does its scientifically strongest job by default.
- **Move pure facts to profile filters** (pets, smoking-as-fact) rather than weighting
  them in the % twice.
- **Rename/relabel the score** in the API/app to *"Values & Dealbreaker Alignment"* (or
  similar) and add a one-line honest explanation — see
  [What This Means For Haerd's Match %](#what-this-means-for-haerds-match-). This is the
  cheapest way to answer the skeptic in the DM and to build trust.
- **Watch low-variance items.** Drop or rewrite questions everyone answers the same way
  (integrity, "is apologising important") — zero variance contributes zero discriminating
  power but dilutes the score.
- **Migration path:** new questions slot cleanly into the existing
  `question_categories` / `questions` / `question_answers` schema; reuse the
  `sort_order` and `reword`-style migrations as the pattern. Cuts should
  set `is_active = FALSE` rather than hard-delete, to preserve existing `user_answers`.

---

## 8. Honest Caveats

- Most of this research is **correlational**, **self-report**, and skewed toward
  **married, Western** samples; effect sizes for value/goal congruence are **small-to-
  moderate** (r ≈ .2–.43).
- The most robust finding in the field is humbling: **long-term success is largely
  unpredictable from pre-relationship data** (§1.1). The redesign maximises the *screening*
  and *congruence* signal a questionnaire can carry — it does not, and cannot, predict
  chemistry. Framing the product around that honesty is a feature, not a weakness.

---

## References

- Joel, S., Eastwick, P. W., et al. (2020). *Machine learning uncovers the most robust
  self-report predictors of relationship quality across 43 longitudinal couples studies.*
  PNAS, 117(32), 19061–19071. https://doi.org/10.1073/pnas.1917036117
- Montoya, R. M., Horton, R. S., & Kirchner, J. (2008). *Is actual similarity necessary
  for attraction? A meta-analysis of actual and perceived similarity.* Journal of Social
  and Personal Relationships, 25(6), 889–922.
- Tidwell, N. D., Eastwick, P. W., & Finkel, E. J. (2013). *Perceived, not actual,
  similarity predicts initial attraction in a live romantic context.* Personal Relationships.
- Dyrenforth, P. S., Kashy, D. A., Donnellan, M. B., & Lucas, R. E. (2010). *Predicting
  relationship and life satisfaction from personality… the relative importance of actor,
  partner, and similarity effects.* JPSP, 99, 690.
- Arránz Becker, O. (2013). *Effects of similarity of life goals, values, and personality
  on relationship satisfaction and stability.* Personal Relationships.
- *The role of goal interdependence in couples' relationship satisfaction: A meta-analysis*
  (2022). Journal of Social and Personal Relationships. (Goal congruence r = .43.)
- Gottman, J. M., & Levenson, R. W. — research on the "Four Horsemen" and the 5:1 ratio;
  The Gottman Institute, *The Four Horsemen.*
- Dew, J. (2009); Dew, J., Britt, S., & Huston, S. (2012). *Examining the Relationship
  Between Financial Issues and Divorce.* Family Relations. Britt, S. L., & Huston, S. J.
  (2012). *The role of money arguments in marriage.* Journal of Family and Economic Issues.
- Heller, D., Watson, D., & Ilies, R. (2004); Karney, B. R., & Bradbury, T. N. (1995);
  Roberts, B. W., et al. (2007) — neuroticism and Big Five predictors of marital
  (dis)satisfaction and divorce.
- Li, T., & Chan, D. K.-S. (2012). *How anxious and avoidant attachment affect romantic
  relationship quality differently: A meta-analytic review.* European Journal of Social
  Psychology, 42, 406–419.
- Reis, H. T., Clark, M. S., & Holmes, J. G. — perceived partner responsiveness as a core
  organising construct in relationship science.
- Rudder, C. (OkTrends, 2011). *The Best Questions for a First Date.* (OkCupid data on the
  three predictive questions; BBC News, 2014, *Is big data dating the key to long-lasting
  romance?*)
