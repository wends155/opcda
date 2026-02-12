//go:build windows

package opcda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wends155/opcda/com"
)

func TestOPCItem_Read_Mocked(t *testing.T) {
	now := time.Now()
	mockGroup := &mockGroupProvider{
		SyncReadFn: func(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []int32, error) {
			return []*com.ItemState{
				{
					Value:     123.45,
					Quality:   192,
					Timestamp: now,
				},
			}, []int32{0}, nil
		},
	}
	item := &OPCItem{
		groupProvider: mockGroup,
		serverHandle:  1,
	}
	val, q, ts, err := item.Read(OPC_DS_CACHE)
	assert.NoError(t, err)
	assert.Equal(t, 123.45, val)
	assert.Equal(t, uint16(192), q)
	assert.Equal(t, now, ts)
}

func TestOPCItem_Write_Mocked(t *testing.T) {
	mockGroup := &mockGroupProvider{
		SyncWriteFn: func(serverHandles []uint32, values []com.VARIANT) ([]int32, error) {
			assert.Equal(t, uint32(1), serverHandles[0])
			return []int32{0}, nil
		},
	}
	item := &OPCItem{
		groupProvider: mockGroup,
		serverHandle:  1,
	}
	err := item.Write(float64(1.23))
	assert.NoError(t, err)
}
