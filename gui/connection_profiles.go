package gui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ConnectionProfile представляет профиль подключения к ККТ
type ConnectionProfile struct {
	SerialNumber   string    `json:"serial_number"`   // Серийный номер ККТ (уникальный идентификатор)
	ConnectionType int       `json:"connection_type"` // 0 = COM, 6 = TCP
	ComName        string    `json:"com_name"`        // Например "COM9"
	BaudRate       int       `json:"baud_rate"`       // Например 115200
	IPAddress      string    `json:"ip_address"`      // Например "192.168.1.100"
	TCPPort        int       `json:"tcp_port"`        // Например 8200
	FirmwareVer    string    `json:"firmware_ver"`    // Версия прошивки ККТ
	ModelName      string    `json:"model_name"`      // Модель ККТ
	LastUsed       time.Time `json:"last_used"`       // Время последнего успешного подключения
}

// ProfilesStorage управляет хранением профилей подключения
type ProfilesStorage struct {
	mu       sync.RWMutex
	profiles []*ConnectionProfile
	filePath string
}

// profilesData используется для сериализации/десериализации JSON
type profilesData struct {
	Profiles []*ConnectionProfile `json:"profiles"`
}

var (
	// Глобальный экземпляр хранилища профилей
	profileStorage *ProfilesStorage
	once           sync.Once
)

// initProfilesStorage инициализирует хранилище профилей
func initProfilesStorage() {
	once.Do(func() {
		// Определяем путь к profiles.json рядом с исполняемым файлом
		exePath, err := os.Executable()
		if err != nil {
			log.Printf("[PROFILES] Ошибка получения пути к исполняемому файлу: %v", err)
			exePath = "." // fallback to current directory
		}
		var dir string
		// Обработка запуска через 'go run' (временная папка)
		if strings.Contains(exePath, "Temp") || strings.Contains(exePath, "go-build") {
			dir, err = os.Getwd()
			if err != nil {
				log.Printf("[PROFILES] Ошибка получения рабочей директории: %v", err)
				dir = "."
			}
		} else {
			dir = filepath.Dir(exePath)
		}
		filePath := filepath.Join(dir, "profiles.json")

		profileStorage = &ProfilesStorage{
			filePath: filePath,
			profiles: make([]*ConnectionProfile, 0),
		}
		// Пытаемся загрузить профили при инициализации
		_ = profileStorage.LoadProfiles()
	})
}

// LoadProfiles загружает профили из JSON-файла
func LoadProfiles() error {
	initProfilesStorage()
	return profileStorage.LoadProfiles()
}

func (s *ProfilesStorage) LoadProfiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[PROFILES] Файл профилей не найден (%s), создаем пустой список", s.filePath)
			s.profiles = make([]*ConnectionProfile, 0)
			return nil
		}
		log.Printf("[PROFILES] Ошибка чтения файла профилей (%s): %v", s.filePath, err)
		return fmt.Errorf("ошибка чтения файла профилей: %w", err)
	}

	var pd profilesData
	if err := json.Unmarshal(data, &pd); err != nil {
		log.Printf("[PROFILES] Ошибка разбора JSON файла профилей (%s): %v, сбрасываем список", s.filePath, err)
		s.profiles = make([]*ConnectionProfile, 0)
		return fmt.Errorf("ошибка разбора JSON: %w", err)
	}

	log.Printf("[PROFILES] Загружено %d профилей из файла %s", len(pd.Profiles), s.filePath)
	s.profiles = pd.Profiles
	return nil
}

// SaveProfiles (Публичный) - сохраняет профили в JSON-файл
func SaveProfiles() error {
	initProfilesStorage()
	return profileStorage.SaveProfiles()
}

func (s *ProfilesStorage) SaveProfiles() error {
	s.mu.Lock() // Блокируем на запись для внешней консистентности
	defer s.mu.Unlock()
	return s.saveProfilesLocked()
}

// saveProfilesLocked (Приватный) - выполняет запись, НЕ блокируя мьютекс (предполагает, что он уже захвачен)
func (s *ProfilesStorage) saveProfilesLocked() error {
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
func GetProfiles() []*ConnectionProfile {
	initProfilesStorage()
	return profileStorage.GetProfiles()
}

func (s *ProfilesStorage) GetProfiles() []*ConnectionProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ConnectionProfile, len(s.profiles))
	copy(result, s.profiles)

	// Сортировка: самый свежий (LastUsed) в начале
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUsed.After(result[j].LastUsed)
	})

	return result
}

// UpsertProfile добавляет или обновляет профиль
func UpsertProfile(profile *ConnectionProfile) error {
	initProfilesStorage()
	return profileStorage.UpsertProfile(profile)
}

func (s *ProfilesStorage) UpsertProfile(profile *ConnectionProfile) error {
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

	// Вызываем внутренние методы, которые НЕ используют Lock(), так как мы уже держим Lock
	return s.saveAndLimitLocked()
}

// DeleteProfile удаляет профиль
func DeleteProfile(serialNumber string) error {
	initProfilesStorage()
	return profileStorage.DeleteProfile(serialNumber)
}

func (s *ProfilesStorage) DeleteProfile(serialNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.profiles {
		if p.SerialNumber == serialNumber {
			s.profiles = append(s.profiles[:i], s.profiles[i+1:]...)
			// Используем внутренний метод сохранения
			return s.saveProfilesLocked()
		}
	}

	return fmt.Errorf("профиль с серийным номером %s не найден", serialNumber)
}

// FindProfile находит профиль
func FindProfile(serialNumber string) *ConnectionProfile {
	initProfilesStorage()
	return profileStorage.FindProfile(serialNumber)
}

func (s *ProfilesStorage) FindProfile(serialNumber string) *ConnectionProfile {
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
func ClearProfiles() error {
	initProfilesStorage()
	return profileStorage.ClearProfiles()
}

func (s *ProfilesStorage) ClearProfiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.profiles = make([]*ConnectionProfile, 0)
	// Используем внутренний метод
	return s.saveProfilesLocked()
}

// saveAndLimitLocked (Приватный) - лимитирует и сохраняет, предполагая, что Lock уже взят
func (s *ProfilesStorage) saveAndLimitLocked() error {
	if len(s.profiles) > 20 {
		// Сортируем
		sort.Slice(s.profiles, func(i, j int) bool {
			return s.profiles[i].LastUsed.After(s.profiles[j].LastUsed)
		})
		s.profiles = s.profiles[:20]
	}
	return s.saveProfilesLocked()
}

// UpdateLastUsed обновляет время последнего использования
func UpdateLastUsed(serialNumber string) error {
	initProfilesStorage()
	return profileStorage.UpdateLastUsed(serialNumber)
}

func (s *ProfilesStorage) UpdateLastUsed(serialNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.profiles {
		if p.SerialNumber == serialNumber {
			s.profiles[i].LastUsed = time.Now()
			// Используем внутренний метод
			return s.saveProfilesLocked()
		}
	}
	return nil
}

// DisplayString возвращает строку для UI
func (p *ConnectionProfile) DisplayString() string {
	var connInfo string
	if p.ConnectionType == 0 {
		connInfo = fmt.Sprintf("%s", p.ComName)
	} else {
		connInfo = fmt.Sprintf("%s:%d", p.IPAddress, p.TCPPort)
	}

	fwVer := p.FirmwareVer
	if fwVer == "" {
		fwVer = "—"
	} else {
		fwVer = "v" + fwVer
	}

	return fmt.Sprintf("SN%s - %s - %s", p.SerialNumber, connInfo, fwVer)
}
