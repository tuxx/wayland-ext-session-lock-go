// Package ext_session_lock implements the ext_session_lock_v1 protocol
package ext_session_lock

import (
	"github.com/neurlang/wayland/wl"
)

// Basic type aliases for compatibility
type BaseProxy = wl.BaseProxy
type Event = wl.Event
type Context = wl.Context
type Proxy = wl.Proxy
type WlSurface = wl.Surface
type WlOutput = wl.Output

// BindSessionLockManager binds to the ext_session_lock_manager_v1 interface
func BindSessionLockManager(r *wl.Registry, name uint32, version uint32) *SessionLockManager {
	// Get the context from the registry
	ctx, _ := wl.GetUserData[wl.Context](r)

	// Create a new manager instance
	manager := NewSessionLockManager(ctx)

	// Bind it to the interface
	_ = r.Bind(name, "ext_session_lock_manager_v1", version, manager)

	return manager
}

// Helper functions to add listeners

// SessionLockManagerAddListener adds a listener for session lock manager events
// No events for the manager currently
func SessionLockManagerAddListener(m *SessionLockManager, h interface{}) {
	// No events to listen for
}

// SessionLockAddListener adds all listeners for session lock events
func SessionLockAddListener(l *SessionLock, h interface{}) {
	if handler, ok := h.(SessionLockLockedHandler); ok {
		l.AddLockedHandler(handler)
	}
	if handler, ok := h.(SessionLockFinishedHandler); ok {
		l.AddFinishedHandler(handler)
	}
}

// SessionLockSurfaceAddListener adds a listener for session lock surface events
func SessionLockSurfaceAddListener(s *SessionLockSurface, h interface{}) {
	if handler, ok := h.(SessionLockSurfaceConfigureHandler); ok {
		s.AddConfigureHandler(handler)
	}
}
