# Multiplayer Map-Based Chat System (GoLang)

This project implements a concurrent multiplayer system in Go where players can:
- Connect to the game
- Move between maps (zones)
- Send and receive messages within their current map

Each player has their own message channel, and all maps support concurrent communication through goroutines and mutexes.

---

## 🛠 Features

- ✅ Connect a player (case-insensitive names, e.g., `john` → `John`)
- ✅ Add players to maps dynamically
- ✅ Switch between maps safely
- ✅ Message broadcasting (within the same map)
- ✅ Fully concurrent using goroutines, mutexes, and channels
- ✅ Thread-safe for multi-client support

---

## 🚀 Getting Started

### Prerequisites
- Go 1.16 or newer

### Running the code
You can use this package in a Go application by importing it or running tests.

To run tests (if you have them):
```bash
go test