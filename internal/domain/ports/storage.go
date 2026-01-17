package ports

import "mitsuscanner/internal/domain/models"

// ProfileRepository определяет интерфейс для хранения и управления профилями подключения.
// Реализация интерфейса будет находится в слое Infrastructure.
type ProfileRepository interface {
	// LoadProfiles загружает все профили из хранилища
	LoadProfiles() ([]*models.ConnectionProfile, error)

	// SaveProfiles сохраняет все профили в хранилище
	SaveProfiles(profiles []*models.ConnectionProfile) error

	// UpsertProfile добавляет или обновляет профиль
	UpsertProfile(profile *models.ConnectionProfile) error

	// DeleteProfile удаляет профиль по серийному номеру
	DeleteProfile(serialNumber string) error

	// FindProfile находит профиль по серийному номеру
	FindProfile(serialNumber string) (*models.ConnectionProfile, error)

	// ClearProfiles очищает все профили
	ClearProfiles() error

	// UpdateLastUsed обновляет время последнего использования профиля
	UpdateLastUsed(serialNumber string) error
}
