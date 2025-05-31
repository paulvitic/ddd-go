package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type users struct {
	filePath string
	mu       sync.RWMutex
}

func NewUsers(filePersistenceConfig *FilePersistenceConfig) model.Users {
	// Use the same data directory as the repository
	dataDir := filePersistenceConfig.DataDir
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create data directory: %v", err))
	}

	return &users{
		filePath: filepath.Join(dataDir, "users.json"),
	}
}

func (u *users) ById(id string) (*model.UserView, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Read the JSON file
	data, err := os.ReadFile(u.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("user not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("user not found with id: %s", id)
	}

	// Parse the JSON data
	var users map[string]*model.UserView
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Find the user
	user, exists := users[id]
	if !exists {
		return nil, fmt.Errorf("user not found with id: %s", id)
	}

	// Convert User to UserView
	return user, nil
}
