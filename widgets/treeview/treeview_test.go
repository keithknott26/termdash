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
	area  image.Rectangle
}

// NewMockCanvas creates a new mock canvas.
func NewMockCanvas(width, height int) *MockCanvas {
	return &MockCanvas{
		Cells: make(map[image.Point]rune),
		area:  image.Rect(0, 0, width, height),
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
	return mc.area
}

// MockMeta is a mock implementation of widgetapi.Meta for testing purposes.
type MockMeta struct {
	area image.Rectangle
}

// NewMockMeta creates a new mock widgetapi.Meta.
func NewMockMeta(width, height int) *MockMeta {
	return &MockMeta{
		area: image.Rect(0, 0, width, height),
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

	tv, err := New(Nodes(root...), Indentation(4), Icons("▶", "▼", "•"))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	if tv.selectedNode == nil {
		t.Errorf("Expected selectedNode to be initialized, got nil")
	}

	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	if len(tv.opts.nodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(tv.opts.nodes))
	}

	if tv.indentationPerLevel != 4 {
		t.Errorf("Expected indentationPerLevel to be 4, got %d", tv.indentationPerLevel)
	}

	if tv.expandedIcon != "▼" || tv.collapsedIcon != "▶" || tv.leafIcon != "•" {
		t.Errorf("Icons not set correctly")
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

	tv, err := New(Nodes(root...))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially selected node should be "Root"
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down
	tv.Next()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down
	tv.Next()
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up
	tv.Previous()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up to root
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

	// Initially, Root should be expanded
	if !root[0].ExpandedState {
		t.Errorf("Expected Root to be expanded by default")
	}

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

	// Allow goroutine to run
	time.Sleep(100 * time.Millisecond)

	if !onClickCalled {
		t.Errorf("Expected OnClick to be called for Child1")
	}

	if child1.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be false after OnClick")
	}
}

// TestMouseScroll tests mouse wheel scrolling functionality.
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

	mouseEvent := &terminalapi.Mouse{
		Button:   mouse.ButtonWheelDown,
		Position: image.Point{X: 0, Y: 0},
	}

	err = tv.Mouse(mouseEvent, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Mouse method returned an error: %v", err)
	}

	// After scrolling down, scrollOffset should be clamped to maxOffset=2
	if tv.scrollOffset != 2 {
		t.Errorf("Expected scrollOffset to be clamped to 2, got %d", tv.scrollOffset)
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

// TestKeyboardScroll tests that keyboard navigation scrolls the viewport to keep the selected node visible.
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

	// Initial selection is "Root" at index 0, scrollOffset = 0
	if tv.scrollOffset != 0 {
		t.Errorf("Expected initial scrollOffset to be 0, got %d", tv.scrollOffset)
	}

	// Navigate down to "Child1" (index 1)
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to remain 0, got %d", tv.scrollOffset)
	}

	// Navigate down to "Child2" (index 2)
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to remain 0, got %d", tv.scrollOffset)
	}

	// Navigate down to "Child3" (index 3) - should adjust scrollOffset to 1
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child3" {
		t.Errorf("Expected selectedNode to be 'Child3', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}

	// Navigate down to "Child4" (index 4) - should adjust scrollOffset to 2
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child4" {
		t.Errorf("Expected selectedNode to be 'Child4', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 2 {
		t.Errorf("Expected scrollOffset to be 2, got %d", tv.scrollOffset)
	}

	// Navigate down to "Child5" (index 5) - scrollOffset should remain 2
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child5" {
		t.Errorf("Expected selectedNode to be 'Child5', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 2 {
		t.Errorf("Expected scrollOffset to remain 2, got %d", tv.scrollOffset)
	}

	// Navigate up to "Child4" (index 4) - scrollOffset should remain 2
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child4" {
		t.Errorf("Expected selectedNode to be 'Child4', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 2 {
		t.Errorf("Expected scrollOffset to remain 2, got %d", tv.scrollOffset)
	}

	// Navigate up to "Child3" (index 3) - scrollOffset should adjust to 1
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child3" {
		t.Errorf("Expected selectedNode to be 'Child3', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}

	// Navigate up to "Child2" (index 2) - scrollOffset should adjust to 0
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to be 0, got %d", tv.scrollOffset)
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

	tv, err := New(Nodes(root...), WaitingIcons([]string{"|", "/", "-", "\\"}))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Simulate the spinner ticker
	tv.spinnerTicker = time.NewTicker(10 * time.Millisecond)
	go tv.runSpinner()
	defer tv.StopSpinnerTicker()

	// Click on "Child1" to trigger OnClick and spinner
	tv.selectedNode = root[0].Children[0]
	err = tv.handleNodeClick(tv.selectedNode)
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	// Spinner should be active
	if !tv.selectedNode.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be true")
	}

	// Wait for OnClick to complete
	time.Sleep(100 * time.Millisecond)

	// Spinner should be inactive
	if tv.selectedNode.ShowSpinner {
		t.Errorf("Expected ShowSpinner to be false after OnClick")
	}

	// OnClick should have been called
	if !onClickCalled {
		t.Errorf("Expected OnClick to have been called")
	}
}

// TestUpdateVisibleNodes tests that visibleNodes are updated correctly based on expansion and scroll.
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

	// Set scrollOffset and verify
	tv.scrollOffset = 2
	tv.canvasHeight = 2
	tv.updateVisibleNodes()
	if tv.scrollOffset != 2 {
		t.Errorf("Expected scrollOffset to be 2, got %d", tv.scrollOffset)
	}

	// Check visibleNodes slice
	expectedVisible := []*TreeNode{root[0].Children[2], root[0].Children[1]}
	for i, node := range expectedVisible {
		if tv.visibleNodes[tv.scrollOffset+i].Label != node.Label {
			t.Errorf("Expected node '%s' at position %d, got '%s'", node.Label, i, tv.visibleNodes[tv.scrollOffset+i].Label)
		}
	}
}

// TestNodeExpansionAndCollapse ensures nodes are expanded and collapsed properly.
func TestNodeExpansionAndCollapse(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{
					Label: "Child1",
					Children: []*TreeNode{
						{Label: "Grandchild1"},
					},
				},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially all nodes should be visible
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 { // Root + Child1 + Grandchild1
		t.Errorf("Expected 3 visible nodes, got %d", len(tv.visibleNodes))
	}

	// Collapse Child1
	root[0].Children[0].ExpandedState = false
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 2 { // Root + Child1
		t.Errorf("Expected 2 visible nodes after collapsing Child1, got %d", len(tv.visibleNodes))
	}

	// Collapse Root
	root[0].ExpandedState = false
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 1 { // Only Root
		t.Errorf("Expected 1 visible node after collapsing Root, got %d", len(tv.visibleNodes))
	}

	// Expand Root and Child1
	root[0].ExpandedState = true
	root[0].Children[0].ExpandedState = true
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 {
		t.Errorf("Expected 3 visible nodes after expanding Root and Child1, got %d", len(tv.visibleNodes))
	}
}

// TestScrollLimits tests that scrollOffset is properly clamped to valid bounds.
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

	// Mock a small canvas height to enable scrolling
	tv.canvasHeight = 2
	tv.updateVisibleNodes()

	// Ensure we can scroll down
	tv.scrollOffset = 1
	tv.updateVisibleNodes()
	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}

	// Ensure scrollOffset does not exceed total height
	tv.scrollOffset = 100 // Excessive scroll offset
	tv.updateVisibleNodes()
	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be clamped to 1, got %d", tv.scrollOffset)
	}

	// Ensure scrollOffset stays at 0 when scrolling up
	tv.scrollOffset = -10 // Excessively low scroll offset
	tv.updateVisibleNodes()
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to be clamped to 0, got %d", tv.scrollOffset)
	}
}

// TestSelectNoVisibleNodes tests the Select method when no nodes are visible.
func TestSelectNoVisibleNodes(t *testing.T) {
	root := []*TreeNode{
		{Label: "Root", ExpandedState: false}, // Root is collapsed
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Try selecting when no nodes are visible
	label, err := tv.Select()
	if err == nil {
		t.Error("Expected an error when selecting with no visible nodes")
	}
	if label != "" {
		t.Errorf("Expected empty label when selecting with no visible nodes, got: %s", label)
	}
}

// TestKeyboardNonArrowKeys tests that non-arrow keys behave correctly.
func TestKeyboardNonArrowKeys(t *testing.T) {
	root := []*TreeNode{
		{Label: "Root"},
	}

	tv, err := New(Nodes(root...))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Press space to simulate selection toggle
	err = tv.Keyboard(&terminalapi.Keyboard{Key: ' '}, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Expected no error for spacebar key, got: %v", err)
	}

	// Press an unrelated key (e.g., 'a') and expect no action
	err = tv.Keyboard(&terminalapi.Keyboard{Key: 'a'}, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Expected no error for unrelated key press, got: %v", err)
	}
}
