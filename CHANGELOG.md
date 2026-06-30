# Changelog

All notable changes to the **Roster** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
