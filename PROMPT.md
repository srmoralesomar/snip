# Snip — Clipboard History Manager CLI

You are building `snip`, a clipboard history manager CLI tool written in Go.

---

## What is snip?

A **lightweight CLI** that:
- Runs as a background daemon
- Watches the system clipboard
- Stores a history of copied items in a local database
- Lets the user search, recall, and paste any entry

**Not** a GUI app, **not** a cloud service, **not** a library. One binary, one config directory (`~/.snip/`), standard Unix CLI behavior.

---

## Your process

1. **Read `prd.json`.** Among tasks with `"done": false`, choose ONE task by priority (see Scope and task selection).
2. **Check the codebase.** There may be partial work from a prior run. Run `git log --oneline -10` and skim relevant files so you don’t duplicate or break existing work.
3. **Implement the task.** Follow these standards:
   - Go modules (`go mod`)
   - `cobra` for CLI commands
   - `bbolt` for local storage (embedded key-value DB, no external deps)
   - Fuzzy matching via `github.com/sahilm/fuzzy`
   - Clipboard via `github.com/atotto/clipboard`
   - Config via `github.com/spf13/viper`
   - Code layout: `cmd/` (CLI commands), `internal/` (business logic)

   You may commit multiple times while working on the task (e.g. one commit per logical subtask);
4. **Run feedback loops before each commit.** Before **any** commit, **all** of these must pass:
   - `go test ./...`
   - `go build -o snip .`
   Do **not** commit if any fail. Fix issues first.
5. **When the task is fully done** (all acceptance criteria met, tests and build pass):
   - Set `"done": true` for **only** the task you completed in `prd.json`. Do not change acceptance criteria or other tasks.
   - **Append** to `progress.txt` (do not overwrite). Keep it short:
     - Task ID and title completed
     - Key decisions or reasoning (if any)
     - Files changed
     - Blockers or notes
   - Make a final commit that includes the `prd.json` and `progress.txt` updates: `git add -A && git commit -m "<descriptive message>"`
   - Exit. Your job is done.
6. **If stuck or tests keep failing after serious attempts:**
   - Commit partial progress with message starting with `WIP:`
   - Append what went wrong to `progress.txt`
   - Exit. Your job is done.

---

## Scope and task selection

- **Source of truth:** `prd.json` in the project root.
- **Which task to pick:** Among tasks with `"done": false`, choose the one with the **highest priority** (one task per run). Use this order:
  1. Architectural decisions and core abstractions (scaffolding, storage, daemon core)
  2. Integration points between modules (e.g. wiring daemon to storage)
  3. Unknown unknowns and spike work
  4. Standard features and implementation (individual commands)
  5. Polish, cleanup, and quick wins (config, status, README, formatting)
- **Dependencies:** Do not pick a task if another task it depends on is still undone (e.g. don't do "wire daemon to storage" before "storage layer" exists). The tasks in `prd.json` are listed in rough dependency order; when in doubt, the **first** undone task is the right choice.
- **Acceptance criteria are mandatory.** Each task has an `acceptance` array. Every item must be satisfied. Do not reinterpret, relax, or skip criteria. If you're unsure whether something “counts,” it counts.

---

## Task size and quality

- **You must complete only one PRD task.** Do not complete more than one task.
- **Within that task, prefer multiple small commits.** If a task feels large, break it into logical subtasks: implement one subtask, run tests and build, commit, then continue. Do not wait until the whole task is done to make your first commit. One huge commit at the end is harder to review and roll back; small commits give better feedback and history.
- **Quality over speed.** This is production-style code: maintainable, tested, and consistent. No shortcuts. The patterns you introduce will be reused; corners you cut will compound. Leave the codebase in better shape than you found it.

---

## Completion

After updating `prd.json`, check: **are all tasks `"done": true`?**

- **If yes:** Run the full test suite once more, run `go vet ./...`, ensure `go build` succeeds, then output: `<promise>COMPLETE</promise>`
- **If no:** Exit normally. The loop will start again and the next run will pick the next task.

---

## Constraints (do not violate)

- Do **not** work on more than one task.
- Do **not** mark a task done unless **all** acceptance criteria are met and tests + build pass.
- Do **not** modify acceptance criteria or other tasks in `prd.json` — only set `"done": true` for the task you completed.