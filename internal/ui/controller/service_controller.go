package controller

import (
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/service/settings"
	timeService "mitsuscanner/internal/service/time"
	"mitsuscanner/internal/ui/view/dialogs"
	"mitsuscanner/internal/ui/viewmodel"
	"net"
	"strconv"

	"github.com/lxn/walk"
)

// ServiceController управляет логикой вкладки "Сервис" приложения, связанной с настройками и синхронизацией времени ККТ.
type ServiceController struct {
	vm              *viewmodel.ServiceViewModel
	settingsService *settings.SettingsService
	timeService     *timeService.TimeService
	onUpdate        func()
}

// ViewModel возвращает ссылку на модель данных для UI
func (c *ServiceController) ViewModel() *viewmodel.ServiceViewModel {
	return c.vm
}

// NewServiceController создает новый экземпляр ServiceController с использованием Dependency Injection.
func NewServiceController(vm *viewmodel.ServiceViewModel, settingsService *settings.SettingsService, timeService *timeService.TimeService) *ServiceController {
	return &ServiceController{
		vm:              vm,
		settingsService: settingsService,
		timeService:     timeService,
	}
}

// SetOnUpdate устанавливает callback для обновления пользовательского интерфейса.
func (c *ServiceController) SetOnUpdate(callback func()) {
	c.onUpdate = callback
}

// ReadSettings читает все настройки ККТ и обновляет ViewModel.
func (c *ServiceController) ReadSettings() error {
	snapshot, err := c.settingsService.ReadAllSettings()
	if err != nil {
		return err
	}

	c.mapDomainToViewModel(snapshot)
	c.notifyUpdate()

	return nil
}

// WriteSettings сохраняет настройки из ViewModel в ККТ.
func (c *ServiceController) WriteSettings(owner interface{}) error {
	snapshot := c.mapViewModelToDomain()
	initial, err := c.settingsService.ReadAllSettings()
	if err != nil {
		return err
	}

	changes := c.settingsService.CompareSettings(initial, snapshot)

	// Преобразуем в доменные модели для диалога
	var domainChanges []models.Change
	for _, ch := range changes {
		domainChanges = append(domainChanges, models.Change{
			ID:          ch.ID,
			Description: ch.Description,
			OldValue:    ch.OldValue,
			NewValue:    ch.NewValue,
			Priority:    ch.Priority,
		})
	}

	// Показываем диалог с изменениями
	if !dialogs.ShowDiffDialog(owner.(walk.Form), domainChanges) {
		return nil // Пользователь отменил изменения
	}

	if err := c.settingsService.ApplyChanges(changes); err != nil {
		return err
	}

	return nil
}

// SyncTime синхронизирует время ККТ с заданным временем из ViewModel.
func (c *ServiceController) SyncTime() error {
	t, err := c.timeService.ParseTime(c.vm.TargetTimeStr)
	if err != nil {
		return err
	}

	if err := c.timeService.SyncTime(t); err != nil {
		return err
	}

	// Обновляем отображение времени ККТ
	if kktTime, err := c.timeService.GetKKTTime(); err == nil {
		c.vm.KktTimeStr = c.timeService.FormatTime(kktTime)
	}

	c.notifyUpdate()

	return nil
}

// SyncWithSystemTime синхронизирует время ККТ с временем системы.
func (c *ServiceController) SyncWithSystemTime() error {
	if err := c.timeService.SyncWithSystemTime(); err != nil {
		return err
	}

	if kktTime, err := c.timeService.GetKKTTime(); err == nil {
		c.vm.KktTimeStr = c.timeService.FormatTime(kktTime)
	}

	c.vm.TargetTimeStr = c.timeService.FormatTime(c.timeService.GetCurrentTime())
	c.notifyUpdate()

	return nil
}

// UpdateKKTTime обновляет отображение времени ККТ.
func (c *ServiceController) UpdateKKTTime() error {
	if kktTime, err := c.timeService.GetKKTTime(); err == nil {
		c.vm.KktTimeStr = c.timeService.FormatTime(kktTime)
		c.notifyUpdate()
	}

	return nil
}

// mapDomainToViewModel преобразует SettingsSnapshot из Domain слоя в ServiceViewModel для UI.
func (c *ServiceController) mapDomainToViewModel(snapshot *models.SettingsSnapshot) {
	// Синхронизация времени
	c.vm.AutoSyncPC = true // По умолчанию включено
	c.vm.TargetTimeStr = c.timeService.FormatTime(c.timeService.GetCurrentTime())
	if kktTime, err := c.timeService.GetKKTTime(); err == nil {
		c.vm.KktTimeStr = c.timeService.FormatTime(kktTime)
	}

	// Настройки ОФД
	c.vm.OfdString = snapshot.Ofd.Addr
	c.vm.OfdClient = snapshot.Ofd.Client
	c.vm.TimerFN = snapshot.Ofd.TimerFN
	c.vm.TimerOFD = snapshot.Ofd.TimerOFD

	// Настройки OISM
	c.vm.OismString = snapshot.Oism.Addr

	// Настройки LAN
	c.vm.LanAddr = snapshot.Lan.Addr
	c.vm.LanPort = snapshot.Lan.Port
	c.vm.LanMask = snapshot.Lan.Mask
	c.vm.LanDns = snapshot.Lan.Dns
	c.vm.LanGw = snapshot.Lan.Gw

	// Настройки принтера
	c.vm.PrintModel = snapshot.Printer.Model
	c.vm.PrintBaud = strconv.Itoa(snapshot.Printer.BaudRate)
	c.vm.PrintPaper = strconv.Itoa(snapshot.Printer.Paper)
	c.vm.PrintFont = strconv.Itoa(snapshot.Printer.Font)

	// Настройки денежного ящика
	c.vm.DrawerPin = snapshot.Drawer.Pin
	c.vm.DrawerRise = snapshot.Drawer.Rise
	c.vm.DrawerFall = snapshot.Drawer.Fall

	// Опции устройства
	c.vm.OptTimezone = strconv.Itoa(snapshot.Timezone)
	c.vm.OptCut = snapshot.Options.B3 == 1
	c.vm.OptAutoTest = snapshot.Options.B4 == 1
	c.vm.OptNearEnd = snapshot.Options.B6 == 1
	c.vm.OptTextQR = snapshot.Options.B7 == 1
	c.vm.OptCountInCheck = snapshot.Options.B8 == 1
	c.vm.OptQRPos = strconv.Itoa(snapshot.Options.B1)
	c.vm.OptRounding = strconv.Itoa(snapshot.Options.B2)
	c.vm.OptDrawerTrig = strconv.Itoa(snapshot.Options.B5)
	c.vm.OptB9_BaseTax = strconv.Itoa(snapshot.Options.B9 & 0x0F) // Маска для получения базовой СНО (без флага полного X-отчета)
	c.vm.OptB9_FullX = (snapshot.Options.B9 & 0x80) == 0x80

	// Клише
	for clicheType, lines := range snapshot.Cliches {
		if clicheType >= 1 && clicheType <= 4 {
			// Ограничиваем количество строк клише до 10
			for i := 0; i < 10 && i < len(lines); i++ {
				if i < len(c.vm.ClicheItems) {
					c.vm.ClicheItems[i].Line.Text = lines[i].Text
					c.vm.ClicheItems[i].Line.Format = lines[i].Format
				}
			}
		}
	}
}

// mapViewModelToDomain преобразует ServiceViewModel из UI в SettingsSnapshot для Domain слоя.
func (c *ServiceController) mapViewModelToDomain() *models.SettingsSnapshot {
	snapshot := models.NewSettingsSnapshot()

	// Настройки ОФД
	ofdAddr, ofdPort := splitHostPort(c.vm.OfdString, 443)
	snapshot.Ofd = models.OfdSettings{
		Addr:     ofdAddr,
		Port:     ofdPort,
		Client:   c.vm.OfdClient,
		TimerFN:  c.vm.TimerFN,
		TimerOFD: c.vm.TimerOFD,
	}

	// Настройки OISM
	oismAddr, oismPort := splitHostPort(c.vm.OismString, 80)
	snapshot.Oism = models.OismSettings{
		Addr: oismAddr,
		Port: oismPort,
	}

	// Настройки LAN
	snapshot.Lan = models.LanSettings{
		Addr: c.vm.LanAddr,
		Port: c.vm.LanPort,
		Mask: c.vm.LanMask,
		Dns:  c.vm.LanDns,
		Gw:   c.vm.LanGw,
	}

	// Настройки принтера
	printBaud, _ := strconv.Atoi(c.vm.PrintBaud)
	printPaper, _ := strconv.Atoi(c.vm.PrintPaper)
	printFont, _ := strconv.Atoi(c.vm.PrintFont)
	snapshot.Printer = models.PrinterSettings{
		Model:    c.vm.PrintModel,
		BaudRate: printBaud,
		Paper:    printPaper,
		Font:     printFont,
	}

	// Настройки денежного ящика
	snapshot.Drawer = models.DrawerSettings{
		Pin:  c.vm.DrawerPin,
		Rise: c.vm.DrawerRise,
		Fall: c.vm.DrawerFall,
	}

	// Часовой пояс
	timezone, _ := strconv.Atoi(c.vm.OptTimezone)
	snapshot.Timezone = timezone

	// Опции устройства
	qrPos, _ := strconv.Atoi(c.vm.OptQRPos)
	rounding, _ := strconv.Atoi(c.vm.OptRounding)
	drawerTrig, _ := strconv.Atoi(c.vm.OptDrawerTrig)
	b9BaseTax, _ := strconv.Atoi(c.vm.OptB9_BaseTax)

	// Обработка опции B9 (базовая СНО + флаг полного Х-отчета)
	b9Value := b9BaseTax & 0x0F // Очищаем старший бит
	if c.vm.OptB9_FullX {
		b9Value |= 0x80
	}

	snapshot.Options = models.DeviceOptions{
		B1: qrPos,
		B2: rounding,
		B3: boolToInt(c.vm.OptCut),
		B4: boolToInt(c.vm.OptAutoTest),
		B5: drawerTrig,
		B6: boolToInt(c.vm.OptNearEnd),
		B7: boolToInt(c.vm.OptTextQR),
		B8: boolToInt(c.vm.OptCountInCheck),
		B9: b9Value,
	}

	// Клише - копируем все типы клише
	for clicheType := 1; clicheType <= 4; clicheType++ {
		// Для каждого типа клише сохраняем все строки
		snapshot.Cliches[clicheType] = make([]models.ClicheLineData, len(c.vm.ClicheItems))
		for i, item := range c.vm.ClicheItems {
			snapshot.Cliches[clicheType][i] = models.ClicheLineData{
				Text:   item.Line.Text,
				Format: item.Line.Format,
			}
		}
	}

	return snapshot
}

// splitHostPort разбивает строку формата "host:port" на хост и порт.
func splitHostPort(addr string, defaultPort int) (string, int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// Если не удалось разобрать, возвращаем исходную строку и порт по умолчанию
		return addr, defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return addr, defaultPort
	}
	return host, port
}

// boolToInt преобразует булевое значение в целое (1 для true, 0 для false).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// notifyUpdate вызывает callback для обновления UI, если он установлен.
func (c *ServiceController) notifyUpdate() {
	if c.onUpdate != nil {
		c.onUpdate()
	}
}
