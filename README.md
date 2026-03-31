# lazysql

A terminal UI for managing SQL databases, inspired by [lazygit](https://github.com/jesseduffield/lazygit).

Navigate connections, browse tables, inspect schemas and run queries — all from your terminal with keyboard-driven navigation.

```
┌─1 Connections (2)─┬──4 Query  NORMAL ─────────────────┐
│ > demo-sqlite      │                                    │
│   prod-pg          │  SELECT * FROM users               │
│                    │  WHERE role = 'admin'              │
├─2 Tables (5)───────│  LIMIT 100;                        │
│ > users            │                                    │
│   orders           │  [i] insert  [ctrl+e] execute     │
│   products         ├──5 Results (3 rows, 2ms) ──────────┤
│   sessions         │                                    │
│   logs             │  id │ email      │ name            │
├─3 Structure (4)────│  1  │ alice@..   │ Alice           │
│ > id INTEGER PK    │  5  │ fiona@..   │ Fiona           │
│   email TEXT       │  8  │ gaston@..  │ Gaston          │
│   name TEXT        │                                    │
│   role TEXT        │                                    │
└────────────────────┴────────────────────────────────────┘
 sqlite | demo.db | 3 rows in 2ms   ? help  S list  q quit
```

## Features

**Multi-database support** — PostgreSQL, MySQL and SQLite with a unified interface.

**Lazygit-style navigation** — Numbered panels, spatial movement (`h`/`l` between columns, `[`/`]` within a column), `tab` cycling, and `1`-`5` direct jump.

**Vim-mode query editor** — `i` to enter INSERT mode, `esc` to return to NORMAL. Global shortcuts stay active in NORMAL mode.

**Table structure** — Auto-populated when selecting a table. Shows column name, type, primary key, nullability, and defaults.

**Saved queries** — Save frequently used queries per connection (`ctrl+s`). Browse and load them from a modal list (`S`). Persisted in the config file.

**Query history** — Navigate previous queries with `ctrl+p` / `ctrl+n`.

**Destructive query protection** — `DELETE`, `DROP`, `TRUNCATE`, and `ALTER` statements require explicit confirmation before execution.

**SSH tunnels** — Connect to remote databases through SSH with key or password authentication.

**Horizontal scroll** — Results with many columns scroll horizontally with `left`/`right` arrows, with a column indicator.

## Install

```bash
go install github.com/lazysql/lazysql@latest
```

Or build from source:

```bash
git clone https://github.com/lazysql/lazysql.git
cd lazysql
go build -o lazysql .
```

Requires Go 1.21+ and CGO enabled (for SQLite support via `mattn/go-sqlite3`).

## Usage

```bash
lazysql
```

On first launch, press `a` to add a connection. The configuration is stored in `~/.config/lazysql/config.yaml`.

### Configuration

```yaml
connections:
  - name: local-pg
    driver: postgres
    host: localhost
    port: 5432
    user: admin
    password: secret
    database: myapp

  - name: production
    driver: mysql
    host: db.example.com
    port: 3306
    user: readonly
    password: hunter2
    database: prod
    ssh:
      host: bastion.example.com
      port: 22
      user: deploy
      key_path: ~/.ssh/id_ed25519

  - name: local-sqlite
    driver: sqlite
    path: ./data/app.db
    saved_queries:
      - name: active-users
        sql: SELECT * FROM users WHERE active = 1
      - name: recent-orders
        sql: SELECT * FROM orders ORDER BY created_at DESC LIMIT 50
```

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Cycle through panels |
| `h` / `l` (or `left` / `right`) | Jump between left and right columns |
| `[` / `]` | Cycle panels within current column |
| `1`-`5` | Jump to panel directly |
| `?` | Help screen |
| `ctrl+d` | Disconnect from database |
| `q` / `ctrl+c` | Quit |

### Connections (panel 1)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate list |
| `enter` | Connect |
| `a` | Add connection |
| `e` | Edit connection |
| `d` | Delete connection |

### Tables (panel 2)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate list |
| `enter` | Generate SELECT query |
| `s` | View table structure |

### Query (panel 4)

| Key | Action |
|-----|--------|
| `i` / `a` | Enter INSERT mode |
| `esc` | Return to NORMAL mode |
| `ctrl+e` / `F5` | Execute query |
| `ctrl+s` | Save query |
| `S` | Open saved queries |
| `ctrl+p` / `ctrl+n` | Query history |

### Results (panel 5)

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll rows |
| `left` / `right` | Scroll columns |

## Architecture

```
pkg/
├── app/          Application bootstrap
├── config/       YAML configuration (connections, saved queries)
├── db/           Database abstraction layer
│   ├── driver.go   Driver interface (Query, Execute, ListTables, DescribeTable)
│   ├── postgres/   PostgreSQL implementation
│   ├── mysql/      MySQL implementation
│   └── sqlite/     SQLite implementation
├── dbconnect/    Connection factory (driver selection, SSH tunnel setup)
├── tunnel/       SSH tunnel management
└── tui/          Terminal UI
    ├── tui.go        Main layout engine and event routing
    ├── keys.go       Keymap definitions
    ├── styles/       Lipgloss styles and color palette
    ├── components/   Reusable UI components (list, table, editor, modal, form, confirm)
    └── views/        Panel views (connections, tables, structure, query, results, help, saved queries)
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Inspiration

The layout and navigation model are directly inspired by [lazygit](https://github.com/jesseduffield/lazygit) — stacked panels on the left, main content on the right, numbered panel titles, spatial navigation between columns, and a contextual status bar. The vim-mode query editor follows the same philosophy: modal editing where shortcuts work in normal mode and text input is explicit.

## License

MIT
