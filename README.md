# wayland-ext-session-lock-go

Go language bindings for the `ext_session_lock_v1` Wayland protocol, enabling secure screen locking functionality on compositors such as Hyprland.

## Overview

This project provides Go bindings and a working example client (`simple_locker`) that demonstrates session locking by interfacing with the `ext_session_lock_manager_v1` protocol.

## Features

- Binds to `ext_session_lock_v1` Wayland protocol
- Lock session and create per-output lock surfaces
- Event handlers for lock lifecycle and surface configuration
- Simple reference implementation (`example/simple_locker`)

## Example Output

```
Found wl_compositor (name: 3)
Found ext_session_lock_manager_v1 (name: 27)
Found wl_output (name: 53)
Attempting to lock session...
Session is now locked!
```

## Usage

1. Clone the repo:
   ```bash
   git clone https://github.com/tuxx/wayland-ext-session-lock-go
   cd wayland-ext-session-lock-go
   ```

2. Build the example:
   ```bash
   go build -o simple_locker ./example/simple_locker
   ```

3. Run it under a Wayland session with `ext_session_lock_v1` support:
   ```bash
   ./simple_locker
   ```

## Integrating in Your Own Program

1. Import the module:
   ```go
   import ext "github.com/tuxx/wayland-ext-session-lock-go"
   ```

2. Use `ext.BindSessionLockManager` to bind the protocol and create a `SessionLock`.

3. Register event handlers for lock state and surface configuration.

4. Create `wl.Surface` and `ext.SessionLockSurface` for each `wl.Output`.

See `example/simple_locker/main.go` for a complete reference.

## Dependencies

- [neurlang/wayland](https://github.com/neurlang/wayland)
- Wayland compositor supporting `ext_session_lock_v1` (e.g., Hyprland)

## License

MIT
