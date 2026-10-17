# Backup History and Selected File Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add selected-file copy, move, and delete actions plus a history page to the current Wails application, using Poom's `database/image.db`.

**Architecture:** Keep the current `App`, Wails startup, `BackupRecord`, and Poom database connection. Add selected-file backend methods in a focused Go file, bind them through Wails, and connect the existing Nuxt pages through `FileList.vue`, `useFolders.ts`, and the folders plugin. Do not include image analysis, Gemini, or Tag Cloud.

**Tech Stack:** Go, Wails v2, GORM/SQLite, Nuxt 4, Vue 3, TypeScript.

**Spec:** `docs/superpowers/specs/2026-10-17-backup-history-design.md`

## Global Constraints

- Use one database file: the existing Poom `database/image.db`.
- Keep the Poom `Image` and `Tag` tables and add `BackupRecord` to the same database.
- Retain the `DATABASE` environment override for tests and alternate local runs; tests must use temporary database files and must not write test rows into the real DB.
- Do not add image analysis, Gemini, or Tag Cloud.
- Do not copy generated folders, dependencies, build output, or executables from the zip.
- Keep code straightforward and consistent with the project's beginner-friendly style.

## Review Focus

- A selected name equal to `..`, containing a path separator, or otherwise not a basename must not read or modify a file outside the chosen folder; test copy, move, and delete validation.
- Copying to an existing destination must fail without changing its contents; moving to an existing destination must skip it and leave the source in place.
- Empty source/destination paths or an empty selection must return a useful error and change no files.
- An unavailable database must stop copy/move before recording history; test the returned error.
- Deleting a destination file must keep its history row so `FindMissingFiles` can report it; test both outcomes.

---

### Task 1: Copy only selected files and record history

**Files:**
- Create: `app_selected.go`
- Modify: `backup_test.go`

**Interfaces:**
- Consumes: `App.db`, `App.dbErr`, `model.BackupRecord`, and existing `App.GetHistory`.
- Produces: `func (a *App) CopySelectedFiles(sourcePath string, destPath string, fileNames []string) (int, error)`.

- [x] **Step 1: Write failing copy tests** in `backup_test.go`:
  - `TestCopySelectedFilesWritesCopiedHistory` checks copied bytes, source remains, count is 1, and the DB row has status `copied`.
  - `TestCopySelectedFilesRefusesExistingDestination` checks an existing destination is unchanged and the call returns an error.
  - `TestCopySelectedFilesRejectsPathTraversal` passes `../outside.txt` and checks no outside file is read or created.
  - `TestCopySelectedFilesRejectsMissingInput` checks blank source/destination paths and an empty selection return errors without changing files.
  - `TestCopySelectedFilesRejectsUnavailableDatabase` checks a DB error stops the operation before any copy.
- [x] **Step 2: Run the tests and confirm they fail** because `CopySelectedFiles` is not defined yet.

  Run: `go test -count=1 -run 'TestCopySelectedFiles' .`

  Expected: FAIL with an undefined-method compile error.
- [x] **Step 3: Implement `CopySelectedFiles`** in `app_selected.go`. Validate app DB state, folder paths, selected basenames, and regular source files. Create destination files with exclusive creation so existing files cannot be overwritten; after a successful copy, insert a `BackupRecord` with status `copied`.
- [x] **Step 4: Run the copy tests** and confirm all five pass.

  Run: `go test -count=1 -run 'TestCopySelectedFiles' .`

  Expected: PASS.

### Task 2: Move selected files and delete selected destination files

**Files:**
- Modify: `app_selected.go`
- Modify: `backup_test.go`

**Interfaces:**
- Consumes: `moveOneFile`, `App.db`, `App.dbErr`, `model.BackupRecord`, and `App.FindMissingFiles`.
- Produces:
  - `func (a *App) MoveSelectedFiles(sourcePath string, destPath string, fileNames []string) (int, error)`
  - `func (a *App) DeleteSelectedFiles(destPath string, fileNames []string) (int, error)`

- [x] **Step 1: Write failing move tests** in `backup_test.go`:
  - `TestMoveSelectedFilesWritesMovedHistory` checks source removal, destination contents, count 1, and status `moved`.
  - `TestMoveSelectedFilesSkipsExistingDestination` checks the destination is unchanged and the source remains.
  - `TestMoveSelectedFilesRejectsPathTraversal` checks a name outside the source folder is rejected.
  - `TestMoveSelectedFilesRejectsMissingInput` checks blank source/destination paths and an empty selection return errors without changing files.
  - `TestMoveSelectedFilesRejectsUnavailableDatabase` checks a DB error stops the operation before any move.
- [x] **Step 2: Run the move tests and confirm they fail** because `MoveSelectedFiles` is not defined.

  Run: `go test -count=1 -run 'TestMoveSelectedFiles' .`

  Expected: FAIL with an undefined-method compile error.
- [x] **Step 3: Implement `MoveSelectedFiles`** in `app_selected.go`. Reuse `moveOneFile`; validate the DB, folder paths, and selected regular file names; skip a destination file that already exists; write a `BackupRecord` with status `moved` after each successful move.
- [x] **Step 4: Run the move tests** and confirm all five pass.

  Run: `go test -count=1 -run 'TestMoveSelectedFiles' .`

  Expected: PASS.
- [x] **Step 5: Write failing delete tests**:
  - `TestDeleteSelectedFilesKeepsHistory` checks that the selected destination file is removed, its DB record remains, and `FindMissingFiles` returns its name.
  - `TestDeleteSelectedFilesRejectsPathTraversal` checks names outside the destination folder are rejected.
  - `TestDeleteSelectedFilesRejectsMissingInput` checks a blank destination path and an empty selection return errors without changing files.
- [x] **Step 6: Run the delete tests and confirm they fail** because `DeleteSelectedFiles` is not defined.

  Run: `go test -count=1 -run 'TestDeleteSelectedFiles' .`

  Expected: FAIL with an undefined-method compile error.
- [x] **Step 7: Implement `DeleteSelectedFiles`** in `app_selected.go`. Validate names and regular files, remove only selected destination files, and leave their `BackupRecord` rows untouched.
- [x] **Step 8: Run the delete tests** and confirm all three pass.

  Run: `go test -count=1 -run 'TestDeleteSelectedFiles' .`

  Expected: PASS.

### Task 3: Connect file selection and actions in the frontend

**Files:**
- Modify: `frontend/app/components/FileList.vue`
- Modify: `frontend/app/composables/useFolders.ts`
- Modify: `frontend/app/plugins/folders.ts`
- Modify: `frontend/app/pages/index.vue`
- Modify: `frontend/app/pages/destination.vue`
- Regenerate: `frontend/wailsjs/go/main/App.js`, `frontend/wailsjs/go/main/App.d.ts`

**Interfaces:**
- Consumes: Wails methods `CopySelectedFiles`, `MoveSelectedFiles`, and `DeleteSelectedFiles` from Tasks 1–2.
- Produces: selectable file rows, source copy/move actions, and a confirmed destination delete action.

- [x] **Step 1: Add selection props and events** to `FileList.vue`; verify selected checkboxes emit the selected file names and reload clears stale selections.
- [x] **Step 2: Add plugin wrappers** for the three Wails methods in `folders.ts`.
- [x] **Step 3: Add selection state and action functions** in `useFolders.ts`. After success, reload the affected file lists, clear selection, and expose the copied/moved counts.
- [x] **Step 4: Add copy/move buttons** to `index.vue` and a delete button with `window.confirm` to `destination.vue`.
- [x] **Step 5: Run `wails dev`** so Wails refreshes generated bindings and starts the app; verify the three method names exist in both generated binding files and the UI opens.

### Task 4: Add the history page and navigation

**Files:**
- Create: `frontend/app/pages/history.vue`
- Modify: `frontend/app/composables/useFolders.ts`
- Modify: `frontend/app/layouts/default.vue`

**Interfaces:**
- Consumes: existing `GetHistory(destinationPath)` and the shared destination path.
- Produces: a History tab and a page listing file name, source path, status, and creation time for the current destination.

- [x] **Step 1: Add a history filter test** named `TestGetHistoryFiltersByDestination` in `backup_test.go`; create records for two destination paths and verify each query returns only the matching record.
- [x] **Step 2: Run the history test** and verify it passes against the existing backend method.

  Run: `go test -count=1 -run TestGetHistoryFiltersByDestination .`

  Expected: PASS.
- [x] **Step 3: Add history state and `reloadHistory`** in `useFolders.ts`; load records for the selected destination and surface any error through the existing error state.
- [x] **Step 4: Create `history.vue`** with empty states for a missing destination and an empty history list, plus the requested history columns.
- [x] **Step 5: Add the History navigation link** to `default.vue`.
- [x] **Step 6: Run the frontend generation command** and confirm the history route and types compile.

  Run from `frontend`: `npm run generate`

  Expected: exit code 0.

### Task 5: Verify the integrated app

**Files:**
- Verify: all files changed by Tasks 1–4.

- [x] **Step 1: Format Go files** with `gofmt -w app_selected.go backup_test.go`.
- [x] **Step 2: Run all Go tests**.

  Run: `go test -count=1 ./...`

  Expected: all Go packages pass; no test rows are written to `database/image.db`.
- [x] **Step 3: Generate the frontend**.

  Run from `frontend`: `npm run generate`

  Expected: exit code 0.
- [x] **Step 4: Review the diff** with `git diff --check` and confirm no generated dependency/build folders or real DB test data were added.
- [x] **Step 5: Manually smoke-test** with temporary source/destination folders: Wails browser mode copied and moved files through Go bindings, and the History page displayed both database rows. The delete confirmation appeared and was dismissed so the temporary file remained; Go tests verify deletion keeps history.
