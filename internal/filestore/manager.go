package filestore

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/sjzar/file-store-mcp/internal/mcp"
	"github.com/sjzar/file-store-mcp/internal/storage"
)

type Manager struct {
	storage *storage.Service
	mcp     *mcp.Service
}

func New() *Manager {

	storage := storage.NewService()

	mcp := mcp.NewService(storage)

	return &Manager{
		storage: storage,
		mcp:     mcp,
	}
}

func (m *Manager) ServeStdio() error {
	return server.ServeStdio(m.mcp.Server)
}

func (m *Manager) NewSSEServer() *server.SSEServer {
	return server.NewSSEServer(m.mcp.Server)
}
