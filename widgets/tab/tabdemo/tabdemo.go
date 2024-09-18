// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/tab"
	"github.com/mum4k/termdash/widgets/text"
)

func main() {
	// Initialize Terminal using tcell
	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to initialize terminal: %v", err)
	}
	defer term.Close()

	// Create instructions text widget
	instructionsText, err := text.New()
	if err != nil {
		log.Fatalf("failed to create instructions text widget: %v", err)
	}
	instructionsText.Write("Use Tab, Left/Right Arrow keys to navigate tabs. Press 'q', Esc, or Ctrl+C to exit.")

	// Create PID Text Widget once with built-in rolling
	pidText, err := text.New(
		text.WrapAtWords(),
		text.RollContent(),
	)
	if err != nil {
		log.Fatalf("failed to create PID text widget: %v", err)
	}

	// Create TabManager
	tabManager := tab.NewTabManager()

	// Create options
	opts := tab.NewOptions(
		tab.ActiveIcon("⦿"),                   // Custom active icon
		tab.InactiveIcon("○"),                 // Custom inactive icon
		tab.NotificationIcon("⚠"),             // Custom notification icon
		tab.EnableLogging(false),              // Enable logging
		tab.LabelColor(cell.ColorYellow),      // Custom label color
		tab.ActiveTabColor(cell.ColorBlue),    // Active tab background color
		tab.InactiveTabColor(cell.ColorBlack), // Inactive tab background color
		tab.FollowNotifications(true),         // Enable following notifications
	)

	// Create Widgets
	// CPU Widgets
	cpuGauge, err := gauge.New(
		gauge.Color(cell.ColorGreen),
	)
	if err != nil {
		log.Fatalf("failed to create CPU gauge: %v", err)
	}
	cpuText, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create CPU text widget: %v", err)
	}

	// GPU Widgets
	gpuLineChart, err := linechart.New(
		linechart.YAxisFormattedValues(func(v float64) string {
			return fmt.Sprintf("%.0f%%", v)
		}),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.XAxisUnscaled(),
	)
	if err != nil {
		log.Fatalf("failed to create GPU line chart: %v", err)
	}
	gpuDonut, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorBlue)),
	)
	if err != nil {
		log.Fatalf("failed to create GPU donut: %v", err)
	}

	// Memory Widgets
	memLineChart, err := linechart.New(
		linechart.YAxisFormattedValues(func(v float64) string {
			return fmt.Sprintf("%.0f%%", v)
		}),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.XAxisUnscaled(),
	)
	if err != nil {
		log.Fatalf("failed to create Memory line chart: %v", err)
	}
	memText, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create Memory text widget: %v", err)
	}

	// Storage Widgets
	storageLineChart, err := linechart.New(
		linechart.YAxisFormattedValues(func(v float64) string {
			return fmt.Sprintf("%.0f%%", v)
		}),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.XAxisUnscaled(),
	)
	if err != nil {
		log.Fatalf("failed to create Storage line chart: %v", err)
	}
	storageText, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create Storage text widget: %v", err)
	}

	// Network Widgets
	networkLineChart, err := linechart.New(
		linechart.YAxisFormattedValues(func(v float64) string {
			return fmt.Sprintf("%.0f Mbps", v)
		}),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.XAxisUnscaled(),
	)
	if err != nil {
		log.Fatalf("failed to create Network line chart: %v", err)
	}
	networkText, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create Network text widget: %v", err)
	}

	// Create Tabs

	// Tab 1: CPU
	tabCPU := &tab.Tab{
		Name: "CPU",
		Content: container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("CPU Usage"),
				container.PlaceWidget(cpuGauge),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("CPU Details"),
						container.PlaceWidget(cpuText),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Running Processes"),
						container.PlaceWidget(pidText),
					),
					container.SplitPercent(50),
				),
			),
			container.SplitPercent(70),
		),
	}
	tabManager.AddTab(tabCPU)

	// Tab 2: GPU
	gpuDetailsText, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create GPU details text widget: %v", err)
	}
	gpuDetailsText.Write("GPU Details")

	tabGPU := &tab.Tab{
		Name: "GPU",
		Content: container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("GPU Usage"),
						container.PlaceWidget(gpuDonut),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("GPU Activity"),
						container.PlaceWidget(gpuLineChart),
					),
					container.SplitPercent(50),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Details"),
				container.PlaceWidget(gpuDetailsText),
			),
			container.SplitPercent(70),
		),
	}
	tabManager.AddTab(tabGPU)

	// Tab 3: Memory
	tabMemory := &tab.Tab{
		Name: "Memory",
		Content: container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Memory Usage"),
				container.PlaceWidget(memLineChart),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Details"),
				container.PlaceWidget(memText),
			),
			container.SplitPercent(70),
		),
	}
	tabManager.AddTab(tabMemory)

	// Tab 4: Storage
	tabStorage := &tab.Tab{
		Name: "Storage",
		Content: container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Storage Activity"),
				container.PlaceWidget(storageLineChart),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Details"),
				container.PlaceWidget(storageText),
			),
			container.SplitPercent(70),
		),
	}
	tabManager.AddTab(tabStorage)

	// Tab 5: Network
	tabNetwork := &tab.Tab{
		Name: "Network",
		Content: container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Network Traffic"),
				container.PlaceWidget(networkLineChart),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Details"),
				container.PlaceWidget(networkText),
			),
			container.SplitPercent(70),
		),
	}
	tabManager.AddTab(tabNetwork)

	// Create TabHeader
	tabHeader, err := tab.NewTabHeader(tabManager, opts)
	if err != nil {
		log.Fatalf("failed to create tab header: %v", err)
	}

	// Create initial content widget
	initialContent, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create initial content widget: %v", err)
	}
	initialContent.Write("Select a tab to view its content.")

	// Create main container with tab header, tab content, and instructions
	cont, err := container.New(
		term,
		container.Border(linestyle.Light),
		container.BorderTitle("System Monitor"),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(tabHeader.Widget()),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(initialContent),
						container.ID("tabContent"),
					),
					container.Bottom(
						container.PlaceWidget(instructionsText),
					),
					container.SplitPercent(90), // 90% content, 10% instructions
				),
			),
			container.SplitPercent(10), // 10% header, 90% rest
		),
	)
	if err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	// Create TabContent and set the active tab's content
	tabContent := tab.NewTabContent(tabManager)
	err = tabContent.Update(cont)
	if err != nil {
		log.Fatalf("failed to update tab content: %v", err)
	}

	// Create context for application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize TabEventHandler for keyboard and mouse events
	tabEventHandler := tab.NewTabEventHandler(
		ctx, // Pass the context here
		term,
		tabManager,
		tabHeader,
		tabContent,
		cont,
		cancel,
		opts,
	)

	// Start a goroutine to simulate data updates
	go func() {
		rand.Seed(time.Now().UnixNano())

		// Data buffers for smooth scrolling
		gpuData := make([]float64, 0)
		gpuTimestamps := make([]string, 0)

		memData := make([]float64, 0)
		memTimestamps := make([]string, 0)

		storageData := make([]float64, 0)
		storageTimestamps := make([]string, 0)

		netData := make([]float64, 0)
		netTimestamps := make([]string, 0)

		// PID lines buffer
		pidLines := make([]string, 0)
		const maxPidLines = 20

		const maxDataPoints = 50

		// Application names for simulation
		appNames := []string{"Chrome", "VSCode", "Terminal", "Slack", "Spotify", "Docker", "Mail", "Zoom", "Notepad"}

		// Simulate notification alarms
		notificationTicker := time.NewTicker(10 * time.Second)
		defer notificationTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				currentTime := time.Now().Format("15:04:05")

				// Simulate CPU usage
				cpuUsage := rand.Float64() * 100
				cpuGauge.Percent(int(cpuUsage), gauge.TextLabel(fmt.Sprintf("%.2f%%", cpuUsage)))

				cpuText.Reset()
				cpuText.Write(fmt.Sprintf("CPU Usage: %.2f%%", cpuUsage))

				// Simulate PIDs and application names
				pid := rand.Intn(10000)
				appName := appNames[rand.Intn(len(appNames))]
				line := fmt.Sprintf("PID %d: %s", pid, appName)

				// Append to pidLines and trim if necessary
				pidLines = append(pidLines, line)
				if len(pidLines) > maxPidLines {
					pidLines = pidLines[1:]
				}

				// Update the pidText widget
				pidText.Reset()
				pidText.Write(strings.Join(pidLines, "\n"))

				// Simulate GPU usage
				gpuUsage := rand.Float64() * 100
				gpuDonut.Percent(int(gpuUsage), donut.Label(fmt.Sprintf("%.2f%%", gpuUsage)))

				if len(gpuData) >= maxDataPoints {
					gpuData = gpuData[1:]
					gpuTimestamps = gpuTimestamps[1:]
				}
				gpuData = append(gpuData, gpuUsage)
				gpuTimestamps = append(gpuTimestamps, currentTime)

				// Create xLabels map for GPU
				gpuXLabels := make(map[int]string)
				for idx, label := range gpuTimestamps {
					gpuXLabels[idx] = label
				}

				if err := gpuLineChart.Series("GPU", gpuData,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
					linechart.SeriesXLabels(gpuXLabels),
				); err != nil {
					log.Printf("failed to update GPU line chart: %v", err)
				}

				// Simulate Memory usage
				memUsage := rand.Float64() * 100
				if len(memData) >= maxDataPoints {
					memData = memData[1:]
					memTimestamps = memTimestamps[1:]
				}
				memData = append(memData, memUsage)
				memTimestamps = append(memTimestamps, currentTime)

				// Create xLabels map for Memory
				memXLabels := make(map[int]string)
				for idx, label := range memTimestamps {
					memXLabels[idx] = label
				}

				if err := memLineChart.Series("Memory", memData,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorGreen)),
					linechart.SeriesXLabels(memXLabels),
				); err != nil {
					log.Printf("failed to update Memory line chart: %v", err)
				}
				memText.Reset()
				memText.Write(fmt.Sprintf("Memory Usage: %.2f%%", memUsage))

				// Simulate Storage activity
				storageActivity := rand.Float64() * 100
				if len(storageData) >= maxDataPoints {
					storageData = storageData[1:]
					storageTimestamps = storageTimestamps[1:]
				}
				storageData = append(storageData, storageActivity)
				storageTimestamps = append(storageTimestamps, currentTime)

				// Create xLabels map for Storage
				storageXLabels := make(map[int]string)
				for idx, label := range storageTimestamps {
					storageXLabels[idx] = label
				}

				if err := storageLineChart.Series("Storage", storageData,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorCyan)),
					linechart.SeriesXLabels(storageXLabels),
				); err != nil {
					log.Printf("failed to update Storage line chart: %v", err)
				}
				// Update storage details
				storageText.Reset()
				driveNames := []string{"Drive A", "Drive B", "Drive C"}
				driveCapacities := []float64{500, 1000, 2000} // in GB
				for i, name := range driveNames {
					used := rand.Float64() * driveCapacities[i]
					percentUsed := (used / driveCapacities[i]) * 100
					storageText.Write(fmt.Sprintf("%s: Used %.2fGB / %.0fGB (%.2f%%)\n", name, used, driveCapacities[i], percentUsed))
				}

				// Simulate Network traffic
				networkTraffic := rand.Float64() * 1000 // Mbps
				if len(netData) >= maxDataPoints {
					netData = netData[1:]
					netTimestamps = netTimestamps[1:]
				}
				netData = append(netData, networkTraffic)
				netTimestamps = append(netTimestamps, currentTime)

				// Create xLabels map for Network
				netXLabels := make(map[int]string)
				for idx, label := range netTimestamps {
					netXLabels[idx] = label
				}

				if err := networkLineChart.Series("Network", netData,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorMagenta)),
					linechart.SeriesXLabels(netXLabels),
				); err != nil {
					log.Printf("failed to update Network line chart: %v", err)
				}
				networkText.Reset()
				networkText.Write(fmt.Sprintf("Network Traffic: %.2f Mbps", networkTraffic))

				// Sleep before next update
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	// Start a goroutine to periodically trigger tab notifications
	go func() {
		rand.Seed(time.Now().UnixNano())
		ticker := time.NewTicker(10 * time.Second) // Every 10 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				numTabs := tabManager.GetTabNum()
				if numTabs == 0 {
					continue
				}
				randomIndex := rand.Intn(numTabs)
				// Set notification to true
				tabManager.SetNotification(randomIndex, true, 5*time.Second)
				if err := tabHeader.Update(); err != nil {
					log.Printf("failed to update tab header: %v", err)
				}
			}
		}
	}()

	// Run Termdash with both exit and tab navigation handlers
	err = termdash.Run(ctx, term, cont,
		termdash.KeyboardSubscriber(tabEventHandler.HandleKeyboard),
		termdash.MouseSubscriber(tabEventHandler.HandleMouse), // Subscribe to mouse events
	)
	if err != nil && err != context.Canceled {
		log.Fatalf("termdash encountered an error: %v", err)
	}
}
