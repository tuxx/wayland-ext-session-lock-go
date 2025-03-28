package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neurlang/wayland/wl"
	"github.com/neurlang/wayland/wlclient"
	ext "github.com/tuxx/wayland-ext-session-lock-go"
)

// LockClient represents our session lock client
type LockClient struct {
	display        *wl.Display
	registry       *wl.Registry
	compositor     *wl.Compositor
	lockManager    *ext.SessionLockManager
	lock           *ext.SessionLock
	surfaces       map[*wl.Output]*ext.SessionLockSurface
	outputs        map[uint32]*wl.Output
	lockedReceived bool
	done           chan struct{}
}

// Create a new lock client
func NewLockClient() *LockClient {
	return &LockClient{
		surfaces: make(map[*wl.Output]*ext.SessionLockSurface),
		outputs:  make(map[uint32]*wl.Output),
		done:     make(chan struct{}),
	}
}

// Handle registry global events (discover available interfaces)
func (c *LockClient) HandleRegistryGlobal(ev wl.RegistryGlobalEvent) {
	if ev.Interface == "wl_compositor" {
		fmt.Printf("Found wl_compositor (name: %d)\n", ev.Name)
		c.compositor = wlclient.RegistryBindCompositorInterface(c.registry, ev.Name, 4)
		fmt.Printf("Bound compositor with ID: %d\n", c.compositor.Id())
	} else if ev.Interface == "ext_session_lock_manager_v1" && ev.Version >= 1 {
		fmt.Printf("Found ext_session_lock_manager_v1 (name: %d)\n", ev.Name)

		// Create the lock manager
		c.lockManager = ext.BindSessionLockManager(c.registry, ev.Name, 1)

		// Print the ID after creation
		fmt.Printf("Bound lock manager with ID: %d\n", c.lockManager.Id())
	} else if ev.Interface == "wl_output" {
		fmt.Printf("Found wl_output (name: %d)\n", ev.Name)

		// Use proper output binding
		output := wlclient.RegistryBindOutputInterface(c.registry, ev.Name, 3)
		fmt.Printf("Bound output with ID: %d\n", output.Id())
		c.outputs[ev.Name] = output
	}
}

// Handle registry global remove events
func (c *LockClient) HandleRegistryGlobalRemove(ev wl.RegistryGlobalRemoveEvent) {
	if output, exists := c.outputs[ev.Name]; exists {
		fmt.Printf("Output removed (name: %d)\n", ev.Name)

		// If we have a lock surface for this output, destroy it
		if lockSurface, ok := c.surfaces[output]; ok {
			lockSurface.Destroy()
			delete(c.surfaces, output)
		}

		delete(c.outputs, ev.Name)
	}
}

// Connect to Wayland display and discover interfaces
func (c *LockClient) Connect() error {
	var err error

	// Connect to Wayland display
	c.display, err = wlclient.DisplayConnect(nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Wayland display: %w", err)
	}

	// Get registry
	c.registry, err = c.display.GetRegistry()
	if err != nil {
		return fmt.Errorf("failed to get registry: %w", err)
	}

	// Add registry listener
	c.registry.AddGlobalHandler(c)
	c.registry.AddGlobalRemoveHandler(c)

	// Process events to discover interfaces
	if err := wlclient.DisplayRoundtrip(c.display); err != nil {
		return fmt.Errorf("failed roundtrip: %w", err)
	}

	if c.lockManager == nil {
		return fmt.Errorf("ext_session_lock_manager_v1 not available")
	}

	return nil
}

// Lock the session
func (c *LockClient) LockSession() error {
	// Log debugging info
	fmt.Println("Attempting to lock session...")

	if c.lockManager == nil {
		return fmt.Errorf("lock manager not initialized")
	}

	fmt.Printf("Lock manager ID: %d\n", c.lockManager.Id())

	// Add roundtrip to ensure previous operations are complete
	if err := wlclient.DisplayRoundtrip(c.display); err != nil {
		return fmt.Errorf("roundtrip failed before lock: %w", err)
	}

	// Try locking
	var err error
	c.lock, err = c.lockManager.Lock()
	if err != nil {
		return fmt.Errorf("failed to create session lock: %w", err)
	}

	// Check if lock was created successfully
	if c.lock != nil {
		fmt.Printf("Created lock with ID: %d\n", c.lock.Id())
	} else {
		fmt.Println("Lock was not created (nil)")
	}

	ext.SessionLockAddListener(c.lock, c)

	// Create lock surfaces for all outputs with better error handling
	for _, output := range c.outputs {
		if err := c.createLockSurface(output); err != nil {
			fmt.Printf("Warning: failed to create lock surface: %v\n", err)
		}
	}

	fmt.Println("Lock surfaces created. Waiting for compositor events...")
	return nil
}

// Create a lock surface for an output
func (c *LockClient) createLockSurface(output *wl.Output) error {
	// Create a wl_surface using wlclient
	surface, err := c.compositor.CreateSurface()
	if err != nil {
		return fmt.Errorf("failed to create surface: %v", err)
	}

	// Create lock surface
	lockSurface, err := c.lock.GetLockSurface(surface, output)
	if err != nil {
		return fmt.Errorf("failed to get lock surface: %v", err)
	}

	// Add lock surface listener
	ext.SessionLockSurfaceAddListener(lockSurface, &LockSurfaceHandler{
		client:    c,
		surface:   lockSurface,
		wlSurface: surface,
	})

	c.surfaces[output] = lockSurface
	return nil
}

// HandleSessionLockLocked implements SessionLockLockedHandler
func (c *LockClient) HandleSessionLockLocked(ev ext.SessionLockLockedEvent) {
	fmt.Println("Session is now locked!")
	c.lockedReceived = true
}

// HandleSessionLockFinished implements SessionLockFinishedHandler
func (c *LockClient) HandleSessionLockFinished(ev ext.SessionLockFinishedEvent) {
	fmt.Println("Lock manager finished the session lock")

	// Clean up based on whether we received locked event
	if c.lockedReceived {
		c.lock.UnlockAndDestroy()
	} else {
		c.lock.Destroy()
	}

	// Signal that we're done
	close(c.done)
}

// LockSurfaceHandler handles lock surface events
type LockSurfaceHandler struct {
	client    *LockClient
	surface   *ext.SessionLockSurface
	wlSurface *wl.Surface
	serial    uint32
	width     uint32
	height    uint32
}

// HandleSessionLockSurfaceConfigure implements SessionLockSurfaceConfigureHandler
func (h *LockSurfaceHandler) HandleSessionLockSurfaceConfigure(ev ext.SessionLockSurfaceConfigureEvent) {
	fmt.Printf("Configure: serial=%d, width=%d, height=%d\n", ev.Serial, ev.Width, ev.Height)

	// Save the serial and dimensions
	h.serial = ev.Serial
	h.width = ev.Width
	h.height = ev.Height

	// Acknowledge the configure event
	h.surface.AckConfigure(ev.Serial)

	// Create a solid color buffer
	createSolidColorBuffer(h.wlSurface, h.width, h.height, 64, 0, 0) // Dark red
}

// Main function
func main() {
	client := NewLockClient()

	// Connect to Wayland
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived signal, unlocking...")

		if client.lock != nil && client.lockedReceived {
			client.lock.UnlockAndDestroy()
		}
	}()

	// Lock the session
	if err := client.LockSession(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to lock session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session lock requested, waiting for events...")
	fmt.Println("Press Ctrl+C to unlock")

	// Main event loop
	go func() {
		for {
			// Wait a bit to avoid 100% CPU usage
			time.Sleep(100 * time.Millisecond)

			// Check if we're done
			select {
			case <-client.done:
				return
			default:
				// Dispatch Wayland events
				if err := wlclient.DisplayDispatch(client.display); err != nil {
					fmt.Fprintf(os.Stderr, "Error dispatching events: %v\n", err)
					return
				}
			}
		}
	}()

	// Wait for completion
	<-client.done
	fmt.Println("Session lock finished")
}

// Helper function to create a solid color buffer for a surface
func createSolidColorBuffer(surface *wl.Surface, width, height uint32, r, g, b uint8) {
	// In a real implementation, you would:
	// 1. Create a shared memory buffer
	// 2. Fill it with the solid color
	// 3. Attach it to the surface
	// 4. Commit the surface

	// This is a simplified version for the example
	fmt.Printf("Would create %dx%d buffer with color #%02x%02x%02x\n", width, height, r, g, b)
}
