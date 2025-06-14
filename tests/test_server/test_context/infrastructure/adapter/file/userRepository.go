package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/repository"
)

// Implements ddd.Repository[model.User]
type userRepository struct {
	logger   *ddd.Logger
	dataDir  string
	filePath string
	eventBus *ddd.EventBus
	mu       sync.RWMutex
}

func NewUserRepository(logger *ddd.Logger, eventBus *ddd.EventBus, filePersistenceConfig *FilePersistenceConfig) repository.UserRepository {
	return &userRepository{
		logger:   logger,
		dataDir:  filePersistenceConfig.DataDir,
		eventBus: eventBus,
	}
}

func (r *userRepository) OnInit() error {
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return err
	}
	r.filePath = filepath.Join(r.dataDir, "users.json")
	r.logger.Info("User repository initialized")
	return nil
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
		return r.eventBus.DispatchFrom(user)
	}
}

func (r *userRepository) Update(user *model.User) error {
	return r.eventBus.DispatchFrom(user)
}

func (r *userRepository) Load(id ddd.ID) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users, err := r.loadAll()
	if err != nil {
		return nil, err
	}

	if _, exists := users[id.String()]; exists {
		return model.LoadUser(id), nil
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
