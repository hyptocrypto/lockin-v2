```
    __               __   ____    
   / /   ____  _____/ /__/  _/___ 
  / /   / __ \/ ___/ //_// // __ \
 / /___/ /_/ / /__/ ,< _/ // / / /
/_____/\____/\___/_/|_/___/_/ /_/ 
```

# LockIn

A lightweight, encrypted credentials manager for the terminal.

## Features

- 🔐 **AES-256-GCM encryption** with PBKDF2 key derivation (100k iterations)
- 💾 **Local SQLite storage** — your data never leaves your machine
- 🔄 **Optional SMB sync** — share your vault across devices on your local network, never over the public internet
- 📋 **Clipboard support** — copy passwords without displaying them
- ✨ **Beautiful TUI** — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)

## Install

```bash
go install lockin@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/lockin.git
cd lockin
go build -o lockin .
```

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

## License

MIT
