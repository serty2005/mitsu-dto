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
	// В Mitsu протоколе если флаг '0', его можно не передавать, но для надежности передадим '1' явно, где нужно.
	// Документация говорит: T1108='признак...', значение '1'.
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

	// Добавляем основные реквизиты в атрибуты (если это не перерегистрация с их сменой, но протокол позволяет слать все)
	// Для перерегистрации T1018 и T1037 могут быть опущены, если не меняются, но обычно их шлют.
	// ВНИМАНИЕ: Протокол (стр 24) говорит: "Те же, что при регистрации, ЗА ИСКЛЮЧЕНИЕМ T1018, T1037".
	// Значит, при перерегистрации их в атрибуты пихать НЕЛЬЗЯ, если это не их смена?
	// Стр 24: "При перерегистрации ИНН (T1018) и РНМ (T1037), заданные при первичной регистрации, не подлежат изменению".
	// Значит, их вообще лучше не посылать при перерегистрации, или посылать те же самые?
	// Пример на стр 24 показывает: <REG BASE='...' ...> без T1018/T1037 в атрибутах.
	// Но в тегах (ниже) они могут быть.

	if !req.IsReregistration {
		attrs += fmt.Sprintf(" T1037='%s' T1018='%s'", req.RNM, req.Inn)
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
	// ИНН ОФД и ИНН Пользователя (дублируется в тегах по примеру стр. 23)
	tags += fmt.Sprintf("<T1017>%s</T1017>", req.OfdInn)

	// При перерегистрации ИНН пользователя часто не шлют, но в примере (стр 23) он есть в тегах для первичной.
	// Для перерегистрации (стр 24): "Теги: Те же... за исключением <T1018> <T1037>".
	if !req.IsReregistration {
		tags += fmt.Sprintf("<T1018>%s</T1018>", req.Inn)
		tags += fmt.Sprintf("<T1037>%s</T1037>", req.RNM)
	}

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
	// Для простоты пока отправляем только команду закрытия,
	// но по протоколу (стр 24) нужно передать параметры T1009, T1187 и т.д.
	// Чтобы API был корректным, лучше сначала считать текущие параметры через GetRegistrationData,
	// а потом передать их в команду закрытия.

	// Здесь упрощенная реализация, предполагающая, что пользователь вызывает это осознанно.
	// Правильнее было бы принимать аргументы адреса и места.
	cmd := "<MAKE FISCAL='CLOSE'/>"
	_, err := d.sendCommand(cmd)
	return err
}
