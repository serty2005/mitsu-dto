package app

import (
	"fmt"
	"sync"

	"github.com/lxn/walk"

	"mitsuscanner/driver"
	"mitsuscanner/internal/storage"
)

// App представляет основное приложение.
type App struct {
	Storage    *storage.ProfilesStore
	MainWindow *walk.MainWindow

	mu     sync.RWMutex
	driver driver.Driver
}

// NewApp создает новый экземпляр приложения.
func NewApp(storagePath string) (*App, error) {
	store := storage.NewProfilesStore(storagePath)

	if err := store.LoadProfiles(); err != nil {
		return nil, fmt.Errorf("failed to load profiles: %w", err)
	}

	return &App{Storage: store}, nil
}

// SetDriver устанавливает активный драйвер (потокобезопасно)
func (a *App) SetDriver(d driver.Driver) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.driver = d
}

// GetDriver возвращает активный драйвер или nil (потокобезопасно)
func (a *App) GetDriver() driver.Driver {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.driver
}
