# Roster

A high-performance CLI engine for managing Ansible inventories using standard YAML formats.

## Features
- **Git-First**: Designed to manage YAML files that live alongside your code.
- **Hierarchical View**: Beautiful tree-based dashboard.
- **Variable Management**: Simple CLI to set and view `host_vars` and `group_vars`.
- **Standard YAML**: No custom drivers needed; works with standard `ansible-playbook`.

## Usage

### Global Flags
- `-i, --inventory <dir>`: Path to the inventory directory (default: `.`)

### Initialize an Inventory
```bash
roster init -i [directory]
```

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

### Dashboard
```bash
roster dashboard
```

## UI & Colors
Roster shares the same visual identity as the **Metadata** app:
- **Folders/Groups**: Magenta 📂
- **Hosts**: Green 🖥
- **Variables**: Cyan
