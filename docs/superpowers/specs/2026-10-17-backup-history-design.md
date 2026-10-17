# Backup History and Selected File Actions

## Goal

Integrate the backup and history features from `Downloads/test.zip` into the
current Wails project. The original zip and the friend's branch remain
untouched. Keep the current Wails application and connect these features to
Poom's `database/image.db`.

## Scope

Implement these features:

- Select files in the source folder and copy or move only the selected files.
- Select files in the destination folder and delete them after confirmation.
- Show backup history for the current destination folder.
- Keep history rows after deleting the destination file, so missing files can
  still be found through `FindMissingFiles`.

Do not add image analysis, Gemini, or Tag Cloud. Do not copy generated folders,
dependencies, build output, or executables from the zip.

## Database

Use one database file: the existing Poom `database/image.db`. Keep the Poom
`Image` and `Tag` tables and add `BackupRecord` to the same database. Retain the
`DATABASE` environment override for tests and alternate local runs; tests must
use temporary database files and must not write test rows into the real DB.

Successful selected copies create a `BackupRecord` with status `copied`.
Successful selected moves create a record with status `moved`. History is
queried by the absolute destination path and ordered newest first.

## User flow

On the source page, users select files and choose copy or move. Copy must not
overwrite an existing destination file; it reports a clear error. Move follows
the existing move behavior and skips a destination file that already exists.
After a successful action, refresh both file lists and clear the selection.

On the destination page, users select files and confirm before deleting them.
Deletion removes the physical file but leaves its history row in the database.

The history page shows the file name, source path, status, and creation time for
the selected destination folder. It should explain when no destination is set
or no history exists.

## Backend and frontend changes

- Add selected-file copy, move, and delete methods to the Wails `App` binding.
- Reuse the current `BackupRecord`, `MoveFiles`, `GetHistory`, and
  `FindMissingFiles` behavior where possible.
- Add selection support to `FileList.vue` and connect the pages through
  `useFolders.ts` and `plugins/folders.ts`.
- Add a history page and a History navigation tab.
- Keep code straightforward and consistent with the project's beginner-friendly
  style.

## Validation

- Add or update Go tests using temporary source/destination folders and a
  temporary SQLite database.
- Verify copy and move create the correct history status, duplicate handling
  does not overwrite files, delete preserves history, and history is filtered
  by destination.
- Run `go test -count=1 ./...` and the frontend's Wails-compatible generation
  or build command.
- Manually confirm that a selected copy/move appears in `database/image.db` and
  in the History page.
