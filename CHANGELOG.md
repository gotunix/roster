# Changelog

All notable changes to the **Roster** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-07-11

### Changed
- **Refactored `internal/tui/menu.go`**: Split the monolithic 2181-line file into 7 focused files (`view.go`, `update.go`, `tree.go`, `vars.go`, `forms.go`, `runners.go`, `menu.go`) for better maintainability.
- **Deduplicated CSV export logic**: Extracted shared export column building, host iteration, and CSV writing into `internal/store/export.go`. Both the CLI `roster export` command and the TUI export feature now call a single `store.ExportInventory` function.

## [0.2.0] - 2026-07-09

### Added
- **Unified TUI Console**: Launched via `roster menu` to unify all roster operations in a single Bubble Tea terminal interface.
- **Collapsible Inventory Explorer Tree**: Beautiful hierarchy tree showing groups and hosts recursively with fold toggles (`▶`/`▼`) and default collapsed state.
- **Full Screen Dynamic Borders**: Standardized all TUI views, error screens, forms, and results cards to scale borders to fill the full terminal width.
- **Variables CRUD Editor**: Created a variables editor in the TUI. Allows adding, editing, and deleting variables directly on any host or group.
- **Nested Variables Support**:
    - Automatic flattening of nested variables (e.g. `nested.sub.key`) in the TUI listing.
    - Added recursive setters and deleters in the back-end store to traverse and build YAML dictionary hierarchies.
- **Vertical Scrollable Selectors**:
    - Replaced text boxes with list selectors for managing hosts and groups, preventing typographical errors.
    - Implemented a 10-line vertical scroll viewport with top/bottom overflow indicators (`▲`/`▼`) for list navigation.
    - Multi-select capability (`[✔]` checkmarks) enabling batch additions, assignments, nesting, and deletions.
- **Asynchronous NetBox Sync & Live Logging**:
    - Rewrote NetBox sync to run asynchronously in a background thread to prevent UI freezing.
    - Added a thread-safe `SafeBuffer` and piped real-time synchronization outputs directly into a scrolling log viewport in the TUI.
- **CSV Export Result Cards**: Displays completion status, row counts, output paths, and target addresses on card views in the TUI.

## [0.1.0] - 2026-06-10

### Changed
- **Project Rename**: Migrated module from `github.com/user/roster` to `gotunix.net/roster`.
- **UI Scaling**: Removed terminal width constraints; UI now expands to fill any terminal size (e.g., full-screen PuTTY).
- **Dashboard Layout**: Refactored to show the `ALL` group at the bottom with a multi-column host grid.
- **List Views**: Redesigned `host list` and `group list` with multi-column grids for higher information density.

### Added
- **NetBox Integration**:
    - `roster sync netbox <url>`: Sync devices, VMs, interfaces, and disks from NetBox.
    - **Smart Mapping**: Maps `primary_ip` to `ansible_host` and syncs host descriptions.
    - **Config Contexts**: Full support for syncing NetBox Config Contexts to both hosts and groups.
    - **Disaster Recovery**: Syncs detailed interface data (MAC, MTU, IPs/Prefixes) and VM disks.
    - **Pagination**: Supports large inventories with automatic API pagination.
- **Core Engine Enhancements**:
    - **Deep Merging**: Implemented recursive variable merging to support nested YAML structures.
    - **File Locking**: Added concurrency protection via `.roster.lock` to prevent data corruption during bulk updates.
    - **Group Nesting**: Added `roster group nest` command for parent/child group relationships.
- **Bulk Operations**:
    - Support for comma-separated hostnames in `roster host add`.
    - Many-to-many assignments in `roster group assign <host1,host2> <group1,group2>`.
- **Advanced Exporting**:
    - **Custom Headers**: Support for `var:Label` syntax to rename CSV columns.
    - **Dot Notation**: Support for accessing nested variables (e.g., `networking.ip`).
    - **Inheritance**: Exporter now respects full Ansible variable precedence.
- **Interactive Features**:
    - Added `--editor` (`-e`) flag to `host edit` and `group edit` to force use of system `$EDITOR`.
    - Redesigned `host view` and `group view` with multi-column grids to prevent UI breakage with large sets of data.
- **Documentation**: Comprehensive `README.md` updates covering sync, export, and configuration.

### Fixed
- **Stability**: Fixed multiple `SIGSEGV` panics related to nil map assignments and context handling.
- **Build**: Added `-buildvcs=false` to `Makefile` to prevent Git ownership errors (exit status 128) in restrictive environments.
- **Reliability**: Improved error reporting for missing SMTP environment variables.
