// treeview.go
package treeview

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

const (
	ScrollStep = 5 // Number of nodes to scroll per mouse wheel event
)

// TreeNode represents a node in the treeview.
type TreeNode struct {
	ID            string
	Label         string
	Level         int
	Parent        *TreeNode
	Children      []*TreeNode
	Value         interface{} // Can hold any data type
	ShowSpinner   bool
	OnClick       func() error
	ExpandedState bool // Unique expanded state for each node
	SpinnerIndex  int  // Current index for spinner icons
	mu            sync.Mutex
}

// SetShowSpinner safely sets the ShowSpinner flag.
func (node *TreeNode) SetShowSpinner(value bool) {
	node.mu.Lock()
	node.mu.Unlock()
	node.ShowSpinner = value
	if !value {
		node.SpinnerIndex = 0 // Reset spinner index when spinner is turned off
	}
}

// GetShowSpinner safely retrieves the ShowSpinner flag.
func (node *TreeNode) GetShowSpinner() bool {
	node.mu.Lock()
	defer node.mu.Unlock()
	return node.ShowSpinner
}

// IncrementSpinner safely increments the SpinnerIndex.
func (node *TreeNode) IncrementSpinner(totalIcons int) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.SpinnerIndex = (node.SpinnerIndex + 1) % totalIcons
}

// IsRoot checks if the node is a root node.
func (node *TreeNode) IsRoot() bool {
	return node.Parent == nil
}

// Treeview represents the treeview widget.
type Treeview struct {
	mu                  sync.Mutex
	position            image.Point // Stores the widget's top-left position
	opts                *options
	selectedNode        *TreeNode
	visibleNodes        []*TreeNode
	logger              *log.Logger
	spinnerTicker       *time.Ticker
	stopSpinner         chan struct{}
	expandedIcon        string
	collapsedIcon       string
	leafIcon            string
	scrollOffset        int
	indentationPerLevel int
	canvasWidth         int
	canvasHeight        int
	totalContentHeight  int
	waitingIcons        []string
	lastClickTime       time.Time // Timestamp of the last handled click
	lastKeyTime         time.Time // Timestamp for debouncing the enter key
}

// New creates a new Treeview instance.
func New(opts ...Option) (*Treeview, error) {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Set default leaf icon if not provided
	if options.leafIcon == "" {
		options.leafIcon = "→"
	}

	// Set default indentation if not provided
	if options.indentation == 0 {
		options.indentation = 2
	}

	for _, node := range options.nodes {
		setParentsAndAssignIDs(node, nil, 0, "")
	}

	// Create a logger to log debugging information to a file if logging is enabled
	var logger *log.Logger
	if options.enableLogging {
		file, err := os.OpenFile("treeview_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// Create a dummy logger that discards all logs
		logger = log.New(io.Discard, "", 0)
	}

	tv := &Treeview{
		opts:                options,
		logger:              logger,
		stopSpinner:         make(chan struct{}),
		expandedIcon:        options.expandedIcon,
		collapsedIcon:       options.collapsedIcon,
		leafIcon:            options.leafIcon,
		scrollOffset:        0,
		indentationPerLevel: options.indentation,
		waitingIcons:        options.waitingIcons,
	}

	setInitialExpandedState(tv, true) // Expand root nodes by default

	if len(options.waitingIcons) > 0 {
		tv.spinnerTicker = time.NewTicker(200 * time.Millisecond)
		go tv.runSpinner()
	}
	tv.updateTotalHeight()

	// Set selectedNode to the first visible node
	visibleNodes := tv.getVisibleNodesList()
	if len(visibleNodes) > 0 {
		tv.selectedNode = visibleNodes[0]
	}
	return tv, nil
}

// generateNodeID creates a consistent node ID.
func generateNodeID(path string, label string) string {
	if path == "" {
		return label
	}
	return fmt.Sprintf("%s/%s", path, label)
}

// setParentsAndAssignIDs assigns parent references, levels, and IDs to nodes recursively.
func setParentsAndAssignIDs(node *TreeNode, parent *TreeNode, level int, path string) {
	node.Parent = parent
	node.Level = level

	node.ID = generateNodeID(path, node.Label)

	for _, child := range node.Children {
		setParentsAndAssignIDs(child, node, level+1, node.ID)
	}
}

// runSpinner updates spinner indices periodically.
func (tv *Treeview) runSpinner() {
	for {
		select {
		case <-tv.spinnerTicker.C:
			tv.mu.Lock()
			visibleNodes := tv.getVisibleNodesList()
			for _, node := range visibleNodes {
				node.mu.Lock()
				if node.GetShowSpinner() && len(tv.waitingIcons) > 0 {
					node.IncrementSpinner(len(tv.waitingIcons))
					tv.logger.Printf("Spinner updated for node: %s (SpinnerIndex: %d)", node.Label, node.SpinnerIndex)
				}
				node.mu.Unlock()
			}
			tv.mu.Unlock()
		case <-tv.stopSpinner:
			return
		}
	}
}

// StopSpinnerTicker stops the spinner ticker.
func (tv *Treeview) StopSpinnerTicker() {
	if tv.spinnerTicker != nil {
		tv.spinnerTicker.Stop()
		close(tv.stopSpinner)
	}
}

// setInitialExpandedState sets the initial expanded state for root nodes.
func setInitialExpandedState(tv *Treeview, expandRoot bool) {
	for _, node := range tv.opts.nodes {
		if node.IsRoot() {
			node.SetExpandedState(expandRoot)
		}
	}
	tv.updateTotalHeight()
}

// calculateHeight calculates the height of a node, including its children if expanded.
func (tv *Treeview) calculateHeight(node *TreeNode) int {
	height := 1 // Start with the height of the current node
	if node.ExpandedState {
		for _, child := range node.Children {
			height += tv.calculateHeight(child)
		}
	}
	return height
}

// calculateTotalHeight calculates the total height of all visible nodes.
func (tv *Treeview) calculateTotalHeight() int {
	totalHeight := 0
	for _, rootNode := range tv.opts.nodes {
		totalHeight += tv.calculateHeight(rootNode)
	}
	return totalHeight
}

// updateTotalHeight updates the totalContentHeight based on visible nodes.
func (tv *Treeview) updateTotalHeight() {
	tv.totalContentHeight = tv.calculateTotalHeight()
}

// getVisibleNodesList retrieves a flat list of all currently visible nodes.
func (tv *Treeview) getVisibleNodesList() []*TreeNode {
	var list []*TreeNode
	var traverse func(node *TreeNode)
	traverse = func(node *TreeNode) {
		list = append(list, node)
		tv.logger.Printf("Visible Node Added: '%s' at Level %d", node.Label, node.Level)
		if node.GetExpandedState() { // Use getter with mutex
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}
	for _, root := range tv.opts.nodes {
		traverse(root)
	}
	return list
}

// getNodePrefix returns the appropriate prefix for a node based on its state.
func (tv *Treeview) getNodePrefix(node *TreeNode) string {
	if node.GetShowSpinner() && len(tv.waitingIcons) > 0 {
		return tv.waitingIcons[node.SpinnerIndex]
	} else if len(node.Children) > 0 {
		if node.ExpandedState {
			return tv.expandedIcon
		} else {
			return tv.collapsedIcon
		}
	} else {
		return tv.leafIcon
	}
}

// drawNode draws nodes based on the nodesToDraw slice.
func (tv *Treeview) drawNode(cvs *canvas.Canvas, nodesToDraw []*TreeNode) error {
	for y, node := range nodesToDraw {
		// Determine if this node is selected
		isSelected := (node.ID == tv.selectedNode.ID)

		// Get the prefix based on node state
		prefix := tv.getNodePrefix(node)
		prefixWidth := runewidth.StringWidth(prefix)

		// Construct the label
		label := fmt.Sprintf("%s %s", prefix, node.Label)
		labelWidth := runewidth.StringWidth(label)
		indentX := node.Level * tv.indentationPerLevel
		availableWidth := tv.canvasWidth - indentX

		if tv.opts.truncate && labelWidth > availableWidth {
			// Truncate the label to fit within the available space
			truncatedLabel := truncateString(label, availableWidth)
			if truncatedLabel != label {
				label = truncatedLabel
			}
			labelWidth = runewidth.StringWidth(label)
		}

		// Log prefix width for debugging
		tv.logger.Printf("Drawing node '%s' with prefix width %d", node.Label, prefixWidth)

		// Determine colors based on selection
		var fgColor cell.Color = tv.opts.labelColor
		var bgColor cell.Color = cell.ColorDefault
		if isSelected {
			fgColor = cell.ColorBlack
			bgColor = cell.ColorWhite
		}

		// Draw the label at the correct position
		if err := tv.drawLabel(cvs, label, indentX, y, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// findNodeByClick determines which node was clicked based on x and y coordinates.
func (tv *Treeview) findNodeByClick(x, y int, visibleNodes []*TreeNode) *TreeNode {
	clickedIndex := y + tv.scrollOffset // Adjust Y-coordinate based on scroll offset
	if clickedIndex < 0 || clickedIndex >= len(visibleNodes) {
		return nil
	}

	node := visibleNodes[clickedIndex]

	label := fmt.Sprintf("%s %s", tv.getNodePrefix(node), node.Label)
	labelWidth := runewidth.StringWidth(label)
	indentX := node.Level * tv.indentationPerLevel
	availableWidth := tv.canvasWidth - indentX

	if tv.opts.truncate && labelWidth > availableWidth {
		truncatedLabel := truncateString(label, availableWidth)
		labelWidth = runewidth.StringWidth(truncatedLabel)
		label = truncatedLabel
	}

	labelStartX := indentX
	labelEndX := labelStartX + labelWidth

	if x >= labelStartX && x < labelEndX {
		tv.logger.Printf("Node '%s' (ID: %s) clicked at [X:%d Y:%d]", node.Label, node.ID, x, y)
		return node
	}

	return nil
}

// handleMouseClick processes mouse click at given x, y coordinates.
func (tv *Treeview) handleMouseClick(x, y int) error {
	tv.logger.Printf("Handling mouse click at (X:%d, Y:%d)", x, y)
	visibleNodes := tv.visibleNodes
	clickedNode := tv.findNodeByClick(x, y, visibleNodes)
	if clickedNode != nil {
		tv.logger.Printf("Node: %s (ID: %s) clicked, expanded: %v", clickedNode.Label, clickedNode.ID, clickedNode.ExpandedState)
		// Update selectedNode to the clicked node
		tv.selectedNode = clickedNode
		if err := tv.handleNodeClick(clickedNode); err != nil {
			tv.logger.Println("Error handling node click:", err)
		}
	} else {
		tv.logger.Printf("No node found at position: (X:%d, Y:%d)", x, y)
	}

	return nil
}

// handleNodeClick toggles the expansion state of a node and manages the spinner.
func (tv *Treeview) handleNodeClick(node *TreeNode) error {
	// Lock the Treeview before modifying shared fields
	tv.mu.Lock()
	defer tv.mu.Unlock()
	tv.logger.Printf("Handling node click for: %s (ID: %s)", node.Label, node.ID)
	if len(node.Children) > 0 {
		// Toggle expansion state
		node.SetExpandedState(!node.GetExpandedState())
		tv.updateTotalHeight()
		tv.logger.Printf("Toggled expansion for node: %s to %v", node.Label, node.ExpandedState)
		return nil
	}

	// Handle leaf node click
	if node.OnClick != nil {
		node.SetShowSpinner(true)
		tv.logger.Printf("Spinner started for node: %s", node.Label)
		go func(n *TreeNode) {
			tv.logger.Printf("Executing OnClick for node: %s", n.Label)
			if err := n.OnClick(); err != nil {
				tv.logger.Printf("Error executing OnClick for node %s: %v", n.Label, err)
			}
			n.SetShowSpinner(false)
			tv.logger.Printf("Spinner stopped for node: %s", n.Label)
		}(node)
	}

	return nil
}

// Mouse handles mouse events with debouncing for ButtonLeft clicks.
// It processes mouse press events and mouse wheel events.
func (tv *Treeview) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	// Ignore mouse release events to avoid handling multiple events per physical click
	if m.Button == mouse.ButtonRelease {
		return nil
	}

	// Adjust coordinates to be relative to the widget's position
	tv.mu.Lock()
	x := m.Position.X - tv.position.X
	y := m.Position.Y - tv.position.Y
	tv.mu.Unlock()

	switch m.Button {
	case mouse.ButtonLeft:
		tv.mu.Lock()
		now := time.Now()
		if now.Sub(tv.lastClickTime) < 100*time.Millisecond {
			// Ignore duplicate click within 100ms
			tv.logger.Printf("Ignored duplicate ButtonLeft click at (X:%d, Y:%d)", x, y)
			tv.mu.Unlock()
			return nil
		}
		tv.lastClickTime = now
		tv.mu.Unlock()
		tv.logger.Printf("MouseDown event at position: (X:%d, Y:%d)", x, y)
		return tv.handleMouseClick(x, y)
	case mouse.ButtonWheelUp:
		tv.mu.Lock()
		defer tv.mu.Unlock()
		tv.logger.Println("Mouse wheel up")
		if tv.scrollOffset >= ScrollStep {
			tv.scrollOffset -= ScrollStep
		} else {
			tv.scrollOffset = 0
		}
		tv.updateVisibleNodes()
		return nil
	case mouse.ButtonWheelDown:
		tv.mu.Lock()
		defer tv.mu.Unlock()
		tv.logger.Println("Mouse wheel down")
		maxOffset := tv.totalContentHeight - tv.canvasHeight
		if maxOffset < 0 {
			maxOffset = 0
		}
		if tv.scrollOffset+ScrollStep <= maxOffset {
			tv.scrollOffset += ScrollStep
		} else {
			tv.scrollOffset = maxOffset
		}
		tv.updateVisibleNodes()
		return nil
	}
	return nil
}

// Keyboard handles keyboard events.
func (tv *Treeview) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)

	if currentIndex == -1 {
		if len(visibleNodes) > 0 {
			tv.selectedNode = visibleNodes[0]
			currentIndex = 0
		} else {
			// No visible nodes to select
			return nil
		}
	}

	// Debounce Enter key to avoid rapid toggling
	now := time.Now()
	if k.Key == keyboard.KeyEnter || k.Key == ' ' {
		if now.Sub(tv.lastKeyTime) < 100*time.Millisecond {
			tv.logger.Printf("Ignored rapid Enter key press")
			return nil
		}
		tv.lastKeyTime = now
	}

	switch k.Key {
	case keyboard.KeyArrowDown:
		if currentIndex < len(visibleNodes)-1 {
			currentIndex++
			tv.selectedNode = visibleNodes[currentIndex]
			// Adjust scrollOffset to keep selectedNode in view
			if currentIndex >= tv.scrollOffset+tv.canvasHeight {
				tv.scrollOffset = currentIndex - tv.canvasHeight + 1
			}
		}
	case keyboard.KeyArrowUp:
		if currentIndex > 0 {
			currentIndex--
			tv.selectedNode = visibleNodes[currentIndex]
			// Adjust scrollOffset to keep selectedNode in view
			if currentIndex < tv.scrollOffset {
				tv.scrollOffset = currentIndex
			}
		}
	case keyboard.KeyEnter, ' ':
		if currentIndex >= 0 && currentIndex < len(visibleNodes) {
			node := visibleNodes[currentIndex]
			tv.selectedNode = node
			if err := tv.handleNodeClick(node); err != nil {
				tv.logger.Println("Error handling node click:", err)
			}
		}
	default:
		// Handle other keys if needed
	}

	return nil
}

// SetExpandedState safely sets the ExpandedState flag.
func (node *TreeNode) SetExpandedState(value bool) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.ExpandedState = value
}

// GetExpandedState safely retrieves the ExpandedState flag.
func (node *TreeNode) GetExpandedState() bool {
	node.mu.Lock()
	defer node.mu.Unlock()
	return node.ExpandedState
}

// getSelectedNodeIndex returns the index of the selected node in the visibleNodes list.
func (tv *Treeview) getSelectedNodeIndex(visibleNodes []*TreeNode) int {
	for idx, node := range visibleNodes {
		if node.ID == tv.selectedNode.ID {
			return idx
		}
	}
	return -1
}

// drawScrollUp draws the scroll up indicator.
func (tv *Treeview) drawScrollUp(cvs *canvas.Canvas) error {
	if _, err := cvs.SetCell(image.Point{X: 0, Y: 0}, '↑', cell.FgColor(cell.ColorWhite)); err != nil {
		return err
	}
	return nil
}

// drawScrollDown draws the scroll down indicator.
func (tv *Treeview) drawScrollDown(cvs *canvas.Canvas) error {
	if _, err := cvs.SetCell(image.Point{X: 0, Y: cvs.Area().Dy() - 1}, '↓', cell.FgColor(cell.ColorWhite)); err != nil {
		return err
	}
	return nil
}

// drawLabel draws the label of a node at the specified position with given foreground and background colors.
func (tv *Treeview) drawLabel(cvs *canvas.Canvas, label string, x, y int, fgColor, bgColor cell.Color) error {
	tv.logger.Printf("Drawing label: '%s' at X: %d, Y: %d with FG: %v, BG: %v", label, x, y, fgColor, bgColor)
	displayWidth := runewidth.StringWidth(label)
	if x+displayWidth > cvs.Area().Dx() {
		displayWidth = cvs.Area().Dx() - x
	}

	truncatedLabel := truncateString(label, displayWidth)

	for i, r := range truncatedLabel {
		if x+i >= cvs.Area().Dx() || y >= cvs.Area().Dy() {
			// If the x or y position exceeds the canvas dimensions, stop drawing
			break
		}
		if _, err := cvs.SetCell(image.Point{X: x + i, Y: y}, r, cell.FgColor(fgColor), cell.BgColor(bgColor)); err != nil {
			return err
		}
	}
	return nil
}

// Draw renders the treeview widget.
func (tv *Treeview) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	tv.updateVisibleNodes()

	visibleNodes := tv.visibleNodes
	totalHeight := len(visibleNodes)
	width := cvs.Area().Dx()
	tv.canvasWidth = width // Set canvasWidth here
	tv.canvasHeight = cvs.Area().Dy()

	// Log canvas dimensions
	tv.logger.Printf("Canvas Area: Dx=%d, Dy=%d", tv.canvasWidth, tv.canvasHeight)

	if tv.canvasWidth <= 0 || tv.canvasHeight <= 0 {
		return fmt.Errorf("canvas too small")
	}

	// Calculate the maximum valid scroll offset
	maxScrollOffset := tv.totalContentHeight - tv.canvasHeight
	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}

	// Clamp scrollOffset to ensure it stays within valid bounds
	if tv.scrollOffset > maxScrollOffset {
		tv.scrollOffset = maxScrollOffset
		tv.logger.Printf("Clamped scrollOffset to maxScrollOffset: %d", tv.scrollOffset)
	}
	if tv.scrollOffset < 0 {
		tv.scrollOffset = 0
		tv.logger.Printf("Clamped scrollOffset to 0")
	}

	tv.logger.Printf("Starting Draw with scrollOffset: %d, totalHeight: %d, canvasHeight: %d", tv.scrollOffset, totalHeight, tv.canvasHeight)

	// Clear the canvas
	if err := cvs.Clear(); err != nil {
		return err
	}

	// Determine the range of nodes to draw
	start := tv.scrollOffset
	end := tv.scrollOffset + tv.canvasHeight
	if end > len(visibleNodes) {
		end = len(visibleNodes)
	}

	// Slice the visibleNodes to only the range to draw
	nodesToDraw := visibleNodes[start:end]

	// Draw nodes
	if err := tv.drawNode(cvs, nodesToDraw); err != nil {
		tv.logger.Printf("Error drawing nodes: %v", err)
		return err
	}

	// Draw scroll indicators if needed
	if tv.scrollOffset > 0 {
		if err := tv.drawScrollUp(cvs); err != nil {
			tv.logger.Printf("Error drawing scroll up indicator: %v", err)
			return err
		}
	}
	if tv.scrollOffset+tv.canvasHeight < totalHeight {
		if err := tv.drawScrollDown(cvs); err != nil {
			tv.logger.Printf("Error drawing scroll down indicator: %v", err)
			return err
		}
	}

	tv.logger.Printf("Finished Draw, final currentY: %d, scrollOffset: %d", end, tv.scrollOffset)
	return nil
}

// Options returns the widget options to satisfy the widgetapi.Widget interface.
func (tv *Treeview) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:              image.Point{10, 3},
		WantKeyboard:             widgetapi.KeyScopeFocused,
		WantMouse:                widgetapi.MouseScopeWidget,
		ExclusiveKeyboardOnFocus: true,
	}
}

// Select returns the label of the selected node.
func (tv *Treeview) Select() (string, error) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	if tv.selectedNode != nil {
		return tv.selectedNode.Label, nil
	}
	return "", errors.New("no option selected")
}

// Next moves the selection down.
func (tv *Treeview) Next() {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)
	if currentIndex >= 0 && currentIndex < len(visibleNodes)-1 {
		currentIndex++
		tv.selectedNode = visibleNodes[currentIndex]
		// Adjust scrollOffset to keep selectedNode in view
		if currentIndex >= tv.scrollOffset+tv.canvasHeight {
			tv.scrollOffset = currentIndex - tv.canvasHeight + 1
		}
	}
}

// Previous moves the selection up.
func (tv *Treeview) Previous() {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)
	if currentIndex > 0 {
		currentIndex--
		tv.selectedNode = visibleNodes[currentIndex]
		// Adjust scrollOffset to keep selectedNode in view
		if currentIndex < tv.scrollOffset {
			tv.scrollOffset = currentIndex
		}
	}
}

// findNearestVisibleNode finds the nearest visible node in the tree
func (tv *Treeview) findNearestVisibleNode(currentNode *TreeNode, visibleNodes []*TreeNode) *TreeNode {
	if currentNode == nil {
		return nil
	}

	if currentNode.Parent != nil {
		parentNode := currentNode.Parent
		for _, node := range visibleNodes {
			if node.ID == parentNode.ID {
				return parentNode
			}
		}
		// If the parent is not visible, recursively search upwards
		return tv.findNearestVisibleNode(parentNode, visibleNodes)
	}

	// If at the root and it's not visible, return the first visible node
	if len(visibleNodes) > 0 {
		return visibleNodes[0]
	}
	return nil // No visible nodes found
}

// findPreviousVisibleNode finds the previous visible node in the tree
func (tv *Treeview) findPreviousVisibleNode(currentNode *TreeNode) *TreeNode {
	if currentNode == nil {
		return nil
	}

	if currentNode.Parent == nil {
		// If at the root, there's no previous node
		return nil
	}

	parent := currentNode.Parent
	siblings := parent.Children
	currentIndex := -1
	for i, sibling := range siblings {
		if sibling.ID == currentNode.ID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		// Node not found among siblings, something is wrong
		return nil
	}

	if currentIndex == 0 {
		// If the current node is the first child, return the parent
		return parent
	}

	previousSibling := siblings[currentIndex-1]
	return tv.findLastVisibleDescendant(previousSibling)
}

// findLastVisibleDescendant finds the last visible descendant of a node
func (tv *Treeview) findLastVisibleDescendant(node *TreeNode) *TreeNode {
	if !node.ExpandedState || len(node.Children) == 0 {
		return node
	}
	// Since node is expanded and has children, go to the last child
	lastChild := node.Children[len(node.Children)-1]
	return tv.findLastVisibleDescendant(lastChild)
}

// updateVisibleNodes updates the visibleNodes slice based on scrollOffset and node expansion.
func (tv *Treeview) updateVisibleNodes() {
	var allVisible []*TreeNode
	var traverse func(node *TreeNode)
	traverse = func(node *TreeNode) {
		allVisible = append(allVisible, node)
		if node.ExpandedState {
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}
	for _, root := range tv.opts.nodes {
		traverse(root)
	}

	tv.totalContentHeight = len(allVisible)

	// Clamp scrollOffset
	if tv.scrollOffset > tv.totalContentHeight-tv.canvasHeight {
		tv.scrollOffset = tv.totalContentHeight - tv.canvasHeight
		if tv.scrollOffset < 0 {
			tv.scrollOffset = 0
		}
	}

	tv.visibleNodes = allVisible
}

// truncateString truncates a string to fit within a specified width, appending "..." if truncated.
func truncateString(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}

	ellipsis := "…"
	ellipsisWidth := runewidth.StringWidth(ellipsis)

	// Start truncating characters from the string
	truncatedWidth := 0
	truncatedString := ""

	for _, r := range s {
		charWidth := runewidth.RuneWidth(r)
		if truncatedWidth+charWidth+ellipsisWidth > maxWidth {
			break
		}
		truncatedString += string(r)
		truncatedWidth += charWidth
	}

	return truncatedString + ellipsis
}
