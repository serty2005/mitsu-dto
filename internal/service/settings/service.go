package settings

import (
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/domain/ports"
)

// Change представляет одно атомарное (или групповое) изменение настроек
type Change struct {
	ID          string                     // Уникальный ID поля (для подсветки в GUI)
	Description string                     // Человекочитаемое описание изменения
	OldValue    interface{}                // Значение "Было" (для отображения)
	NewValue    interface{}                // Значение "Стало" (для отображения)
	Priority    models.Priority            // Приоритет выполнения
	ApplyFunc   func(d ports.Driver) error // ApplyFunc - замыкание, которое применит это изменение к драйверу
}

// SettingsService отвечает за логику работы с настройками ККТ
type SettingsService struct {
	driver ports.Driver
}

// NewSettingsService создает новый экземпляр SettingsService
func NewSettingsService(driver ports.Driver) *SettingsService {
	return &SettingsService{
		driver: driver,
	}
}

// ReadAllSettings читает все настройки ККТ и возвращает снапшот
func (s *SettingsService) ReadAllSettings() (*models.SettingsSnapshot, error) {
	snap := models.NewSettingsSnapshot()

	// Читаем основные настройки
	ofdSettings, err := s.driver.GetOfdSettings()
	if err == nil {
		snap.Ofd = *ofdSettings
	}

	oismSettings, err := s.driver.GetOismSettings()
	if err == nil {
		snap.Oism = *oismSettings
	}

	lanSettings, err := s.driver.GetLanSettings()
	if err == nil {
		snap.Lan = *lanSettings
	}

	timezone, err := s.driver.GetTimezone()
	if err == nil {
		snap.Timezone = timezone
	}

	printerSettings, err := s.driver.GetPrinterSettings()
	if err == nil {
		snap.Printer = *printerSettings
	}

	drawerSettings, err := s.driver.GetMoneyDrawerSettings()
	if err == nil {
		snap.Drawer = *drawerSettings
	}

	options, err := s.driver.GetOptions()
	if err == nil {
		snap.Options = *options
	}

	// Читаем клише
	for i := 1; i <= 4; i++ {
		lines, err := s.driver.GetHeader(i)
		if err == nil {
			snap.Cliches[i] = lines
		}
	}

	return snap, nil
}

// ApplyChanges применяет список изменений к ККТ в правильном порядке
func (s *SettingsService) ApplyChanges(changes []Change) error {
	// Группируем изменения по приоритету
	var simpleChanges, clicheChanges, networkChanges []Change

	for _, ch := range changes {
		switch ch.Priority {
		case models.PriorityNormal:
			simpleChanges = append(simpleChanges, ch)
		case models.PriorityCliche:
			clicheChanges = append(clicheChanges, ch)
		case models.PriorityNetwork:
			networkChanges = append(networkChanges, ch)
		}
	}

	// Применяем изменения в правильном порядке: простые → клише → сетевые
	for _, ch := range simpleChanges {
		if err := ch.ApplyFunc(s.driver); err != nil {
			return err
		}
	}

	for _, ch := range clicheChanges {
		if err := ch.ApplyFunc(s.driver); err != nil {
			return err
		}
	}

	for _, ch := range networkChanges {
		if err := ch.ApplyFunc(s.driver); err != nil {
			return err
		}
	}

	return nil
}

// CompareSettings сравнивает два снапшота настроек
func (s *SettingsService) CompareSettings(initial, current *models.SettingsSnapshot) []Change {
	return Compare(initial, current)
}

// GetRegistrationData возвращает данные о регистрации ККТ
func (s *SettingsService) GetRegistrationData() (*models.RegData, error) {
	return s.driver.GetRegistrationData()
}

// GetTaxRates возвращает налоговые ставки
func (s *SettingsService) GetTaxRates() (*models.TaxRates, error) {
	return s.driver.GetTaxRates()
}

// TechReset выполняет технологическое обнуление ККТ
func (s *SettingsService) TechReset() error {
	return s.driver.TechReset()
}

// PrintXReport печатает X-отчет
func (s *SettingsService) PrintXReport() error {
	return s.driver.PrintXReport()
}

// FeedAndCut выполняет прогон бумаги и отрезку
func (s *SettingsService) FeedAndCut() error {
	if err := s.driver.Feed(5); err != nil {
		return err
	}
	return s.driver.Cut()
}

// OpenDrawer открывает денежный ящик
func (s *SettingsService) OpenDrawer() error {
	return s.driver.DeviceJob(2)
}

// ResetMGM сбрасывает МГМ
func (s *SettingsService) ResetMGM() error {
	return s.driver.ResetMGM()
}
