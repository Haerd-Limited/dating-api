# Create Project Plan

> **Recommended setup:** Planning rewards careful reasoning. Prefer a high-reasoning model for this command — shallow models tend to produce shallow phase decomposition, miss ambiguities, hallucinate constraints, and reason poorly about the steel thread. Those problems compound during implementation and are expensive to fix later. A bit more effort on the planning step pays for itself many times over.

Given a task reference (Linear URL, research,markdown file, issue description, or high-level goal), produce a high-level project plan that breaks the work into named phases. Use plan mode — no edits until the user confirms.

This command produces **project plans**, not implementation plans. A project plan captures context, constraints, design decisions, and a phased breakdown. It contains no code-level details (file paths, function signatures, code snippets). Those belong in implementation plans created via `/create-implementation-plan` for each named phase.

The plan is drafted at `.cursor/plans/<slug>.md` and committed to git as a durable project artifact. Supporting research is drafted at `.cursor/plans/<slug>-research.md` and committed alongside it. (`.cursor/plans/` is where `/critique-plan` and `/deviation-analysis` look for plans, so everything stays in one place.)

## Step 0: Research Phase

Before gathering context for the plan, complete a structured research phase using an annotation/iteration protocol. This is an **iterative process** where the agent writes research findings, the human annotates with corrections and context, and the agent revises — repeating 1-6 times until the research is solid.

**Why this works**: The written artifact creates a verification surface that forces both the agent and human to demonstrate understanding before planning begins, catching misunderstandings early when they're cheap to fix rather than during implementation when they're expensive.

**Purpose**: Build deep understanding of the current system before planning. The research phase produces a written artifact that serves as a verification checkpoint — you can review it to ensure the agent actually understands the system before making architectural decisions.

**Process**:

1. **Initial Research**: Use parallel explore subagents to deeply research the codebase. Write findings to `.cursor/plans/<slug>-research.md` with these sections:

   **Current Implementation**
   - How the relevant systems work today
   - Architecture, data flow, key components across the handler → service → repository layers
   - Affected domains (`internal/{domain}/`) and API layers (`internal/api/{domain}/`)
   - Dependencies and integration points (router registration in `internal/http/router/router.go`, wiring in `cmd/main.go`, migrations in `migrations/`, generated entities in `internal/entity/`)

   **Existing Patterns**
   - Similar implementations elsewhere in the codebase (e.g. how `interaction`, `profile`, `safety`, or `conversation` does it)
   - Established patterns that should be followed per `AGENTS.md` and `internal/README.MD`
   - Common conventions (naming, layering, mappers, error handling, testing)
   - Reference implementations to use as templates

   **API Surface Analysis**
   - What APIs/interfaces/DTOs will change
   - What consumes these APIs (the Haerd app, the admin dashboard, other services)
   - Breaking vs non-breaking change analysis

   **Testing & Validation Patterns**
   - How similar features are currently tested (service unit tests, repository tests, handler tests)
   - Test infrastructure available, mock generation (`make mock`), entity generation (`make entity`)
   - Coverage gaps

   **Known Constraints**
   - Technical constraints (performance, compatibility, dependencies, schema/migration ordering)
   - Team constraints (coordination needed, cross-repo impact, domain expertise)
   - Product constraints (commitments, timelines, backwards compatibility)
   - GDPR / PII constraints — what data is touched, logging hygiene (no `zap.Any` on structs that may carry PII)

   **Risks and Unknowns**
   - Areas where the codebase is unclear or inconsistent
   - Missing documentation or tribal knowledge gaps
   - Potential gotchas or edge cases
   - Questions that need human input to resolve

2. **Annotation Cycle**: Tell the user:

   > **Research complete**: I've written my findings to `.cursor/plans/<slug>-research.md`.
   >
   > Please review it and add inline annotations directly in the file:
   >
   > - Correct any misunderstandings about how the system works
   > - Add product context (who needs this, why, urgency)
   > - Add domain knowledge (historical context, architectural vision)
   > - Flag non-obvious constraints
   > - Note what's missing or wrong
   >
   > When you're done, say **"update research"** and I'll read your annotations and revise the document.

   **Wait for the user** to annotate and signal they're ready. Do not proceed to Step 1 until they do.

3. **Iteration**: When the user says "update research":
   - Re-read the annotated `.cursor/plans/<slug>-research.md`
   - Update the file based on their corrections and additions
   - Show them what you changed
   - Ask if they want another round or if research is complete

   Repeat the annotation cycle **1-6 times** (typical: 2-3 rounds) until the user approves.

4. **Keep in repo**: When research is approved, the final research document stays at `.cursor/plans/<slug>-research.md` and is committed to git alongside the plan in Step 5. This preserves the research as a durable artifact in the repo — the same place the plan and code live — rather than scattered across ticket comments.

5. **Proceed to Planning**: Continue to Step 1 (Gather Context) with the research findings as your foundation.

**Rules**:

- **Do not skip this step** unless the task is trivial (in which case, use `/create-implementation-plan` instead)
- **Do not proceed to Step 1** until the user has reviewed and approved the research
- Use the research findings to inform all subsequent steps — do not re-litigate what was discovered in research

## Step 1: Gather Context

**If Step 0 (Research Phase) was completed:**

- Read `.cursor/plans/<slug>-research.md` as your foundation
- Use the research findings to inform context gathering
- Focus additional exploration on areas flagged as unclear or risky in the research
- Do not re-litigate findings from the research phase unless new information conflicts

**Standard context gathering** (whether research was done or not):

Use parallel explore subagents to build a complete picture of the problem space:

**In the Workspace:**

- Locate and read the task description and related documentation
- Read `AGENTS.md` and `internal/README.MD` for the conventions you must follow
- Find prior art in the codebase — existing implementations of similar scope, patterns already established (`interaction`, `profile`, `safety`, `conversation`, etc.)
- Identify stakeholders and consumers — who is this for, who will give feedback, which clients (the Haerd app, the admin dashboard) consume the affected APIs
- Identify the deployment and integration surface — migrations, generated entities, router and `cmd/main.go` wiring, cross-repo contracts

**In Linear (via the Linear MCP server):**

- Fetch the task ticket to get full context
- Search for related Linear tickets, projects, and project updates that touch similar areas
- Check if adjacent work is in progress that might conflict or complement this project

**From the Team:**

- Ask product or engineering team members if any work has happened in this area
- Surface tribal knowledge, past experiments, or lessons learned that aren't documented
- Identify who has domain expertise in the relevant subsystems

Spend time here — a thorough read upfront prevents bad decisions later.

## Step 2: Identify and Resolve Ambiguities

Before planning, surface every decision that would materially change the plan. Ask questions using the `AskQuestion` tool:

- **Design alternatives**: If there are multiple valid approaches, present the tradeoffs concisely and ask the user to pick
- **Scope boundaries**: What's in vs. out of this project
- **Dependencies**: External dependencies, team coordination, deployment ordering, migration ordering
- **Constraints**: Timeline, compatibility, security, performance, GDPR/PII, product commitments
- **Integration points**: What systems does this touch, what APIs/DTOs change, which clients need to coordinate

**Questions up front:** All clarifying questions must be asked and answered before the first draft of the plan. Do not create the first plan until every material ambiguity is resolved. No questions may appear inside or after the first plan.

**Decisions and evidence:** When the user picks an option in response to a question, that choice must be recorded in the plan as an explicit decision (see Step 3). Where it would help, ask for evidence or rationale for the decision so the plan can note it.

Rules for questions:

- Ask 1-4 targeted questions at a time, not a wall of questions
- If answers raise new ambiguities, ask follow-up rounds until resolved
- **Architectural and scoping** decisions must be settled before the plan is created
- Code-level concerns (naming, exact signatures, error messages) belong in implementation plans, not here

## Step 3: Create the Plan

Write the plan to `.cursor/plans/<slug>.md` using the `Write` tool. The slug should be descriptive and kebab-cased (e.g., `conversation-reveal-flow`, `discover-quota-overhaul`).

Required sections:

### Goal

What success looks like and who it's for.

### Background and Context

Current state, why the project exists, product context, prior art.

### Constraints

Hard constraints: timeline, compatibility, security, performance, GDPR/PII, etc.

### Design Decisions and Tradeoffs

Numbered decisions. For each: the chosen option, alternatives considered (as a table with pros/cons/verdict), and rationale.

**Decision-table row order:** the **accepted option sits on the first row**, with the verdict cell bolded (e.g. `**Accept**`). Rejected options follow below with their verdicts. This makes the plan scannable — a reader skimming down the decisions sees the chosen path immediately rather than having to find the bolded row in each table. Highlight the accepted row with bold on the approach cell too for extra visibility.

Nothing in the plan is "optional" — every path is decided and required.

### Architecture Overview

High-level diagram (mermaid or text) showing how the pieces fit together. No code — this is about system-level relationships across the handler, service, and repository layers and the systems that consume them.

### Phases

Named phases with descriptive names (e.g. "Backend — Reveal Eligibility Service", "API — Expose Reveal Status Endpoint"). The name is the primary identifier used everywhere: in implementation plans, Linear tickets, conversations, and PR references. Numbers may appear for ordering but are secondary.

**Steel thread**: The early phases must form a steel thread — the thinnest possible end-to-end slice that proves the architecture works. Mark which phases constitute the steel thread and justify why they are the thinnest viable end-to-end slice. Later phases widen the implementation, add edge cases, polish UX, and harden for production.

Each phase declares:

| Field                   | Description                                                                                                                                                                                                                                       |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**                | Descriptive name (primary identifier)                                                                                                                                                                                                            |
| **Description**         | 1-3 sentences on what this phase achieves                                                                                                                                                                                                        |
| **Dependencies**        | Other phases (by name) that must complete first                                                                                                                                                                                                  |
| **Key decisions**       | Decisions already made that affect this phase                                                                                                                                                                                                    |
| **Validation criteria** | How do we know this phase works? Concrete risk/evidence checks, not vague statements (see guidance below)                                                                                                                                        |
| **Doc deliverables**    | What docs are written or updated as part of this phase (e.g. `internal/README.MD`, endpoint references consumed by the Haerd app / admin dashboard)                                                                                               |
| **Name quality**        | Descriptive, self-contained, makes sense in a one-line status update to someone outside the project (e.g. "Persist reveal decisions", not "Phase 4b"). If the phase can't be named descriptively, it isn't well-defined — rework scope until it can. |

**Validation criteria quality:** Project-plan validation criteria should identify the risks that matter and the evidence expected at phase boundaries. Do not turn this into a long CI checklist, and do not require tests just to prove static plan details were copied correctly. Match the check to the risk:

- **Service-layer business logic and error paths** → unit tests.
- **Non-trivial SQL** → repository tests.
- **Status-code mapping and request validation** → handler tests.
- **Anything that compiles and ships** → `make lint` and `make build` pass; `make mock` after interface changes, `make entity` after migrations.
- **Behaviour hard to cover automatically** → a named manual verification step (e.g. hitting the endpoint locally with a sample payload).

Also keep the repo's engineering bar in view (from `AGENTS.md`): respect the layering (handlers → services → repositories), update `mapErrorsToStatusCodeAndUserFriendlyMessages` when adding error types, and never `zap.Any` on structs that may carry PII. The implementation plan will refine these criteria into a code-level test plan distributed across each Part.

**Phase sizing:** Each phase must fit in a single context window so that `/create-implementation-plan` can execute it in one session. If a phase feels too large for that, split it. Do **not** include effort estimates — they are unreliable at this level and create false expectations.

**No code in project plans.** Project plans contain _what_ and _why_, never _how_ at the code level. Code snippets, file paths, function signatures, and struct definitions belong in implementation plans. When you've done deep codebase exploration to inform the plan, distill the findings into architectural descriptions, not code.

<good-example>
### Phase 2 — API: Reveal Eligibility Endpoint

**Description:** Add a read-only endpoint that reports whether a matched pair is
eligible to reveal identities, based on the message-count and mutual-consent
rules established in Phase 1. Reuses the existing conversation service and
repository.

**Dependencies:** Phase 1 (Reveal Eligibility Service)

**Key decisions:** Decision 1 (eligibility owned by the service layer), Decision 3 (read-only, no state change)

**Validation criteria:**

- An eligible pair returns `eligible: true` with the satisfied criteria
- An ineligible pair returns `eligible: false` with the unmet criteria, not an error
- Requests from a user not in the conversation are rejected with 403
- Existing conversation read paths are unchanged
- `make lint` and `make build` pass

**Doc deliverables:**

- Update `internal/README.MD` to list the new endpoint under the conversation domain
- Note the new response shape for the Haerd app team consuming the reveal flow

</good-example>

<bad-example>
### Phase 2 — API: Reveal Eligibility Endpoint

**Implementation details:**

In `internal/api/conversation/handler.go`:

```go
func (h *handler) RevealEligibility() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, _ := commoncontext.UserIDFromContext(r.Context())
        result, err := h.service.RevealEligibility(r.Context(), userID, convID)
        // ...
    }
}
```

Add an `ErrNotParticipant` error variable to `internal/conversation/service.go`.

This is wrong: exact file paths, function signatures, and error variables belong
in the implementation plan produced by `/create-implementation-plan`, not in the
project plan. Distill it into the architectural description and validation
criteria instead.

</bad-example>

### Risks and Open Questions

Unresolved items that don't block planning but need attention during implementation. Include mitigation strategies where possible.

### References

Links to docs, product context, prior art, Linear tickets, and relevant sections of `AGENTS.md` / `internal/README.MD`.

**No optional pathways:** The plan must not contain optional branches, "optional" steps, or alternative paths. Everything in it is decided and required.

## Step 4: Iterate on the Plan

The plan is a living document until the user approves it.

1. The user annotates the plan inline with corrections, rejections, or extra constraints.
2. When they signal they've added notes, re-read the plan, address every annotation, and update it with the `Write` tool.
3. Show what you changed.

### Completeness check before approval

Before asking whether another round is needed, check the plan against these **project plan criteria** and report which pass and which don't:

- Every material ambiguity from Step 2 is resolved and recorded as an explicit decision (no questions remain inside or after the plan).
- The required structure is present: Goal, Background and Context, Constraints, Design Decisions and Tradeoffs, Architecture Overview, Phases, Risks and Open Questions, References.
- Decision tables put the accepted option on the first row with a bolded verdict.
- Phases have descriptive, self-contained names and declare all required fields.
- The steel thread is explicitly marked and justified as the thinnest viable end-to-end slice.
- Each phase fits in a single context window (split if not) and carries no effort estimates.
- Validation criteria are concrete and risk-based, matched to the right check (unit / repository / handler tests, `make lint` / `make build`, or named manual steps).
- There is **no** code, file paths, function signatures, or snippets anywhere in the plan.
- No optional branches or alternative paths remain — everything is decided and required.

### Iteration expectations for project plans

Typical: 1-3 rounds.

- Fewer rounds (1-2): plan addresses a well-understood area with clear patterns, limited scope, or follows existing architecture closely.
- More rounds (2-3): plan involves novel architecture, cross-repo coordination (Haerd app / admin dashboard), significant design decisions, or explores unfamiliar territory.

If the last round made significant changes to phasing, architecture, or key decisions, another round is usually warranted.

### Edits and updates

- Edits go directly to the `.cursor/plans/<slug>.md` file using the `Write` tool.
- Each round should produce a cleaner, more precise plan.
- Project plans are **high-level** — focus on functionality, flow, and architecture. No code snippets, file paths, or implementation details. Those belong in implementation plans.

Do not proceed to Step 5 until the user explicitly approves the plan.

## Step 5: Commit

After the plan is approved locally (Step 4 complete), remind the user to commit it:

> **Plan ready.** The plan is at `.cursor/plans/<slug>.md` and the supporting research is at `.cursor/plans/<slug>-research.md`. Next steps:
>
> 1. Commit the plan and research doc to a branch. Both files live in the repo as durable artifacts.
> 2. For each named phase, run `/create-implementation-plan` to produce the code-level plan, then `/critique-plan` to stress-test it before implementing.
>
> Keep the project plan as the source of truth for _what_ and _why_; the implementation plans handle _how_.
