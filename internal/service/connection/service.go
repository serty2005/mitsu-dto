package connection

import (
	"sort"
	"time"

	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"

	"go.bug.st/serial"
)

// ConnectionService отвечает за логику подключения к ККТ и управлением профилями
type ConnectionService struct {
	driver ports.Driver
	repo   ports.ProfileRepository
}

// NewConnectionService создает новый экземпляр ConnectionService
func NewConnectionService(driver ports.Driver, repo ports.ProfileRepository) *ConnectionService {
	return &ConnectionService{
		driver: driver,
		repo:   repo,
	}
}

// GetSystemPorts возвращает список доступных в системе COM-портов
func (s *ConnectionService) GetSystemPorts() ([]string, error) {
	portsList, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	sort.Strings(portsList)
	return portsList, nil
}

// UpdateConnectionSettings обновляет настройки драйвера перед подключением
func (s *ConnectionService) UpdateConnectionSettings(profile models.ConnectionProfile) error {
	return s.driver.SetConnectionSettings(profile)
}

// Connect устанавливает соединение с ККТ
func (s *ConnectionService) Connect() error {
	return s.driver.Connect()
}

// Disconnect разрывает соединение с ККТ
func (s *ConnectionService) Disconnect() error {
	return s.driver.Disconnect()
}

// IsConnected проверяет, активно ли соединение
func (s *ConnectionService) IsConnected() bool {
	// В реальности это может быть реализовано через проверку состояния драйвера
	// Для текущей версии просто возвращаем true, так как драйвер хранит состояние
	return true
}

// GetFiscalInfo возвращает фискальную информацию о ККТ
func (s *ConnectionService) GetFiscalInfo() (*models.FiscalInfo, error) {
	return s.driver.GetFiscalInfo()
}

// GetModel возвращает модель ККТ
func (s *ConnectionService) GetModel() (string, error) {
	return s.driver.GetModel()
}

// GetVersion возвращает информацию о версии ККТ
func (s *ConnectionService) GetVersion() (string, string, string, error) {
	return s.driver.GetVersion()
}

// GetShiftStatus возвращает статус смены
func (s *ConnectionService) GetShiftStatus() (*models.ShiftStatus, error) {
	return s.driver.GetShiftStatus()
}

// LoadProfiles загружает все профили подключения
func (s *ConnectionService) LoadProfiles() ([]*models.ConnectionProfile, error) {
	return s.repo.LoadProfiles()
}

// SaveProfile сохраняет или обновляет профиль подключения
func (s *ConnectionService) SaveProfile(profile *models.ConnectionProfile) error {
	profile.LastUsed = time.Now()
	return s.repo.UpsertProfile(profile)
}

// DeleteProfile удаляет профиль по серийному номеру
func (s *ConnectionService) DeleteProfile(serialNumber string) error {
	return s.repo.DeleteProfile(serialNumber)
}

// FindProfile находит профиль по серийному номеру
func (s *ConnectionService) FindProfile(serialNumber string) (*models.ConnectionProfile, error) {
	return s.repo.FindProfile(serialNumber)
}

// ClearProfiles удаляет все профили
func (s *ConnectionService) ClearProfiles() error {
	return s.repo.ClearProfiles()
}

// UpdateLastUsed обновляет время последнего использования профиля
func (s *ConnectionService) UpdateLastUsed(serialNumber string) error {
	return s.repo.UpdateLastUsed(serialNumber)
}

// CreateProfileFromConnection создает профиль подключения на основе текущего соединения
func (s *ConnectionService) CreateProfileFromConnection(connectionType int, comName string, baudRate int, ipAddress string, tcpPort int) (*models.ConnectionProfile, error) {
	model, err := s.driver.GetModel()
	if err != nil {
		return nil, err
	}

	_, serial, fwVersion, err := s.driver.GetVersion()
	if err != nil {
		return nil, err
	}

	return &models.ConnectionProfile{
		SerialNumber:   serial,
		ConnectionType: connectionType,
		ComName:        comName,
		BaudRate:       baudRate,
		IPAddress:      ipAddress,
		TCPPort:        tcpPort,
		FirmwareVer:    fwVersion,
		ModelName:      model,
		LastUsed:       time.Now(),
	}, nil
}
