// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"context"
	"image"
	"io"
	"log"
	"os"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// EventHandler handles keyboard and mouse events for tab navigation.
type EventHandler struct {
	term           terminalapi.Terminal // Terminal interface.
	tm             *Manager             // Reference to the Tab Manager.
	th             *Header              // Reference to the Header.
	tc             *Content             // Reference to the Content.
	container      *container.Container // Container for updating content.
	ctx            context.Context      // Context for cancellation.
	cancel         context.CancelFunc   // Function to cancel the context.
	opts           *Options             // Configuration options.
	logger         *log.Logger          // Logger for event handling.
	lastSwitchTime time.Time            // Tracks when the last tab switch due to notification occurred.
}

// NewEventHandler initializes a new EventHandler.
func NewEventHandler(ctx context.Context, term terminalapi.Terminal, tm *Manager, th *Header, tc *Content, cont *container.Container, cancel context.CancelFunc, opts *Options) *EventHandler {
	var logger *log.Logger

	// Only create the log file if EnableLogging is true.
	if opts.EnableLogging {
		file, err := os.OpenFile("tab_demo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("failed to open log file: %v", err)
			logger = log.New(io.Discard, "", 0) // Discard logging if file creation fails.
		} else {
			logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	} else {
		// Discard logs if logging is disabled, and do not create the file.
		logger = log.New(io.Discard, "", 0)
	}

	eh := &EventHandler{
		term:      term,
		tm:        tm,
		th:        th,
		tc:        tc,
		container: cont,
		ctx:       ctx,
		cancel:    cancel,
		opts:      opts,
		logger:    logger,
	}

	// Start the notification watcher if FollowNotifications is enabled.
	if opts.FollowNotifications {
		go eh.startNotificationWatcher()
	}

	return eh
}

// startNotificationWatcher monitors for tab notifications and switches tabs.
func (eh *EventHandler) startNotificationWatcher() {
	var notifiedTabsQueue []*Tab
	var currentTab *Tab
	var tabViewStartTime time.Time // Tracks how long the current tab has been viewed.

	for {
		select {
		case <-eh.ctx.Done():
			return
		default:
			currentTime := time.Now()

			// Get the list of current notifications.
			notifiedTabs := eh.tm.GetNotifiedTabs()

			// Update the queue with new notifications.
			for _, tab := range notifiedTabs {
				if !containsTab(notifiedTabsQueue, tab) {
					notifiedTabsQueue = append(notifiedTabsQueue, tab)
				}
			}

			// Remove expired notifications from the queue.
			for i := 0; i < len(notifiedTabsQueue); {
				if !notifiedTabsQueue[i].HasNotification() {
					notifiedTabsQueue = append(notifiedTabsQueue[:i], notifiedTabsQueue[i+1:]...)
				} else {
					i++
				}
			}

			// Ensure FollowNotifications is enabled and there's a tab to switch to.
			if eh.opts.FollowNotifications && len(notifiedTabsQueue) > 0 {
				if currentTab == nil || currentTime.Sub(eh.lastSwitchTime) >= 1*time.Second {
					// Switch to the next tab with a notification.
					currentTab = notifiedTabsQueue[0]
					index := eh.tm.GetTabIndex(currentTab)
					if index != -1 {
						eh.tm.SetActiveTab(index)
						if err := eh.th.Update(); err != nil {
							eh.logger.Printf("failed to update tab header: %v", err)
						}
						if err := eh.tc.Update(eh.container); err != nil {
							eh.logger.Printf("failed to update tab content: %v", err)
						}
						eh.lastSwitchTime = currentTime
						tabViewStartTime = currentTime                                              // Reset view start time.
						notifiedTabsQueue = append(notifiedTabsQueue[:0], notifiedTabsQueue[1:]...) // Remove first tab from the queue.
					}
				}
			}

			// Clear the notification if viewed for more than 1 second.
			if currentTab != nil && currentTime.Sub(tabViewStartTime) >= 1*time.Second {
				index := eh.tm.GetTabIndex(currentTab)
				if index != -1 {
					eh.tm.SetNotification(index, false, 0) // Clear the notification.
					if err := eh.th.Update(); err != nil {
						eh.logger.Printf("failed to update tab header: %v", err)
					}
				}
				currentTab = nil // Reset current tab after clearing notification.
			}

			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Helper function to check if a tab is in the queue.
func containsTab(tabs []*Tab, target *Tab) bool {
	for _, tab := range tabs {
		if tab == target {
			return true
		}
	}
	return false
}

// HandleKeyboard processes keyboard events for tab switching.
func (eh *EventHandler) HandleKeyboard(k *terminalapi.Keyboard) {
	eh.logger.Printf("Keyboard event: key=%v", k.Key)
	// Handle Ctrl+C and 'q' to exit.
	if k.Key == keyboard.KeyCtrlC || k.Key == keyboard.KeyEsc || k.Key == 'q' || k.Key == 'Q' {
		eh.cancel()
		return
	}

	// Switch to the next tab with Tab key or Right arrow key.
	if k.Key == keyboard.KeyTab || k.Key == keyboard.KeyArrowRight {
		eh.tm.NextTab()
		if err := eh.th.Update(); err != nil {
			eh.logger.Printf("failed to update tab header: %v", err)
		}
		if err := eh.tc.Update(eh.container); err != nil {
			eh.logger.Printf("failed to update tab content: %v", err)
		}
		return
	}

	// Switch to the previous tab with Left arrow key.
	if k.Key == keyboard.KeyArrowLeft {
		eh.tm.PreviousTab()
		if err := eh.th.Update(); err != nil {
			eh.logger.Printf("failed to update tab header: %v", err)
		}
		if err := eh.tc.Update(eh.container); err != nil {
			eh.logger.Printf("failed to update tab content: %v", err)
		}
		return
	}
}

// HandleMouse processes mouse events for tab switching.
func (eh *EventHandler) HandleMouse(m *terminalapi.Mouse) {
	eh.logger.Printf("Mouse event: button=%v, position=%v", m.Button, m.Position)

	if m.Button != mouse.ButtonLeft {
		return // Only handle left-click presses.
	}

	// Get terminal size.
	size := eh.term.Size()
	height := size.Y

	// Calculate the height of the Header.
	headerHeight := eh.th.Height()
	if headerHeight == 0 {
		headerHeight = int(float64(height) * 0.1) // Fallback to 10% of terminal height.
		if headerHeight == 0 {
			headerHeight = 1
		}
	}

	// Check if the click is within the Header area.
	if m.Position.Y >= headerHeight {
		// Click is outside the Header.
		return
	}

	// Adjust the mouse position to be relative to the Header.
	adjustedPosition := image.Point{
		X: m.Position.X,
		Y: m.Position.Y, // Since Header starts at Y=0, this remains the same.
	}

	// Determine which tab was clicked.
	clickedTabIndex := eh.th.GetClickedTab(adjustedPosition)
	if clickedTabIndex != -1 {
		eh.tm.SetActiveTab(clickedTabIndex)
		if err := eh.th.Update(); err != nil {
			eh.logger.Printf("failed to update tab header: %v", err)
		}
		if err := eh.tc.Update(eh.container); err != nil {
			eh.logger.Printf("failed to update tab content: %v", err)
		}
	}
}
