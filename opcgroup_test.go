//go:build windows

package opcda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOPCGroup_SetName_Mocked(t *testing.T) {
	mockGroup := &mockGroupProvider{
		SetNameFn: func(name string) error {
			assert.Equal(t, "new_name", name)
			return nil
		},
	}
	group := &OPCGroup{
		groupProvider: mockGroup,
		groupName:     "old_name",
	}
	err := group.SetName("new_name")
	assert.NoError(t, err)
	assert.Equal(t, "new_name", group.GetName())
}

func TestOPCGroup_IsActive_Mocked(t *testing.T) {
	mockGroup := &mockGroupProvider{
		GetStateFn: func() (uint32, bool, string, int32, float32, uint32, uint32, uint32, error) {
			return 1000, false, "mock", 0, 0, 1033, 0, 0, nil
		},
	}
	group := &OPCGroup{
		groupProvider: mockGroup,
	}
	assert.False(t, group.GetIsActive())
}
