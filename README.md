# GoST — Modular ECS Terminal Emulator

GoST is a lightweight, modular **Go + Ebiten terminal emulator** built around a custom **Entity–Component–System (ECS)** architecture.  
It simulates a terminal screen using ANSI parsing, PTY integration, and a 2D renderer, all designed for clarity, speed, and extensibility.

---

## 🚀 Features

- ✅ Full **ECS architecture** — each subsystem runs independently.
- ✅ **PTY-backed shell** (runs your `/bin/bash` or `$SHELL`).
- ✅ **ANSI parser** — supports colors, cursor movement, clearing, etc.
- ✅ **Bitmap terminal renderer** — fast and lightweight with basic fonts.
- ✅ **Keyboard input system** with keymap handling.
- ✅ **Mouse drag selection** — copy text to system clipboard.
- ✅ **Static cursor overlay** showing current input position.
- ✅ Modular directory structure — easy to expand or swap systems.

---

## 🧩 Project Structure

```

gost/
├── cmd/
│   └── gost/
│       └── main.go           # Ebiten main loop & ECS bootstrap
├── internal/
│   ├── ecs/
│   │   └── world.go          # ECS world manager
│   ├── events/
│   │   └── bus.go            # Pub/sub message bus
│   ├── components/
│   │   └── term_buffer.go    # Terminal buffer model
│   └── systems/
│       ├── input/            # Keyboard and keymap handling
│       │   ├── keys.go
│       │   ├── map.go
│       │   ├── system.go
│       │   └── util.go
│       ├── parser/           # ANSI parser
│       │   └── system.go
│       ├── pty/              # PTY bridge to real shell
│       │   └── system.go
│       ├── render/           # Terminal renderer
│       │   └── system.go
│       ├── cursor/           # Cursor overlay
│       │   └── system.go
│       └── selection/        # Mouse selection and clipboard
│           └── system.go

````

Each subsystem implements the `ecs.System` interface:

```go
type System interface {
    UpdateECS()
}
````

All communication occurs via the central **event bus** (`internal/events/bus.go`),
so systems stay fully decoupled yet synchronized.

---

## 🧠 Architecture Overview

| Layer         | Role                                            |
| ------------- | ----------------------------------------------- |
| **ECS**       | Owns system list, runs updates each frame       |
| **Bus**       | Decouples systems (publish/subscribe)           |
| **PTY**       | Spawns interactive shell and streams output     |
| **Parser**    | Interprets shell output (ANSI, cursor, etc.)    |
| **Render**    | Draws text buffer to screen using Ebiten        |
| **Input**     | Handles key events, sends them to PTY           |
| **Cursor**    | Draws static cursor overlay                     |
| **Selection** | Handles mouse drag selection and clipboard copy |

---

## 🧰 Development

### Run the Terminal

```bash
go mod tidy
go run ./cmd/gost
```

You’ll see:

* A black terminal window running your shell
* Normal typing and command execution
* ANSI colors, cursor, and backspace working correctly
* Drag-select to copy text to clipboard

### Debug Logging

`main.go` logs system startup and shell status:

```bash
GoST — modular ECS terminal emulator starting...
[PTYSystem] started shell: /bin/bash
```

---

## 🧩 Extending GoST

You can easily add new systems or features:

1. Create a new directory under `internal/systems/<your_system>/`
2. Add a `system.go` implementing:

   ```go
   type System struct { bus *events.Bus }
   func (s *System) UpdateECS() { ... }
   ```
3. Subscribe or publish to the event bus as needed.
4. Register it in `cmd/gost/main.go`:

   ```go
   newSys := your_system.NewSystem(bus)
   world.AddSystem(newSys)
   ```

This modular pattern supports:

* Status overlays
* Debug widgets
* Network shells
* Logging or replay tools
* Future VT100 / UTF-8 feature sets

---

## 🧱 Dependencies

| Package                                                                                     | Purpose                          |
| ------------------------------------------------------------------------------------------- | -------------------------------- |
| [`github.com/hajimehoshi/ebiten/v2`](https://github.com/hajimehoshi/ebiten)                 | Game loop & rendering            |
| [`github.com/creack/pty`](https://github.com/creack/pty)                                    | Pseudo-terminal shell backend    |
| [`golang.org/x/image/font/basicfont`](https://pkg.go.dev/golang.org/x/image/font/basicfont) | Built-in bitmap font             |
| [`github.com/atotto/clipboard`](https://github.com/atotto/clipboard)                        | Cross-platform clipboard support |

All dependencies are pure Go and compile cross-platform (Linux, macOS, Windows).

---

## 🧩 Future Enhancements

* [ ] UTF-8 + wide-character rendering
* [ ] Scrollback history buffer
* [ ] Custom font loader
* [ ] Configurable keymaps
* [ ] Split-pane support
* [ ] GPU-accelerated text rendering

---

## 🧑‍💻 License

MIT © 2025 — David Winfrey
GoST is open-source and hackable. Build your own terminal, game shell, or graphical REPL!

