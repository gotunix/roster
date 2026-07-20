# TODO

## Organization

- [x] Split `internal/tui/menu.go` (~2182 lines) into per-state/per-view files
- [ ] Split `internal/store/store.go` (~1348 lines) into separate files (inventory.go, vars.go, config.go, etc.)
- [ ] `LoadRosterConf` returns 3 unnamed values — use a struct instead

## Bugs & Fragile Code

- [ ] **Tree filter state leak**: pressing `esc` on the vars screen clears the filter but doesn't exit to menu (needs two `esc` presses) — `menu.go`
- [ ] **`sort.Slice` intransitivity**: forcing `"all"` group to front of sorted list is fragile — `menu.go:255`
- [ ] **Lock not scoped to directory**: `globalFlock` means concurrent syncs on different inventories collide — `store.go:46`
- [ ] **`topoSortGroups` silently drops cycles** instead of erroring
- [ ] **Export result card**: shows `"Exported N hosts to "` with empty path when output is stdout

## Dead / Stale Code

- [ ] Remove stale `Version = "v0.1.0"` in `internal/ui/styles.go` (real version is `v0.3.0` in `internal/version/`)
- [ ] Remove dead nil checks in `MergeHostVars`/`MergeGroupVars` — `GetHostVars`/`GetGroupVars` always return non-nil maps

## Missing Test Coverage

- [ ] `internal/interactive/` — no tests
- [ ] `internal/email/` — no tests
- [ ] `cmd/roster/` — no end-to-end tests
- [ ] `internal/netbox/` — only `MergeNetboxVars` tested (no sync/integration tests)

## Security

- [ ] **SMTP credentials sent in plaintext**: `smtp.PlainAuth` with no TLS enforcement in `internal/email/email.go:129`. `smtp.SendMail` connects without TLS by default; credentials leak on servers without STARTTLS. No implicit TLS (port 465) support.
- [ ] **World-readable file permissions (0644)**: All YAML, vars, and CSV files written with `0644`. Inventory variables may contain secrets readable by any local user.
- [ ] **Path traversal via hostnames/group names**: `filepath.Join(baseDir, "host_vars", hostname+".yaml")` — a hostname like `../malicious` escapes the inventory directory. Affects all `GetHostVars`, `SetHostVar`, `RemoveHost`, etc. in `internal/store/store.go`.
- [ ] **Unsanitized `$EDITOR` execution**: `os.Getenv("EDITOR")` passed to `exec.Command` in `internal/tui/update.go:260,300` and `internal/interactive/forms.go:54,76`. Mitigated by no shell invocation, but still user-controlled program execution.
- [ ] **NetBox token & SMTP passwords in plaintext on disk**: Stored unencrypted in `roster.conf` / `.roster.conf`.

## API Inconsistency

- [ ] `SaveGroup` takes `*models.Group` but other store functions take individual params
