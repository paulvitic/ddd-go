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

// Implements ddd.Repository[model.User]
type userRepository struct {
	filePath string
	eventBus ddd.EventBus
	mu       sync.RWMutex
}

func NewUserRepository(eventBus ddd.EventBus) ddd.Repository[model.User] {
	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create data directory: %v", err))
	}

	return &userRepository{
		filePath: filepath.Join(dataDir, "users.json"),
		eventBus: eventBus,
	}
}

func (r *userRepository) Save(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Read existing data
	users, err := r.loadAll()
	if err != nil {
		return err
	}

	// Add or update user
	users[user.ID().String()] = user

	// Write back to file
	if err := r.saveAll(users); err != nil {
		return err
	} else {
		for _, event := range user.GetAllEvents() {
			r.eventBus.Dispatch(event)
		}
		return nil
	}
}

func (r *userRepository) Update(user *model.User) error {
	return r.Save(user)
}

func (r *userRepository) Load(id ddd.ID) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users, err := r.loadAll()
	if err != nil {
		return nil, err
	}

	if user, exists := users[id.String()]; exists {
		return user, nil
	}

	return nil, fmt.Errorf("user not found with id: %s", id)
}

func (r *userRepository) Delete(id ddd.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	users, err := r.loadAll()
	if err != nil {
		return err
	}

	delete(users, id.String())
	return r.saveAll(users)
}

func (r *userRepository) loadAll() (map[string]*model.User, error) {
	// Create file if it doesn't exist
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return make(map[string]*model.User), nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return make(map[string]*model.User), nil
	}

	var users map[string]*model.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return users, nil
}

func (r *userRepository) saveAll(users map[string]*model.User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
