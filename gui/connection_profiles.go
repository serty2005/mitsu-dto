package gui

import (
	"mitsuscanner/internal/models"
	"mitsuscanner/internal/storage"
)

// LoadProfiles загружает профили из JSON-файла
func LoadProfiles(store *storage.ProfilesStore) error {
	return store.LoadProfiles()
}

// SaveProfiles сохраняет профили в JSON-файл
func SaveProfiles(store *storage.ProfilesStore) error {
	return store.SaveProfiles()
}

// GetProfiles возвращает все профили
func GetProfiles(store *storage.ProfilesStore) []*models.ConnectionProfile {
	return store.GetProfiles()
}

// UpsertProfile добавляет или обновляет профиль
func UpsertProfile(store *storage.ProfilesStore, profile *models.ConnectionProfile) error {
	return store.UpsertProfile(profile)
}

// DeleteProfile удаляет профиль
func DeleteProfile(store *storage.ProfilesStore, serialNumber string) error {
	return store.DeleteProfile(serialNumber)
}

// FindProfile находит профиль
func FindProfile(store *storage.ProfilesStore, serialNumber string) *models.ConnectionProfile {
	return store.FindProfile(serialNumber)
}

// ClearProfiles очищает все профили
func ClearProfiles(store *storage.ProfilesStore) error {
	return store.ClearProfiles()
}

// UpdateLastUsed обновляет время последнего использования
func UpdateLastUsed(store *storage.ProfilesStore, serialNumber string) error {
	return store.UpdateLastUsed(serialNumber)
}
