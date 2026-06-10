# Roster

A high-performance CLI engine for managing Ansible inventories using standard YAML formats.

## Features
- **Git-First**: Designed to manage YAML files that live alongside your code.
- **Hierarchical View**: Beautiful tree-based dashboard.
- **Variable Management**: Simple CLI to set and view `host_vars` and `group_vars`.
- **Standard YAML**: No custom drivers needed; works with standard `ansible-playbook`.
- **Export & Reporting**: Aggregate data across multiple inventories to CSV or email reports.

## Installation

### Prerequisites
- **Go**: 1.26.3 or higher

### Build from Source
```bash
git clone https://gotunix.net/roster.git
cd roster
make build
sudo make install
```

## Usage

### Global Flags
- `-i, --inventory <dir>`: Path to the inventory directory (default: `.`). Can be specified multiple times for aggregate operations.

### Initialize an Inventory
```bash
roster init -i [directory]
```

### Dashboard
```bash
roster dashboard
```
Visualizes your inventory in a hierarchical tree. Use `Ctrl+E` on a selected host or group to open your `$EDITOR` with automatic YAML syntax highlighting.

### Manage Hosts
```bash
roster host add <hostname>
roster host list
roster host view <hostname>
roster host edit <hostname>
roster host move <hostname> <destination_dir>
roster host remove <hostname>
```

### Manage Groups
```bash
roster group add <groupname>
roster group assign <hostname> <group1,group2,...>
roster group copy <source_group> <dest_group>
roster group list
roster group view <groupname>
roster group edit <groupname>
```

### Manage Variables
```bash
roster vars set <host|group> <name> <key>=<value>
roster vars edit <host|group> <name>
```

### Export & Reporting
Aggregate host data across multiple inventories:
```bash
# Export to CSV
roster export -i ./inv1 -i ./inv2 --vars ansible_host,ansible_user -o report.csv

# Exclude specific groups
roster export --exclude production,testing

# Email report directly
roster export --email admin@example.com
```

## Configuration

### SMTP Settings
To use the `--email` feature in `roster export`, configure the following environment variables:
- `ROSTER_SMTP_HOST`: SMTP server address
- `ROSTER_SMTP_PORT`: SMTP server port
- `ROSTER_SMTP_USER`: SMTP username
- `ROSTER_SMTP_PASS`: SMTP password
- `ROSTER_SMTP_FROM`: Sender email address

## UI & Colors
Roster shares a consistent visual identity:
- **Folders/Groups**: Magenta 📂
- **Hosts**: Green 🖥
- **Variables**: Cyan

## License
GNU General Public License v3.0 or later. See [LICENSE](LICENSE) for details.
