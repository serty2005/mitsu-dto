package driver

import (
	"fmt"
	"strings"
)

// RegistrationRequest содержит параметры для регистрации/перерегистрации ККТ.
// Поля соответствуют атрибутам и тегам команды <REG> (стр. 23 документации).

// Register выполняет первичную регистрацию ККТ (5.1).
func (d *mitsuDriver) Register(req RegistrationRequest) error {
	req.IsReregistration = false
	req.Base = "0" // Для первичной регистрации BASE всегда '0'
	return d.performRegistration(req)
}

// Reregister выполняет перерегистрацию ККТ (5.2).
// reasons - список кодов причин (см. стр 12, например: 1 - замена ФН, 3 - смена реквизитов).
func (d *mitsuDriver) Reregister(req RegistrationRequest, reasons []int) error {
	req.IsReregistration = true
	// Формируем строку BASE="1,3,..."
	var strReasons []string
	for _, r := range reasons {
		strReasons = append(strReasons, fmt.Sprintf("%d", r))
	}
	req.Base = strings.Join(strReasons, ",")
	if req.Base == "" {
		return fmt.Errorf("не указаны причины перерегистрации")
	}

	return d.performRegistration(req)
}

// performRegistration формирует XML команду <REG> и отправляет её.
func (d *mitsuDriver) performRegistration(req RegistrationRequest) error {
	// Сборка атрибутов
	// Обязательные атрибуты согласно стр. 23
	attrs := fmt.Sprintf("BASE='%s' T1062='%s'", req.Base, req.TaxSystems)

	// Добавляем опциональные атрибуты (флаги режимов)
	if req.InternetCalc {
		attrs += " T1108='1'"
	}
	if req.Service {
		attrs += " T1109='1'"
	}
	if req.BSO {
		attrs += " T1110='1'"
	}
	if req.Lottery {
		attrs += " T1126='1'"
	}
	if req.Gambling {
		attrs += " T1193='1'"
	}
	if req.Excise {
		attrs += " T1207='1'"
	}
	if req.Marking {
		attrs += " MARK='1'"
	}
	if req.PawnShop {
		attrs += " PAWN='1'"
	}
	if req.Insurance {
		attrs += " INS='1'"
	}
	if req.Catering {
		attrs += " DINE='1'"
	}
	if req.Wholesale {
		attrs += " OPT='1'"
	}
	if req.Vending {
		attrs += " VEND='1'"
	}
	if req.AutomatMode {
		attrs += " T1001='1'"
	}
	if req.AutonomousMode {
		attrs += " T1002='1'"
	}
	if req.Encryption {
		attrs += " T1056='1'"
	}
	if req.PrinterAutomat {
		attrs += " T1221='1'"
	}

	// Версия ФФД
	if req.FfdVer != "" {
		attrs += fmt.Sprintf(" T1209='%s'", req.FfdVer)
	}
	// Номер автомата
	if req.AutomatNumber != "" {
		attrs += fmt.Sprintf(" T1036='%s'", req.AutomatNumber)
	}
	// Базовая СНО
	if req.TaxSystemBase != "" {
		attrs += fmt.Sprintf(" T1062_Base='%s'", req.TaxSystemBase)
	}

	// Сборка вложенных тегов
	tags := ""
	tags += fmt.Sprintf("<T1048>%s</T1048>", escapeXML(req.OrgName))
	tags += fmt.Sprintf("<T1009>%s</T1009>", escapeXML(req.Address))
	tags += fmt.Sprintf("<T1187>%s</T1187>", escapeXML(req.Place))
	tags += fmt.Sprintf("<T1046>%s</T1046>", escapeXML(req.OfdName))
	// ИНН ОФД
	tags += fmt.Sprintf("<T1017>%s</T1017>", req.OfdInn)

	// ИСПРАВЛЕНИЕ: Передаем ИНН и РНМ всегда, даже при перерегистрации.
	// Ошибка 103 (TAG 1037) указывает на то, что устройство требует эти данные.
	tags += fmt.Sprintf("<T1018>%s</T1018>", req.Inn)
	tags += fmt.Sprintf("<T1037>%s</T1037>", req.RNM)

	tags += fmt.Sprintf("<T1060>%s</T1060>", escapeXML(req.FnsSite))
	tags += fmt.Sprintf("<T1117>%s</T1117>", escapeXML(req.SenderEmail))

	// Номер автомата также дублируется в тегах в примере
	if req.AutomatNumber != "" {
		tags += fmt.Sprintf("<T1036>%s</T1036>", escapeXML(req.AutomatNumber))
	}

	// Итоговая команда
	xmlCmd := fmt.Sprintf("<REG %s>%s</REG>", attrs, tags)

	_, err := d.sendCommand(xmlCmd)
	return err
}

// CloseFiscalArchive закрывает фискальный режим (5.4).
func (d *mitsuDriver) CloseFiscalArchive() error {
	// Для закрытия ФН нужно отправить MAKE FISCAL='CLOSE'
	// И передать данные для отчета о закрытии (адрес, место, ОЗФН)
	cmd := "<MAKE FISCAL='CLOSE'/>"
	_, err := d.sendCommand(cmd)
	return err
}
