//go:build windows

package opcda

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wends155/opcda/com"
)

func TestOPCServer_GetServerState_Mocked(t *testing.T) {
	mock := &mockServerProvider{
		GetStatusFn: func() (*com.ServerStatus, error) {
			return &com.ServerStatus{
				ServerState: OPC_STATUS_RUNNING,
			}, nil
		},
	}
	server := newOPCServerWithProvider(mock, "mock", "localhost")
	state, err := server.GetServerState()
	assert.NoError(t, err)
	assert.Equal(t, OPC_STATUS_RUNNING, state)
}

func TestOPCServer_GetLocaleID_Mocked(t *testing.T) {
	mock := &mockServerProvider{
		GetLocaleIDFn: func() (uint32, error) {
			return 1033, nil
		},
	}
	server := newOPCServerWithProvider(mock, "mock", "localhost")
	id, err := server.GetLocaleID()
	assert.NoError(t, err)
	assert.Equal(t, uint32(1033), id)
}
