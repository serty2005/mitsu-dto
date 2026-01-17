package controller

import (
	"fmt"
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/service/connection"
	"mitsuscanner/internal/ui/viewmodel"
	"strconv"
	"strings"
	"time"
)

// MainController управляет логикой главной вкладки приложения, связанной с подключением к ККТ.
type MainController struct {
	vm          *viewmodel.MainViewModel
	connService *connection.ConnectionService
	onUpdate    func()
}

// NewMainController создает новый экземпляр MainController с использованием Dependency Injection.
func NewMainController(vm *viewmodel.MainViewModel, connService *connection.ConnectionService) *MainController {
	return &MainController{
		vm:          vm,
		connService: connService,
	}
}

// Initialize подготавливает начальные данные (вызывать из View при старте)
func (c *MainController) Initialize() {
	c.RefreshConnectionList()
}

// RefreshConnectionList обновляет список доступных подключений во ViewModel
func (c *MainController) RefreshConnectionList() {
	var items []string

	// 1. Профили
	profiles, _ := c.connService.LoadProfiles()
	// Сортировка профилей (обычно репозиторий уже отдает отсортированные по LastUsed, но можно перестраховаться)
	for _, p := range profiles {
		// Формируем строку отображения (логика перенесена из старого GUI)
		display := fmt.Sprintf("SN%s - ", p.SerialNumber)
		if p.ConnectionType == 0 {
			display += fmt.Sprintf("%s", p.ComName)
		} else {
			display += fmt.Sprintf("%s:%d", p.IPAddress, p.TCPPort)
		}

		fw := p.FirmwareVer
		if fw == "" {
			fw = "—"
		} else {
			fw = "v" + fw
		}
		display += fmt.Sprintf(" - %s", fw)

		items = append(items, display)
	}

	// 2. COM-порты (исключая те, что уже есть в профилях)
	systemPorts, _ := c.connService.GetSystemPorts()
	for _, port := range systemPorts {
		if !isPortInProfiles(port, profiles) {
			items = append(items, port)
		}
	}

	// 3. Пункт поиска
	items = append(items, "Поиск в сети / Ввести IP...")

	// Обновляем VM
	c.vm.ConnectionList = items

	// Если список не пуст и ничего не выбрано, выбираем первый элемент
	if c.vm.ConnectionString == "" && len(items) > 0 {
		c.vm.ConnectionString = items[0]
	}

	c.vm.UpdateUIState()
	c.notifyUpdate()
}

// isPortInProfiles проверяет, используется ли порт в сохраненных профилях
func isPortInProfiles(port string, profiles []*models.ConnectionProfile) bool {
	for _, p := range profiles {
		if p.ConnectionType == 0 && p.ComName == port {
			return true
		}
	}
	return false
}

// В методе ClearProfiles добавь вызов обновления списка:
func (c *MainController) ClearProfiles() error {
	if err := c.connService.ClearProfiles(); err != nil {
		return err
	}
	c.RefreshConnectionList() // <--- Добавлено
	return nil
}

// ViewModel возвращает ViewModel для главной вкладки.
func (c *MainController) ViewModel() *viewmodel.MainViewModel {
	return c.vm
}

// SetOnUpdate устанавливает callback для обновления пользовательского интерфейса.
func (c *MainController) SetOnUpdate(callback func()) {
	c.onUpdate = callback
}

// Connect устанавливает соединение с ККТ.
func (c *MainController) Connect() error {
	input := c.vm.ConnectionString

	// ЗАЩИТА: Если пользователь выбрал "Поиск...", но каким-то образом вызвался Connect
	if input == "Поиск в сети / Ввести IP..." {
		return c.SearchDevice()
	}

	var profile models.ConnectionProfile

	// СЦЕНАРИЙ А: Выбран профиль (строка начинается с SN...)
	if strings.HasPrefix(input, "SN") {
		sn := extractSN(input)
		foundProfile, err := c.connService.FindProfile(sn)
		if err == nil && foundProfile != nil {
			profile = *foundProfile
		} else {
			// Если профиль не найден, пробуем парсить строку
			host, port, isCom := parseConnectionString(input)
			if isCom {
				profile.ConnectionType = 0
				profile.ComName = host
				profile.BaudRate = port
			} else {
				profile.ConnectionType = 6
				profile.IPAddress = host
				profile.TCPPort = port
			}
		}
	} else {
		// СЦЕНАРИЙ Б: Ручной ввод
		host, port, isCom := parseConnectionString(input)
		if isCom {
			profile.ConnectionType = 0
			profile.ComName = host
			profile.BaudRate = port
		} else {
			profile.ConnectionType = 6
			profile.IPAddress = host
			profile.TCPPort = port
		}
	}

	// 3. Передаем настройки в драйвер!
	if err := c.connService.UpdateConnectionSettings(profile); err != nil {
		return err
	}

	// 4. Подключаемся
	if err := c.connService.Connect(); err != nil {
		return err
	}

	c.vm.IsConnected = true
	c.vm.UpdateUIState()

	// Загружаем инфо после подключения
	if err := c.LoadFiscalInfo(); err != nil {
		return err
	}

	// Создаем или обновляем профиль подключения (как в старом gui/tab_main.go)
	// Получаем модель, серийный номер и версию прошивки для профиля
	model, err := c.connService.GetModel()
	if err != nil {
		return err
	}

	_, serial, fwVersion, err := c.connService.GetVersion()
	if err != nil {
		return err
	}

	// Обновляем профиль с данными из ККТ
	profile.SerialNumber = serial
	profile.FirmwareVer = fwVersion
	profile.ModelName = model
	profile.LastUsed = time.Now()

	// Сохраняем профиль
	if err := c.connService.SaveProfile(&profile); err != nil {
		return err
	}

	// Обновляем список подключений, чтобы показать новый/обновленный профиль
	c.RefreshConnectionList()

	c.notifyUpdate()

	return nil
}

// Disconnect разрывает соединение с ККТ.
func (c *MainController) Disconnect() error {
	if err := c.connService.Disconnect(); err != nil {
		return err
	}

	c.vm.IsConnected = false
	c.vm.UpdateUIState()
	c.notifyUpdate()

	return nil
}

// SearchDevice ищет доступные устройства для подключения.
func (c *MainController) SearchDevice() error {
	// Логика поиска устройства будет реализована позже
	// Для текущей версии устанавливаем пустую строку подключения
	c.vm.ConnectionString = "Поиск в сети / Ввести IP..."
	c.vm.UpdateUIState()
	c.notifyUpdate()

	return nil
}

// LoadFiscalInfo загружает информацию о ККТ и обновляет ViewModel.
func (c *MainController) LoadFiscalInfo() error {
	fiscalInfo, err := c.connService.GetFiscalInfo()
	if err != nil {
		return err
	}

	c.vm.SerialNumber = fiscalInfo.SerialNumber

	// Загружаем модель ККТ
	model, err := c.connService.GetModel()
	if err == nil {
		c.vm.ModelName = model
	}

	// Загружаем статус смены для получения информации о неотправленных чеках
	shiftStatus, err := c.connService.GetShiftStatus()
	if err == nil {
		c.vm.UnsentDocsCount = shiftStatus.Ofd.Count
	}

	c.notifyUpdate()

	return nil
}

// LoadProfiles загружает профили подключения.
func (c *MainController) LoadProfiles() ([]*models.ConnectionProfile, error) {
	return c.connService.LoadProfiles()
}

// SaveProfile сохраняет или обновляет профиль подключения.
func (c *MainController) SaveProfile(profile *models.ConnectionProfile) error {
	return c.connService.SaveProfile(profile)
}

// DeleteProfile удаляет профиль подключения.
func (c *MainController) DeleteProfile(serialNumber string) error {
	return c.connService.DeleteProfile(serialNumber)
}

// notifyUpdate вызывает callback для обновления UI, если он установлен.
func (c *MainController) notifyUpdate() {
	if c.onUpdate != nil {
		c.onUpdate()
	}
}

func createProfileStruct(host string, port int, isCom bool) models.ConnectionProfile {
	p := models.ConnectionProfile{
		LastUsed: time.Now(),
	}
	if isCom {
		p.ConnectionType = 0
		p.ComName = host
		p.BaudRate = port
	} else {
		p.ConnectionType = 6
		p.IPAddress = host
		p.TCPPort = port
	}
	return p
}

func parseConnectionString(input string) (host string, port int, isCom bool) {
	input = strings.TrimSpace(input)

	// Очистка от суффикса версии (если это строка профиля)
	if idx := strings.Index(input, " - v"); idx != -1 {
		input = strings.TrimSpace(input[:idx])
	}
	// Очистка от старого формата (если просто версия в конце)
	if idx := strings.Index(input, " v"); idx != -1 {
		input = strings.TrimSpace(input[:idx])
	}

	isCom = strings.HasPrefix(strings.ToUpper(input), "COM")

	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		host = parts[0]
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		}
	} else {
		host = input
	}

	if port == 0 {
		if isCom {
			port = 115200
		} else {
			port = 8200
		}
	}
	return host, port, isCom
}

func extractSN(s string) string {
	// Format: SN123456 - ...
	parts := strings.Split(s, " - ")
	if len(parts) > 0 {
		return strings.TrimPrefix(parts[0], "SN")
	}
	return ""
}
