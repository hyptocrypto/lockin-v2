```
    __               __   ____    
   / /   ____  _____/ /__/  _/___ 
  / /   / __ \/ ___/ //_// // __ \
 / /___/ /_/ / /__/ ,< _/ // / / /
/_____/\____/\___/_/|_/___/_/ /_/ 
```

A lightweight, encrypted credentials manager for the terminal.

## Install

```bash
git clone https://github.com/yourusername/lockin.git
cd lockin
chmod +x install.sh && ./install.sh
```

## SMB Sync

Use your vault across multiple devices without ever sending data over the public internet. Configure a local network share in `~/.config/lockin/config.yaml`:

```yaml
smb:
  enabled: true
  host: "192.168.1.100"
  port: "445"
  share: "backups"
  user: "username"
  password: "password"
```

Your encrypted vault syncs automatically after each change — all traffic stays on your local network.

## Usage

```bash
lockin
```

On first run, you'll create a master password. Use the keyboard to navigate:

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate list |
| `a` | Add new entry |
| `e` | Edit selected |
| `d` | Delete selected |
| `/` | Search |
| `c` | Copy password |
| `q` | Quit |

## License

MIT
