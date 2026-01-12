package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"mitsuscanner/internal/models"
)

// profilesData используется для сериализации/десериализации JSON
type profilesData struct {
	Profiles []*models.ConnectionProfile `json:"profiles"`
}

// ProfilesStore управляет хранением профилей подключения
type ProfilesStore struct {
	mu       sync.RWMutex
	profiles []*models.ConnectionProfile
	filePath string
}

// NewProfilesStore создает новый экземпляр хранилища профилей
func NewProfilesStore(path string) *ProfilesStore {
	return &ProfilesStore{
		filePath: path,
		profiles: make([]*models.ConnectionProfile, 0),
	}
}

// LoadProfiles загружает профили из JSON-файла
func (s *ProfilesStore) LoadProfiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[PROFILES] Файл профилей не найден (%s), создаем пустой список", s.filePath)
			s.profiles = make([]*models.ConnectionProfile, 0)
			return nil
		}
		log.Printf("[PROFILES] Ошибка чтения файла профилей (%s): %v", s.filePath, err)
		return fmt.Errorf("ошибка чтения файла профилей: %w", err)
	}

	var pd profilesData
	if err := json.Unmarshal(data, &pd); err != nil {
		log.Printf("[PROFILES] Ошибка разбора JSON файла профилей (%s): %v, сбрасываем список", s.filePath, err)
		s.profiles = make([]*models.ConnectionProfile, 0)
		return fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	log.Printf("[PROFILES] Загружено %d профилей из файла %s", len(pd.Profiles), s.filePath)
	s.profiles = pd.Profiles
	return nil
}

// SaveProfiles сохраняет профили в JSON-файл
func (s *ProfilesStore) SaveProfiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveProfilesLocked()
}

// saveProfilesLocked выполняет запись, предполагая, что мьютекс уже захвачен
func (s *ProfilesStore) saveProfilesLocked() error {
	log.Printf("[PROFILES] Начинаем сохранение %d профилей в файл %s", len(s.profiles), s.filePath)

	data := profilesData{
		Profiles: s.profiles,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("[PROFILES] Ошибка сериализации JSON: %v", err)
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	if err := os.WriteFile(s.filePath, jsonData, 0644); err != nil {
		log.Printf("[PROFILES] Ошибка записи файла профилей (%s): %v", s.filePath, err)
		return fmt.Errorf("ошибка записи файла профилей: %w", err)
	}

	log.Printf("[PROFILES] Успешно сохранено %d профилей", len(s.profiles))
	return nil
}

// GetProfiles возвращает все профили
func (s *ProfilesStore) GetProfiles() []*models.ConnectionProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.ConnectionProfile, len(s.profiles))
	copy(result, s.profiles)

	// Сортировка: самый свежий (LastUsed) в начале
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUsed.After(result[j].LastUsed)
	})

	return result
}

// UpsertProfile добавляет или обновляет профиль
func (s *ProfilesStore) UpsertProfile(profile *models.ConnectionProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[PROFILES] Добавляем/обновляем профиль SN: %s", profile.SerialNumber)

	found := false
	for i, p := range s.profiles {
		if p.SerialNumber == profile.SerialNumber {
			s.profiles[i] = profile
			found = true
			break
		}
	}

	if !found {
		s.profiles = append(s.profiles, profile)
	}

	return s.saveAndLimitLocked()
}

// DeleteProfile удаляет профиль
func (s *ProfilesStore) DeleteProfile(serialNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.profiles {
		if p.SerialNumber == serialNumber {
			s.profiles = append(s.profiles[:i], s.profiles[i+1:]...)
			return s.saveProfilesLocked()
		}
	}

	return fmt.Errorf("профиль с серийным номером %s не найден", serialNumber)
}

// FindProfile находит профиль
func (s *ProfilesStore) FindProfile(serialNumber string) *models.ConnectionProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.profiles {
		if p.SerialNumber == serialNumber {
			return p
		}
	}
	return nil
}

// ClearProfiles очищает все профили
func (s *ProfilesStore) ClearProfiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.profiles = make([]*models.ConnectionProfile, 0)
	return s.saveProfilesLocked()
}

// saveAndLimitLocked лимитирует и сохраняет, предполагая, что Lock уже взят
func (s *ProfilesStore) saveAndLimitLocked() error {
	if len(s.profiles) > 20 {
		sort.Slice(s.profiles, func(i, j int) bool {
			return s.profiles[i].LastUsed.After(s.profiles[j].LastUsed)
		})
		s.profiles = s.profiles[:20]
	}
	return s.saveProfilesLocked()
}

// UpdateLastUsed обновляет время последнего использования
func (s *ProfilesStore) UpdateLastUsed(serialNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.profiles {
		if p.SerialNumber == serialNumber {
			s.profiles[i].LastUsed = time.Now()
			return s.saveProfilesLocked()
		}
	}
	return nil
}
