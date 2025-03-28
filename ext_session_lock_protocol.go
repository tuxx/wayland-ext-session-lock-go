// Package ext_session_lock implements the ext_session_lock_v1 protocol
package ext_session_lock

import (
	"sync"
	// wl is used indirectly through type aliases
	"github.com/neurlang/wayland/wl"
)

// Use the wl package explicitly in type declarations to avoid the "imported and not used" error
var _ wl.BaseProxy // This line ensures the wl package is used

// Error constants for ext_session_lock_v1
const (
	LockErrorInvalidDestroy uint32 = iota
	LockErrorInvalidUnlock
	LockErrorRole
	LockErrorDuplicateOutput
	LockErrorAlreadyConstructed
)

// Error constants for ext_session_lock_surface_v1
const (
	SurfaceErrorCommitBeforeFirstAck uint32 = iota
	SurfaceErrorNullBuffer
	SurfaceErrorDimensionsMismatch
	SurfaceErrorInvalidSerial
)

// Protocol request/event constants for ext_session_lock_manager_v1
const (
	ManagerRequestDestroy uint32 = iota
	ManagerRequestLock
)

// Protocol request/event constants for ext_session_lock_v1
const (
	LockRequestDestroy uint32 = iota
	LockRequestGetLockSurface
	LockRequestUnlockAndDestroy
)

const (
	LockEventLocked uint32 = iota
	LockEventFinished
)

// Protocol request/event constants for ext_session_lock_surface_v1
const (
	SurfaceRequestDestroy uint32 = iota
	SurfaceRequestAckConfigure
)

const (
	SurfaceEventConfigure uint32 = iota
)

// SessionLockManager represents an ext_session_lock_manager_v1 object
type SessionLockManager struct {
	BaseProxy
}

// NewSessionLockManager is a constructor for the SessionLockManager object
func NewSessionLockManager(ctx *Context) *SessionLockManager {
	ret := new(SessionLockManager)
	ctx.Register(ret)
	return ret
}

// Destroy destroys the session lock manager object
func (m *SessionLockManager) Destroy() error {
	return m.Context().SendRequest(m, ManagerRequestDestroy)
}

// Lock attempts to lock the session
func (m *SessionLockManager) Lock() (*SessionLock, error) {
	retId := NewSessionLock(m.Context())
	return retId, m.Context().SendRequest(m, ManagerRequestLock, retId)
}

// Dispatch dispatches event for SessionLockManager
func (m *SessionLockManager) Dispatch(event *Event) {
	// No events to dispatch for the manager
}

// SessionLock represents an ext_session_lock_v1 object
type SessionLock struct {
	BaseProxy
	mu                         sync.RWMutex
	privateSessionLockLocked   []SessionLockLockedHandler
	privateSessionLockFinished []SessionLockFinishedHandler
}

// NewSessionLock is a constructor for the SessionLock object
func NewSessionLock(ctx *Context) *SessionLock {
	ret := new(SessionLock)
	ctx.Register(ret)
	return ret
}

// Destroy destroys the session lock
func (l *SessionLock) Destroy() error {
	return l.Context().SendRequest(l, LockRequestDestroy)
}

// GetLockSurface creates a lock surface for a given output
func (l *SessionLock) GetLockSurface(surface *WlSurface, output *WlOutput) (*SessionLockSurface, error) {
	retId := NewSessionLockSurface(l.Context())
	return retId, l.Context().SendRequest(l, LockRequestGetLockSurface, retId, surface, output)
}

// UnlockAndDestroy unlocks the session and destroys the object
func (l *SessionLock) UnlockAndDestroy() error {
	return l.Context().SendRequest(l, LockRequestUnlockAndDestroy)
}

// Dispatch dispatches event for SessionLock
func (l *SessionLock) Dispatch(event *Event) {
	switch event.Opcode {
	case LockEventLocked:
		if len(l.privateSessionLockLocked) > 0 {
			ev := SessionLockLockedEvent{}
			l.mu.RLock()
			for _, h := range l.privateSessionLockLocked {
				h.HandleSessionLockLocked(ev)
			}
			l.mu.RUnlock()
		}
	case LockEventFinished:
		if len(l.privateSessionLockFinished) > 0 {
			ev := SessionLockFinishedEvent{}
			l.mu.RLock()
			for _, h := range l.privateSessionLockFinished {
				h.HandleSessionLockFinished(ev)
			}
			l.mu.RUnlock()
		}
	}
}

// SessionLockLockedEvent represents the locked event
type SessionLockLockedEvent struct {
}

// SessionLockFinishedEvent represents the finished event
type SessionLockFinishedEvent struct {
}

// SessionLockLockedHandler is the handler interface for SessionLockLockedEvent
type SessionLockLockedHandler interface {
	HandleSessionLockLocked(SessionLockLockedEvent)
}

// AddLockedHandler adds the Locked handler
func (l *SessionLock) AddLockedHandler(h SessionLockLockedHandler) {
	if h != nil {
		l.mu.Lock()
		l.privateSessionLockLocked = append(l.privateSessionLockLocked, h)
		l.mu.Unlock()
	}
}

// RemoveLockedHandler removes the Locked handler
func (l *SessionLock) RemoveLockedHandler(h SessionLockLockedHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, e := range l.privateSessionLockLocked {
		if e == h {
			l.privateSessionLockLocked = append(l.privateSessionLockLocked[:i], l.privateSessionLockLocked[i+1:]...)
			break
		}
	}
}

// SessionLockFinishedHandler is the handler interface for SessionLockFinishedEvent
type SessionLockFinishedHandler interface {
	HandleSessionLockFinished(SessionLockFinishedEvent)
}

// AddFinishedHandler adds the Finished handler
func (l *SessionLock) AddFinishedHandler(h SessionLockFinishedHandler) {
	if h != nil {
		l.mu.Lock()
		l.privateSessionLockFinished = append(l.privateSessionLockFinished, h)
		l.mu.Unlock()
	}
}

// RemoveFinishedHandler removes the Finished handler
func (l *SessionLock) RemoveFinishedHandler(h SessionLockFinishedHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, e := range l.privateSessionLockFinished {
		if e == h {
			l.privateSessionLockFinished = append(l.privateSessionLockFinished[:i], l.privateSessionLockFinished[i+1:]...)
			break
		}
	}
}

// SessionLockSurface represents an ext_session_lock_surface_v1 object
type SessionLockSurface struct {
	BaseProxy
	mu                                 sync.RWMutex
	privateSessionLockSurfaceConfigure []SessionLockSurfaceConfigureHandler
}

// NewSessionLockSurface is a constructor for the SessionLockSurface object
func NewSessionLockSurface(ctx *Context) *SessionLockSurface {
	ret := new(SessionLockSurface)
	ctx.Register(ret)
	return ret
}

// Destroy destroys the lock surface object
func (s *SessionLockSurface) Destroy() error {
	return s.Context().SendRequest(s, SurfaceRequestDestroy)
}

// AckConfigure acknowledges a configure event
func (s *SessionLockSurface) AckConfigure(serial uint32) error {
	return s.Context().SendRequest(s, SurfaceRequestAckConfigure, serial)
}

// Dispatch dispatches event for SessionLockSurface
func (s *SessionLockSurface) Dispatch(event *Event) {
	switch event.Opcode {
	case SurfaceEventConfigure:
		if len(s.privateSessionLockSurfaceConfigure) > 0 {
			ev := SessionLockSurfaceConfigureEvent{}
			ev.Serial = event.Uint32()
			ev.Width = event.Uint32()
			ev.Height = event.Uint32()
			s.mu.RLock()
			for _, h := range s.privateSessionLockSurfaceConfigure {
				h.HandleSessionLockSurfaceConfigure(ev)
			}
			s.mu.RUnlock()
		}
	}
}

// SessionLockSurfaceConfigureEvent represents the configure event
type SessionLockSurfaceConfigureEvent struct {
	Serial uint32
	Width  uint32
	Height uint32
}

// SessionLockSurfaceConfigureHandler is the handler interface for SessionLockSurfaceConfigureEvent
type SessionLockSurfaceConfigureHandler interface {
	HandleSessionLockSurfaceConfigure(SessionLockSurfaceConfigureEvent)
}

// AddConfigureHandler adds the Configure handler
func (s *SessionLockSurface) AddConfigureHandler(h SessionLockSurfaceConfigureHandler) {
	if h != nil {
		s.mu.Lock()
		s.privateSessionLockSurfaceConfigure = append(s.privateSessionLockSurfaceConfigure, h)
		s.mu.Unlock()
	}
}

// RemoveConfigureHandler removes the Configure handler
func (s *SessionLockSurface) RemoveConfigureHandler(h SessionLockSurfaceConfigureHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.privateSessionLockSurfaceConfigure {
		if e == h {
			s.privateSessionLockSurfaceConfigure = append(s.privateSessionLockSurfaceConfigure[:i], s.privateSessionLockSurfaceConfigure[i+1:]...)
			break
		}
	}
}
