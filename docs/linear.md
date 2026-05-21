# Linear (this repo)

Work for the tutor timetable app is tracked in Linear (team **Kubac**, workspace-specific).

## Tutor Timetable project for issues

**Assign every issue that ships or defines this app** (Phase 1 implementation, Phase 2 features, infra, tests, deployment) to the Linear project **Tutor Timetable**.

Do not assign those issues to separate Linear projects that only represent a portfolio entry or another product. The **issue** always stays on **Tutor Timetable** unless you explicitly choose otherwise.

Project URL: https://linear.app/kuabc/project/tutor-timetable-6767972a978d

## Creating issues via MCP

When using `save_issue`, set `project` to **`Tutor Timetable`** for work in this repository. `team` is **`Kubac`**.

Suggested workflow (same as homepage / trading):

- One Linear ticket → one focused PR
- Title and description reference acceptance criteria from `PLAN.md` / `CLAUDE.md`
- Link the PR in the issue when opening it

## Scope reference

| Phase | Doc |
|-------|-----|
| Phase 1 features, API, schema | [`PLAN.md`](../PLAN.md) |
| Stack, conventions, routes | [`CLAUDE.md`](../CLAUDE.md) |
| Phase 2 (planned) | [`PLAN.md`](../PLAN.md) § Phase 2, [`CLAUDE.md`](../CLAUDE.md) § Phase 2 |

## Phase 1 backlog (suggested order)

Work through **Todo** issues in the [Tutor Timetable project](https://linear.app/kuabc/project/tutor-timetable-6767972a978d) roughly in this order (one ticket → one PR):

| Order | Issue | Focus |
|------:|-------|--------|
| 1 | KUA-104 | Git + GitHub remote |
| 2 | KUA-108 | Isolated ticket Docker/worktree workflow |
| 3 | KUA-91 | Docker smoke test & runbook |
| 4 | KUA-94 | Scheduler unit tests (baseline) |
| 5 | KUA-92 | Scheduler cost function + working window |
| 6 | KUA-93 | Scheduler local search |
| 7 | KUA-95 | `moveLesson` validation |
| 8 | KUA-96 | Timetable drag-and-drop |
| 9 | KUA-97 | Timetable grid hours from settings |
| 10 | KUA-98 | i18n weekdays / copy audit |
| 11 | KUA-99 | API error code translations |
| 12 | KUA-100 | Locale in settings |
| 13 | KUA-101 | Post-generate “not placed” feedback |
| 14 | KUA-102 | Phase 1 manual QA checklist |
| 15 | KUA-103 | CI (Go test + `npm run check`) |
| 16 | KUA-105 | Production deployment |

Phase 2 backlog (after Phase 1 QA): **KUA-106** (share links), **KUA-107** (Czech locale).
