//go:build windows

package opcda

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wends155/opcda/com"
)

// mockBrowserAddressSpace is a mock implementation of the browserAddressSpace interface.
// It simulates a simple hierarchical address space for testing.
type mockBrowserAddressSpace struct {
	currentPath string
	branches    map[string][]string
	leaves      map[string][]string
}

func newMockBrowserAddressSpace() *mockBrowserAddressSpace {
	return &mockBrowserAddressSpace{
		currentPath: "",
		branches: map[string][]string{
			"":        {"Folder1", "Folder2"},
			"Folder1": {"SubFolder1"},
		},
		leaves: map[string][]string{
			"":           {"RootItem1"},
			"Folder1":    {"Item1", "Item2"},
			"SubFolder1": {"SubItem1"},
		},
	}
}

func (m *mockBrowserAddressSpace) GetItemID(leaf string) (string, error) {
	if leaf == "" {
		return m.currentPath, nil
	}
	if m.currentPath == "" {
		return leaf, nil
	}
	return fmt.Sprintf("%s.%s", m.currentPath, leaf), nil
}

func (m *mockBrowserAddressSpace) QueryOrganization() (com.OPCNAMESPACETYPE, error) {
	return OPC_NS_HIERARCHIAL, nil
}

func (m *mockBrowserAddressSpace) BrowseOPCItemIDs(filterType com.OPCBROWSETYPE, filter string, dataType uint16, accessRights uint32) ([]string, error) {
	switch filterType {
	case OPC_BRANCH:
		return m.branches[m.currentPath], nil
	case OPC_LEAF, OPC_FLAT:
		return m.leaves[m.currentPath], nil
	default:
		return nil, errors.New("invalid filter type")
	}
}

func (m *mockBrowserAddressSpace) ChangeBrowsePosition(dir com.OPCBROWSEDIRECTION, name string) error {
	switch dir {
	case OPC_BROWSE_UP:
		if m.currentPath == "SubFolder1" {
			m.currentPath = "Folder1"
		} else {
			m.currentPath = ""
		}
	case OPC_BROWSE_DOWN:
		if m.currentPath == "" && (name == "Folder1" || name == "Folder2") {
			m.currentPath = name
		} else if m.currentPath == "Folder1" && name == "SubFolder1" {
			m.currentPath = name
		} else {
			return errors.New("branch not found")
		}
	case OPC_BROWSE_TO:
		m.currentPath = name
	}
	return nil
}

func (m *mockBrowserAddressSpace) Release() uint32 {
	return 0
}

func TestOPCBrowser_MockNavigation(t *testing.T) {
	mock := newMockBrowserAddressSpace()
	browser := newOPCBrowserWithInterface(mock, nil)

	// Test Initial State
	pos, _ := browser.GetCurrentPosition()
	assert.Equal(t, "", pos)

	// Test Browse Branches
	err := browser.ShowBranches()
	assert.NoError(t, err)
	assert.Equal(t, 2, browser.GetCount())
	name, _ := browser.Item(0)
	assert.Equal(t, "Folder1", name)

	// Test Move Down
	err = browser.MoveDown("Folder1")
	assert.NoError(t, err)
	pos, _ = browser.GetCurrentPosition()
	assert.Equal(t, "Folder1", pos)

	// Test Browse Leafs
	err = browser.ShowLeafs(false)
	assert.NoError(t, err)
	assert.Equal(t, 2, browser.GetCount())
	name, _ = browser.Item(0)
	assert.Equal(t, "Item1", name)

	// Test GetItemID
	id, err := browser.GetItemID("Item1")
	assert.NoError(t, err)
	assert.Equal(t, "Folder1.Item1", id)

	// Test Move Up
	err = browser.MoveUp()
	assert.NoError(t, err)
	pos, _ = browser.GetCurrentPosition()
	assert.Equal(t, "", pos)
}

func ExampleOPCBrowser_ShowLeafs_mock() {
	// Initialize browser with mock address space
	mock := newMockBrowserAddressSpace()
	browser := newOPCBrowserWithInterface(mock, nil)

	// Navigate to Folder1
	browser.MoveDown("Folder1")

	// Show leafs in Folder1
	browser.ShowLeafs(false)
	count := browser.GetCount()
	for i := 0; i < count; i++ {
		name, _ := browser.Item(i)
		fmt.Println("Leaf:", name)
	}

	// Get fully qualified ItemID
	itemID, _ := browser.GetItemID("Item1")
	fmt.Println("ItemID:", itemID)

	// Output:
	// Leaf: Item1
	// Leaf: Item2
	// ItemID: Folder1.Item1
}

func ExampleConnect_error() {
	// Attempt to connect to a non-existent server
	_, err := Connect("NonExistent.ProgID.1", "localhost")
	if err != nil {
		fmt.Println("Caught expected connection error")
	}
	// Output: Caught expected connection error
}

func ExampleOPCServer_CreateBrowser_error() {
	// Attempt to create a browser from a nil server
	var server *OPCServer
	_, err := server.CreateBrowser()
	if err != nil {
		fmt.Println("Caught expected browser creation error")
	}
	// Output: Caught expected browser creation error
}
