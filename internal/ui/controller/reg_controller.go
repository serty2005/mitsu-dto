package controller

import (
	"mitsuscanner/internal/domain/models"
	"mitsuscanner/internal/service/registration"
	"mitsuscanner/internal/ui/viewmodel"
	"regexp"
	"strconv"
	"strings"
)

// RegistrationController управляет логикой вкладки "Регистрация" приложения.
type RegistrationController struct {
	vm              *viewmodel.RegistrationViewModel
	registrationSvc *registration.RegistrationService
	onUpdate        func()
}

// NewRegistrationController создает новый экземпляр RegistrationController с использованием Dependency Injection.
func NewRegistrationController(vm *viewmodel.RegistrationViewModel, registrationSvc *registration.RegistrationService) *RegistrationController {
	return &RegistrationController{
		vm:              vm,
		registrationSvc: registrationSvc,
	}
}

// ViewModel возвращает ссылку на модель данных для UI.
func (c *RegistrationController) ViewModel() *viewmodel.RegistrationViewModel {
	return c.vm
}

// SetOnUpdate устанавливает callback для обновления пользовательского интерфейса.
func (c *RegistrationController) SetOnUpdate(callback func()) {
	c.onUpdate = callback
}

// ReadFromDevice читает данные регистрации из ККТ и обновляет ViewModel.
func (c *RegistrationController) ReadFromDevice() error {
	regData, err := c.registrationSvc.GetFullRegistrationData()
	if err != nil {
		return err
	}

	c.mapDomainToViewModel(regData)
	c.notifyUpdate()

	return nil
}

// CalculateRNM вычисляет РНМ по алгоритму CRC16-CCITT на основе данных из ViewModel.
func (c *RegistrationController) CalculateRNM(orderNum, inn, serial string) error {
	rnm, err := c.registrationSvc.CalculateRNM(orderNum, inn, serial)
	if err != nil {
		return err
	}

	c.vm.RNM = rnm
	c.notifyUpdate()

	return nil
}

// Register выполняет регистрацию ККТ с данными из ViewModel.
func (c *RegistrationController) Register() (*models.RegResponse, error) {
	req, err := c.mapViewModelToDomain(false)
	if err != nil {
		return nil, err
	}

	resp, err := c.registrationSvc.Register(*req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Reregister выполняет перерегистрацию ККТ с данными из ViewModel и причинами перерегистрации.
func (c *RegistrationController) Reregister() (*models.RegResponse, error) {
	if c.vm.Reasons == "" {
		return nil, &ValidationError{Message: "Не выбраны причины перерегистрации"}
	}

	reasons, err := parseReasons(c.vm.Reasons)
	if err != nil {
		return nil, err
	}

	req, err := c.mapViewModelToDomain(true)
	if err != nil {
		return nil, err
	}

	resp, err := c.registrationSvc.Reregister(*req, reasons)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CloseFn закрывает фискальный архив и возвращает данные для отчета.
func (c *RegistrationController) CloseFn() (*models.ReportFnCloseData, error) {
	return c.registrationSvc.CloseFiscalArchive()
}

// RefreshFnInfo обновляет информацию о ФН.
func (c *RegistrationController) RefreshFnInfo() error {
	// TODO: Реализовать отдельный метод в сервисе для получения только статуса ФН
	// Для сейчас повторяем логику ReadFromDevice
	return c.ReadFromDevice()
}

// mapDomainToViewModel преобразует RegData из Domain слоя в RegistrationViewModel для UI.
func (c *RegistrationController) mapDomainToViewModel(data *models.RegData) {
	// Основные данные
	c.vm.RNM = data.RNM
	c.vm.INN = data.Inn
	c.vm.OrgName = data.OrgName
	c.vm.Address = data.Address
	c.vm.Place = data.Place
	c.vm.Email = data.EmailSender
	c.vm.Site = data.Site
	c.vm.OFDINN = data.OfdInn
	c.vm.OFDName = data.OfdName
	c.vm.FFD = data.FfdVer

	// Парсинг атрибутов режимов работы
	modeInt := int(data.ModeMask)
	extModeInt := int(data.ExtModeMask)

	c.vm.ModeEncryption = hasBit(modeInt, 0) // Шифрование
	c.vm.ModeAutonomous = hasBit(modeInt, 1) // Автономный режим
	c.vm.ModeService = hasBit(modeInt, 3)    // Услуги
	c.vm.ModeBSO = hasBit(modeInt, 4)        // БСО
	c.vm.ModeInternet = hasBit(modeInt, 5)   // Интернет
	c.vm.ModeCatering = hasBit(modeInt, 6)   // Общепит
	c.vm.ModeWholesale = hasBit(modeInt, 7)  // Оптовая торговля

	c.vm.ModeExcise = hasBit(extModeInt, 0)    // Подакцизные товары
	c.vm.ModeGambling = hasBit(extModeInt, 1)  // Азартные игры
	c.vm.ModeLottery = hasBit(extModeInt, 2)   // Лотереи
	c.vm.ModeAutomat = hasBit(extModeInt, 3)   // Принтер в автомате
	c.vm.ModeMarking = hasBit(extModeInt, 4)   // Маркированные товары
	c.vm.ModePawn = hasBit(extModeInt, 5)      // Ломбард
	c.vm.ModeInsurance = hasBit(extModeInt, 6) // Страхование
	c.vm.ModeVending = hasBit(extModeInt, 7)   // Вендинг

	// Парсинг систем налогообложения
	c.vm.TaxOSN = false
	c.vm.TaxUSN = false
	c.vm.TaxUSN_M = false
	c.vm.TaxENVD = false
	c.vm.TaxESHN = false
	c.vm.TaxPat = false

	taxParts := strings.Split(data.TaxSystems, ",")
	for _, t := range taxParts {
		trimmedT := strings.TrimSpace(t)
		switch trimmedT {
		case "0":
			c.vm.TaxOSN = true
		case "1":
			c.vm.TaxUSN = true
		case "2":
			c.vm.TaxUSN_M = true
		case "3":
			c.vm.TaxENVD = true
		case "4":
			c.vm.TaxESHN = true
		case "5":
			c.vm.TaxPat = true
		}
	}

	c.vm.TaxSystemBase = data.TaxBase

	// Информация о ФН
	c.vm.FnNumber = data.FnSerial
	c.vm.FnValidDate = data.FnEdition // TODO: Проверить соответствие полей
	c.decodeFnPhase(data.FnSerial)    // TODO: Проверить источник данных для фазы ФН
}

// mapViewModelToDomain преобразует RegistrationViewModel из UI в RegistrationRequest для Domain слоя.
func (c *RegistrationController) mapViewModelToDomain(isReregistration bool) (*models.RegistrationRequest, error) {
	// Валидация
	if !regexp.MustCompile(`^\d+$`).MatchString(c.vm.INN) || (len(c.vm.INN) != 10 && len(c.vm.INN) != 12) {
		return nil, &ValidationError{Message: "ИНН должен состоять только из цифр и иметь длину 10 или 12 символов"}
	}

	if strings.TrimSpace(c.vm.OrgName) == "" {
		return nil, &ValidationError{Message: "Поле 'Наименование' обязательно для заполнения"}
	}

	if strings.TrimSpace(c.vm.Address) == "" {
		return nil, &ValidationError{Message: "Поле 'Адрес расчетов' обязательно для заполнения"}
	}

	if strings.TrimSpace(c.vm.Place) == "" {
		return nil, &ValidationError{Message: "Поле 'Место расчетов' обязательно для заполнения"}
	}

	var taxCodes []string
	if c.vm.TaxOSN {
		taxCodes = append(taxCodes, "0")
	}
	if c.vm.TaxUSN {
		taxCodes = append(taxCodes, "1")
	}
	if c.vm.TaxUSN_M {
		taxCodes = append(taxCodes, "2")
	}
	if c.vm.TaxENVD {
		taxCodes = append(taxCodes, "3")
	}
	if c.vm.TaxESHN {
		taxCodes = append(taxCodes, "4")
	}
	if c.vm.TaxPat {
		taxCodes = append(taxCodes, "5")
	}

	return &models.RegistrationRequest{
		IsReregistration: isReregistration,
		RNM:              c.vm.RNM,
		Inn:              c.vm.INN,
		OrgName:          c.vm.OrgName,
		Address:          c.vm.Address,
		Place:            c.vm.Place,
		SenderEmail:      c.vm.Email,
		FnsSite:          c.vm.Site,
		FfdVer:           c.vm.FFD,
		OfdName:          c.vm.OFDName,
		OfdInn:           c.vm.OFDINN,
		AutonomousMode:   c.vm.ModeAutonomous,
		Encryption:       c.vm.ModeEncryption,
		Service:          c.vm.ModeService,
		InternetCalc:     c.vm.ModeInternet,
		BSO:              c.vm.ModeBSO,
		Gambling:         c.vm.ModeGambling,
		Lottery:          c.vm.ModeLottery,
		Excise:           c.vm.ModeExcise,
		Marking:          c.vm.ModeMarking,
		PawnShop:         c.vm.ModePawn,
		Insurance:        c.vm.ModeInsurance,
		Catering:         c.vm.ModeCatering,
		Wholesale:        c.vm.ModeWholesale,
		Vending:          c.vm.ModeVending,
		PrinterAutomat:   c.vm.ModeAutomat,
		TaxSystemBase:    c.vm.TaxSystemBase,
		TaxSystems:       strings.Join(taxCodes, ","),
	}, nil
}

// decodeFnPhase декодирует фазу ФН и устанавливает текст и цвет.
func (c *RegistrationController) decodeFnPhase(phase string) {
	phase = strings.TrimPrefix(strings.ToLower(phase), "0x")
	val, err := strconv.ParseInt(phase, 16, 32)
	if err != nil {
		c.vm.FnPhaseText = "Неизвестно"
		c.vm.FnPhaseColor = "#000000"
		return
	}

	switch val {
	case 0x01:
		c.vm.FnPhaseText = "Готов к фискализации"
		c.vm.FnPhaseColor = "#0000FF" // Синий
	case 0x03:
		c.vm.FnPhaseText = "Боевой режим"
		c.vm.FnPhaseColor = "#008000" // Зелёный
	case 0x07:
		c.vm.FnPhaseText = "ФН закрыт"
		c.vm.FnPhaseColor = "#FF0000" // Красный
	case 0x0F:
		c.vm.FnPhaseText = "ФР в архиве"
		c.vm.FnPhaseColor = "#0000FF" // Синий
	default:
		c.vm.FnPhaseText = "Неизвестная фаза"
		c.vm.FnPhaseColor = "#808080" // Серый
	}
}

// parseReasons парсит строку с причинами перерегистрации в массив целых чисел.
func parseReasons(reasonsStr string) ([]int, error) {
	var reasons []int
	parts := strings.Split(reasonsStr, ",")
	for _, p := range parts {
		code, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, &ValidationError{Message: "Некорректный формат причин перерегистрации"}
		}
		reasons = append(reasons, code)
	}
	return reasons, nil
}

// hasBit проверяет, установлен ли бит в целом числе.
func hasBit(value int, bit int) bool {
	return (value & (1 << bit)) != 0
}

// ValidationError представляет ошибку валидации данных.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// notifyUpdate вызывает callback для обновления UI, если он установлен.
func (c *RegistrationController) notifyUpdate() {
	if c.onUpdate != nil {
		c.onUpdate()
	}
}
