package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type usersView struct {
	filePath string
	mu       sync.RWMutex
}

func NewUsersView(filePersistenceConfig *FilePersistenceConfig) model.UsersView {

	dataDir := filePersistenceConfig.DataDir

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create data directory: %v", err))
	}

	return &usersView{
		filePath: filepath.Join(dataDir, "users.json"),
	}
}

func (u *usersView) ById(id string) (*model.UserProjection, error) {
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
	var users map[string]*model.UserProjection
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

func (u *usersView) SubscribedTo() map[string]ddd.HandleEvent {
	subscriptions := make(map[string]ddd.HandleEvent)
	subscriptions[ddd.EventType(model.UserRegistered{})] = u.onUserRegistered
	return subscriptions
}

func (u *usersView) onUserRegistered(event ddd.Event) error {
	// Identify event type
	if event.Type() != ddd.EventType(model.UserRegistered{}) {
		return fmt.Errorf("unexpected event type: %s", event.Type())
	}

	return nil
}
