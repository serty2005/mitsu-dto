package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"
)

// FileProfileRepository реализует интерфейс ports.ProfileRepository с использованием JSON-файла для хранения.
type FileProfileRepository struct {
	mu       sync.Mutex
	filePath string
	profiles []*models.ConnectionProfile
}

// NewFileProfileRepository создает новый экземпляр FileProfileRepository с указанным путем к файлу.
func NewFileProfileRepository(filePath string) (ports.ProfileRepository, error) {
	repo := &FileProfileRepository{
		filePath: filePath,
	}

	// Загружаем профили при инициализации
	if err := repo.loadFromFile(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации репозитория: %w", err)
	}

	return repo, nil
}

// LoadProfiles загружает все профили из хранилища.
func (r *FileProfileRepository) LoadProfiles() ([]*models.ConnectionProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.loadFromFile(); err != nil {
		return nil, err
	}

	// Возвращаем копию для защиты данных от внешнего изменения
	result := make([]*models.ConnectionProfile, len(r.profiles))
	copy(result, r.profiles)
	return result, nil
}

// SaveProfiles сохраняет все профили в хранилище.
func (r *FileProfileRepository) SaveProfiles(profiles []*models.ConnectionProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.profiles = make([]*models.ConnectionProfile, len(profiles))
	copy(r.profiles, profiles)

	return r.saveToFile()
}

// UpsertProfile добавляет или обновляет профиль.
func (r *FileProfileRepository) UpsertProfile(profile *models.ConnectionProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	found := false
	for i, p := range r.profiles {
		if p.SerialNumber == profile.SerialNumber {
			r.profiles[i] = profile
			found = true
			break
		}
	}

	if !found {
		r.profiles = append(r.profiles, profile)
	}

	return r.saveToFile()
}

// DeleteProfile удаляет профиль по серийному номеру.
func (r *FileProfileRepository) DeleteProfile(serialNumber string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.profiles {
		if p.SerialNumber == serialNumber {
			r.profiles = append(r.profiles[:i], r.profiles[i+1:]...)
			return r.saveToFile()
		}
	}

	return fmt.Errorf("профиль с серийным номером %s не найден", serialNumber)
}

// FindProfile находит профиль по серийному номеру.
func (r *FileProfileRepository) FindProfile(serialNumber string) (*models.ConnectionProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.profiles {
		if p.SerialNumber == serialNumber {
			return p, nil
		}
	}

	return nil, nil
}

// ClearProfiles очищает все профили.
func (r *FileProfileRepository) ClearProfiles() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.profiles = make([]*models.ConnectionProfile, 0)
	return r.saveToFile()
}

// UpdateLastUsed обновляет время последнего использования профиля.
func (r *FileProfileRepository) UpdateLastUsed(serialNumber string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.profiles {
		if p.SerialNumber == serialNumber {
			r.profiles[i].LastUsed = time.Now()
			return r.saveToFile()
		}
	}

	return nil
}

// loadFromFile загружает профили из JSON-файла (не потокобезопасно, предназначено для внутреннего использования).
func (r *FileProfileRepository) loadFromFile() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			r.profiles = make([]*models.ConnectionProfile, 0)
			return nil
		}
		return fmt.Errorf("ошибка чтения файла профилей: %w", err)
	}

	var pd struct {
		Profiles []*models.ConnectionProfile `json:"profiles"`
	}

	if err := json.Unmarshal(data, &pd); err != nil {
		r.profiles = make([]*models.ConnectionProfile, 0)
		return fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	r.profiles = pd.Profiles
	return nil
}

// saveToFile сохраняет профили в JSON-файл (не потокобезопасно, предназначено для внутреннего использования).
func (r *FileProfileRepository) saveToFile() error {
	// Создаем директорию, если она не существует
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	data := struct {
		Profiles []*models.ConnectionProfile `json:"profiles"`
	}{
		Profiles: r.profiles,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	if err := os.WriteFile(r.filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла профилей: %w", err)
	}

	return nil
}
