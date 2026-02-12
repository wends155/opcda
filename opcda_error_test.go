//go:build windows

package opcda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilServerMethods(t *testing.T) {
	var s *OPCServer
	// Test methods on nil server
	assert.NotPanics(t, func() { s.Disconnect() })
	assert.NotPanics(t, func() { s.GetOPCGroups() })
	_, err := s.CreateBrowser()
	assert.Error(t, err)
	_, err = s.GetLocaleID()
	assert.Error(t, err)
}

func TestNilBrowserMethods(t *testing.T) {
	var b *OPCBrowser
	// Test methods on nil browser
	assert.NotPanics(t, func() { b.Release() })
	_, err := b.GetCurrentPosition()
	assert.Error(t, err)
	err = b.ShowBranches()
	assert.Error(t, err)
}

func TestNilGroupsMethods(t *testing.T) {
	var gs *OPCGroups
	assert.NotPanics(t, func() { gs.GetCount() })
	_, err := gs.Add("test")
	assert.Error(t, err)
}

func TestNilItemsMethods(t *testing.T) {
	var is *OPCItems
	assert.NotPanics(t, func() { is.GetCount() })
	_, _, err := is.AddItems([]string{"tag"})
	assert.Error(t, err)
}
