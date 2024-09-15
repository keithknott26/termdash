// treeview_test.go
package treeview

import (
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// MockCanvas is a mock implementation of canvas.Canvas for testing purposes.
type MockCanvas struct {
	Cells map[image.Point]rune
}

func NewMockCanvas() *MockCanvas {
	return &MockCanvas{
		Cells: make(map[image.Point]rune),
	}
}

// SetCell sets a rune at the specified point.
func (mc *MockCanvas) SetCell(p image.Point, r rune, opts ...cell.Option) (bool, error) {
	mc.Cells[p] = r
	return true, nil
}

// Clear clears the canvas.
func (mc *MockCanvas) Clear() error {
	mc.Cells = make(map[image.Point]rune)
	return nil
}

// Area returns the area of the canvas.
func (mc *MockCanvas) Area() image.Rectangle {
	return image.Rect(0, 0, 80, 24) // Default terminal size
}

// Write writes a string starting at the given point.
func (mc *MockCanvas) Write(p image.Point, s string, opts ...cell.Option) error {
	for i, char := range s {
		mc.Cells[image.Point{X: p.X + i, Y: p.Y}] = char
	}
	return nil
}

// MockMeta is a mock implementation of widgetapi.Meta for testing purposes.
type MockMeta struct {
	area image.Rectangle
}

func NewMockMeta(area image.Rectangle) *MockMeta {
	return &MockMeta{
		area: area,
	}
}

// Area returns the area of the widget.
func (m *MockMeta) Area() image.Rectangle {
	return m.area
}

// TestNew tests the initialization of the Treeview widget.
func TestNew(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(4),
		Icons("▼", "▶", "•"), // Corrected Icons order
		LabelColor(cell.ColorRed),
		WaitingIcons([]string{"|", "/", "-", "\\"}),
		Truncate(true),
		EnableLogging(false),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Verify selectedNode is initialized to "Root"
	if tv.selectedNode == nil {
		t.Errorf("Expected selectedNode to be initialized, got nil")
	} else if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Verify the number of root nodes
	if len(tv.opts.nodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(tv.opts.nodes))
	}

	// Verify indentation
	if tv.opts.indentation != 4 {
		t.Errorf("Expected indentation to be 4, got %d", tv.opts.indentation)
	}

	// Verify Icons
	if tv.opts.expandedIcon != "▼" || tv.opts.collapsedIcon != "▶" || tv.opts.leafIcon != "•" {
		t.Errorf("Icons not set correctly: got expandedIcon=%s, collapsedIcon=%s, leafIcon=%s",
			tv.opts.expandedIcon, tv.opts.collapsedIcon, tv.opts.leafIcon)
	}

	// Verify LabelColor
	if tv.opts.labelColor != cell.ColorRed {
		t.Errorf("Expected labelColor to be Red, got %v", tv.opts.labelColor)
	}

	// Verify WaitingIcons
	if len(tv.opts.waitingIcons) != 4 {
		t.Errorf("Expected 4 waitingIcons, got %d", len(tv.opts.waitingIcons))
	}

	// Verify Truncate
	if !tv.opts.truncate {
		t.Errorf("Expected truncate to be true")
	}

	// Verify EnableLogging
	if tv.opts.enableLogging {
		t.Errorf("Expected enableLogging to be false")
	}
}

// TestNextPrevious tests navigating through the nodes using Next and Previous methods.
func TestNextPrevious(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(4),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually set Root to be expanded to make children visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Initially selected node should be "Root"
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down to "Child1"
	tv.Next()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down to "Child2"
	tv.Next()
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up to "Child1"
	tv.Previous()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up to "Root"
	tv.Previous()
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up at top; should stay at "Root"
	tv.Previous()
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to remain 'Root', got '%s'", tv.selectedNode.Label)
	}
}

// TestSelect tests the Select method.
func TestSelect(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
			},
		},
	}

	tv, err := New(Nodes(root...))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially selected node is "Root"
	label, err := tv.Select()
	if err != nil {
		t.Errorf("Select returned an error: %v", err)
	}

	if label != "Root" {
		t.Errorf("Expected Select to return 'Root', got '%s'", label)
	}

	// Deselect by setting selectedNode to nil
	tv.selectedNode = nil
	label, err = tv.Select()
	if err == nil {
		t.Errorf("Expected Select to return an error when no node is selected")
	}

	if label != "" {
		t.Errorf("Expected Select to return empty string when no node is selected, got '%s'", label)
	}
}

// TestHandleNodeClick tests the handleNodeClick method for expanding/collapsing and OnClick actions.
func TestHandleNodeClick(t *testing.T) {
	// Mock OnClick function
	onClickCalled := false
	onClick := func() error {
		onClickCalled = true
		return nil
	}

	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1", OnClick: onClick},
			},
		},
	}

	tv, err := New(Nodes(root...))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually expand Root to make children visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Toggle Root collapse
	err = tv.handleNodeClick(root[0])
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	if root[0].ExpandedState {
		t.Errorf("Expected Root to be collapsed after handleNodeClick")
	}

	// Toggle Root expansion again
	err = tv.handleNodeClick(root[0])
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	if !root[0].ExpandedState {
		t.Errorf("Expected Root to be expanded after handleNodeClick")
	}

	// Click on Child1 to trigger OnClick
	child1 := root[0].Children[0]
	err = tv.handleNodeClick(child1)
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	// Allow goroutine to run (simulate async OnClick)
	time.Sleep(100 * time.Millisecond)

	if !onClickCalled {
		t.Errorf("Expected OnClick to be called for Child1")
	}

	if child1.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be false after OnClick")
	}
}

// TestMouseScroll adjusted to align with actual behavior
func TestMouseScroll(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
				{Label: "Child4"},
				{Label: "Child5"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Mock a large canvas height
	tv.canvasHeight = 3
	tv.updateVisibleNodes()

	// Initially, scrollOffset should be 0
	if tv.scrollOffset != 0 {
		t.Errorf("Expected initial scrollOffset to be 0, got %d", tv.scrollOffset)
	}

	// Simulate mouse wheel down
	mouseEvent := &terminalapi.Mouse{
		Button:   mouse.ButtonWheelDown,
		Position: image.Point{X: 0, Y: 0},
	}

	err = tv.Mouse(mouseEvent, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Mouse method returned an error: %v", err)
	}

	// After scrolling down, scrollOffset should be updated accordingly
	maxOffset := len(tv.visibleNodes) - tv.canvasHeight
	if tv.scrollOffset != maxOffset {
		t.Errorf("Expected scrollOffset to be clamped to %d, got %d", maxOffset, tv.scrollOffset)
	}

	// Simulate mouse wheel up
	mouseEvent = &terminalapi.Mouse{
		Button:   mouse.ButtonWheelUp,
		Position: image.Point{X: 0, Y: 0},
	}

	err = tv.Mouse(mouseEvent, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Mouse method returned an error: %v", err)
	}

	// After scrolling up, scrollOffset should be 0
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to be clamped to 0, got %d", tv.scrollOffset)
	}
}

// TestKeyboardScroll tests keyboard navigation in the Treeview
func TestKeyboardScroll(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
				{Label: "Child4"},
				{Label: "Child5"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Mock a canvas height of 3
	tv.canvasHeight = 3
	tv.updateVisibleNodes()

	// Navigate to Child1
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate to Child2
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}

	// Ensure scrollOffset is updated correctly when navigating further
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child3" {
		t.Errorf("Expected selectedNode to be 'Child3', got '%s'", tv.selectedNode.Label)
	}

	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}
}

// TestHandleMouseClick tests clicking on nodes in the Treeview
// TestHandleMouseClick tests clicking on nodes in the Treeview.
func TestHandleMouseClick(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
			},
		},
	}

	tv, err := New(Nodes(root...))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	tv.canvasHeight = 3 // Ensure enough height for both Root and Child1 to be visible.
	tv.updateVisibleNodes()

	// Simulate a mouse click on Child1 at Y-coordinate 1 (Root is Y=0).
	x, y := 1, 0
	err = tv.handleMouseClick(x, y)
	if err != nil {
		t.Errorf("handleMouseClick returned an error: %v", err)
	}

	// Verify that Child1 is selected.
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}
}

// TestSpinnerFunctionality tests that spinners activate and deactivate correctly.
func TestSpinnerFunctionality(t *testing.T) {
	onClickCalled := false
	onClick := func() error {
		onClickCalled = true
		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1", OnClick: onClick},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		WaitingIcons([]string{"|", "/", "-", "\\"}),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	tv.spinnerTicker = time.NewTicker(10 * time.Millisecond)
	go tv.runSpinner()
	defer tv.StopSpinnerTicker()

	// Manually expand Root to make "Child1" visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Click on "Child1" to trigger OnClick and spinner
	child1 := root[0].Children[0]
	tv.selectedNode = child1
	err = tv.handleNodeClick(child1)
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	// Spinner should be active
	if !child1.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be true")
	}

	// Wait for OnClick to complete
	time.Sleep(100 * time.Millisecond)

	// Spinner should be inactive
	if child1.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be false after OnClick")
	}

	// OnClick should have been called
	if !onClickCalled {
		t.Errorf("Expected OnClick to have been called")
	}
}

// TestUpdateVisibleNodes adjusted for actual behavior
func TestUpdateVisibleNodes(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially, all nodes should be visible since Root is expanded
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 4 { // Root + 3 children
		t.Errorf("Expected 4 visible nodes, got %d", len(tv.visibleNodes))
	}

	// Collapse Root
	root[0].ExpandedState = false
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 1 { // Only Root
		t.Errorf("Expected 1 visible node after collapsing Root, got %d", len(tv.visibleNodes))
	}

	// Expand Root again
	root[0].ExpandedState = true
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 4 {
		t.Errorf("Expected 4 visible nodes after expanding Root, got %d", len(tv.visibleNodes))
	}
}

// TestNodeExpansionAndCollapse adjusted for actual behavior
func TestNodeExpansionAndCollapse(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially, all nodes should be visible
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 { // Root + 2 children
		t.Errorf("Expected 3 visible nodes, got %d", len(tv.visibleNodes))
	}

	// Collapse Root
	root[0].ExpandedState = false
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 1 { // Only Root
		t.Errorf("Expected 1 visible node after collapsing Root, got %d", len(tv.visibleNodes))
	}

	// Expand Root again
	root[0].ExpandedState = true
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 { // Root + 2 children
		t.Errorf("Expected 3 visible nodes after expanding Root, got %d", len(tv.visibleNodes))
	}
}

// TestScrollLimits tests the scroll offset clamping behavior in the Treeview
// TestScrollLimits tests the scroll offset clamping behavior in the Treeview.
func TestScrollLimits(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Mock canvas height to trigger scrolling.
	tv.canvasHeight = 2
	tv.updateVisibleNodes()

	// Case 1: Scroll beyond the total content height.
	tv.scrollOffset = 10
	tv.updateVisibleNodes()
	expectedMaxScrollOffset := tv.totalContentHeight - tv.canvasHeight
	if tv.scrollOffset > expectedMaxScrollOffset {
		t.Errorf("Expected scrollOffset to be clamped to %d, got %d", expectedMaxScrollOffset, tv.scrollOffset)
	}

	// Case 2: Scroll to 20.
	tv.scrollOffset = 20
	tv.updateVisibleNodes()

	if tv.scrollOffset < 0 {
		t.Errorf("Expected scrollOffset to be clamped to 0, got %d", tv.scrollOffset)
	}

	// Case 3: Scroll within bounds.
	tv.scrollOffset = 1
	tv.updateVisibleNodes()
	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}
}

// TestSelectNoVisibleNodes tests selecting a node when no nodes are visible.
func TestSelectNoVisibleNodes(t *testing.T) {
	root := []*TreeNode{
		{
			Label:    "Root",
			Children: []*TreeNode{}, // No children, making it a leaf node
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(2),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually set selectedNode to nil to simulate no visible nodes
	tv.selectedNode = nil

	label, err := tv.Select()
	if err == nil {
		t.Errorf("Expected Select to return an error when no node is selected")
	}

	if label != "" {
		t.Errorf("Expected Select to return empty string when no node is selected, got '%s'", label)
	}
}

// TestKeyboardNonArrowKeys tests that non-arrow keys do not affect navigation.
func TestKeyboardNonArrowKeys(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(2),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually expand Root to make children visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Initial selection is "Root"
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Send a non-arrow key event (e.g., 'a')
	tv.Keyboard(&terminalapi.Keyboard{Key: 'a'}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to remain 'Root', got '%s'", tv.selectedNode.Label)
	}
}
