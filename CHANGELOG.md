# Changelog

All notable changes to the **Roster** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-10

### Changed
- **Project Rename**: Migrated module from `github.com/user/roster` to `gotunix.net/roster`.

### Added
- **Core Engine**: High-performance YAML-based inventory store supporting standard Ansible structures (`main.yaml`, `host_vars/`, `group_vars/`).
- **Interactive UI (Charm/BubbleTea)**:
    - **Dashboard**: Hierarchical tree view (`roster dashboard`) visualizing groups, hosts, and children relationships.
    - **Enhanced Host Display**: Hosts now show their `description` variable inline (e.g., `hostname (description)`) with improved color coding for tree structure.
    - **Host/Group Views**: Detailed aggregate views with inherited variable resolution and Lipgloss-styled borders.
    - **Interactive Forms**: User-friendly prompts for host creation and variable editing.
    - **Editor Integration**: Configured external editor (`Ctrl+E`) to use `.yaml` extension for automatic syntax highlighting.
- **Multi-Inventory Support**:
    - Global `-i / --inventory` flag now supports multiple paths for aggregate operations.
    - `roster export`: New command to aggregate hosts from multiple inventories into a single CSV report.
    - **Exclusion Filters**: Added `-e / --exclude` flag to `roster export` to filter out hosts belonging to specific groups.
    - **Email Integration**: Added `--email` flag to `roster export` to send generated reports directly via SMTP (configured via environment variables).
- **Host Management**:
    - CRUD operations for inventory hosts.
    - `roster host move`: Sophisticated host migration between inventories, including automatic transfer of `host_vars` and group membership.
- **Group Management**:
    - Comprehensive group listing and view commands.
    - `roster group copy`: Tooling to clone groups, their structure, and associated `group_vars`.
    - Host-to-group assignment logic with multi-file synchronization.
- **Variable Management**:
    - Dedicated CLI commands (`roster vars set`, `roster vars edit`) for managing Ansible variables.
    - YAML-aware variable formatting with syntax highlighting in the terminal.
- **Scaffolding & Tooling**:
    - `roster init`: Bootstraps new Ansible inventories with standard directory hierarchies.
    - **Makefile**: Streamlined build system for compilation (`make build`), installation (`make install`), and environment cleanup.
- **Ansible Environment**:
    - Included `docker.yaml` for containerized inventory testing.
    - `requirements.txt` and `requirements.yaml` for managing Python and Ansible collection dependencies.
- **Documentation**: Initial `README.md` with usage instructions and feature overview.
