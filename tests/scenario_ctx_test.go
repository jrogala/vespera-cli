package tests

import (
	"github.com/jrogala/vespera-cli/client"
)

// scenarioCtx holds per-scenario state.
type scenarioCtx struct {
	mockConn *mockFTPConn
	client   *client.FTPClient

	// results from the latest "When" step
	lastErr          error
	statusResult     any
	observationsList any
	filesList        any
	treeResult       any
}

func newScenarioCtx() *scenarioCtx {
	mock := newMockFTPConn()
	// Create a client with the mock connection
	c := client.NewFTPClientWithConn("10.0.0.1", 21, mock)
	return &scenarioCtx{
		mockConn: mock,
		client:   c,
	}
}
