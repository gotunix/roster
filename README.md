# Roster

A high-performance CLI engine for managing Ansible inventories using standard YAML formats, with native **NetBox** synchronization.

## Features
- **Git-First**: Designed to manage YAML files that live alongside your code.
- **NetBox Sync**: Seamlessly import hosts, groups, config contexts, and network details from NetBox.
- **Hierarchical View**: Beautiful, adaptive tree-based dashboard that scales to any terminal size.
- **Deep Variable Management**: Support for nested variables and recursive merging.
- **Standard YAML**: No custom drivers needed; works with standard `ansible-playbook`.
- **Export & Reporting**: Aggregate data across multiple inventories with custom headers and dot-notation support.

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
roster dashboard           # Full overview
roster dashboard <group>   # Focused group view
roster dashboard groups    # Global group hierarchy only
```

### NetBox Sync
Keep your inventory in sync with NetBox (requires `NETBOX_TOKEN`):
```bash
# Sync all active devices and VMs
export NETBOX_TOKEN="your_token"
roster sync netbox https://netbox.example.com --filter "status=active"
```

### Manage Hosts
```bash
roster host add <host1,host2,...>
roster host list [group]            # Compact grid
roster host list groups             # Detailed tree grid
roster host view <hostname>
roster host edit <hostname> [-e]    # Use -e for external $EDITOR
roster host remove <hostname>
```

### Manage Groups
```bash
roster group add <group1,group2,...>
roster group assign <h1,h2> <g1,g2> # Many-to-many assignment
roster group nest <child> <parent>  # Hierarchical grouping
roster group list                   # Multi-column list with hierarchy info
roster group view <groupname>
```

### Export & Reporting
```bash
# Export with custom headers and nested variables
roster export --vars "site:Location, networking.ip:Internal IP" -o report.csv

# Email report directly
roster export --email admin@example.com
```

## Configuration

### SMTP Settings (for Email Export)
- `ROSTER_SMTP_HOST`, `ROSTER_SMTP_PORT`
- `ROSTER_SMTP_USER`, `ROSTER_SMTP_PASS` (Optional)
- `ROSTER_SMTP_FROM`

## UI & Colors
Roster adapts to your terminal size:
- **Folders/Groups**: Magenta 📂
- **Hosts**: Green 🖥
- **Nested Groups**: Cyan 📂
- **Metadata/Descriptions**: Subtle Gray

## License
GNU General Public License v3.0 or later. See [LICENSE](LICENSE) for details.
