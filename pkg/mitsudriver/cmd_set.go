package mitsudriver

import (
	"fmt"
	"strings"
	"time"
)

// SetTimezone (3.35, добавлено в FW 1.2.18)
func (d *mitsuDriver) SetTimezone(value int) error {
	cmd := fmt.Sprintf("<SET TIMEZONE='%d'/>", value)
	_, err := d.sendCommand(cmd)
	return err
}

// --- Реализация методов SET ---

// SetDateTime (4.3)
// ВНИМАНИЕ! Если часы кассы установлены по ошибке на будущую дату/время и оформлен
// хотя бы один фискальный документ, вернуть назад часы невозможно.
// В этом случае потребуется замена ФН.
func (d *mitsuDriver) SetDateTime(t time.Time) error {
	dateStr := t.Format("2006-01-02")
	timeStr := t.Format("15:04:05")
	cmd := fmt.Sprintf("<SET DATE='%s' TIME='%s' />", dateStr, timeStr)

	// Ответ на эту команду содержит установленные дату и время, но если нет ошибки протокола,
	// считаем операцию успешной.
	_, err := d.sendCommand(cmd)
	return err
}

// SetCashier (4.4)
// name: Идентификатор (ФИО, должность), макс 64 символа.
// inn: ИНН кассира (необязательно).
// Необходимо устанавливать кассира перед открытием каждого чека.
func (d *mitsuDriver) SetCashier(name string, inn string) error {
	// Экранируем имя, так как оно может содержать кавычки и т.д.
	safeName := escapeXMLText(name)
	cmd := fmt.Sprintf("<SET CASHIER='%s' INN='%s'/>", safeName, inn)
	_, err := d.sendCommand(cmd)
	return err
}

// SetComSettings (4.5)
func (d *mitsuDriver) SetComSettings(speed int32) error {
	cmd := fmt.Sprintf("<SET COM='%d'/>", speed)
	_, err := d.sendCommand(cmd)
	return err
}

// Модель принтера (0 – нет принтера; 1 – Mitsu RP-809; 2 – Mitsu F80)

// SetPrinterSettings (4.6)
func (d *mitsuDriver) SetPrinterSettings(s PrinterSettings) error {
	cmd := fmt.Sprintf(
		"<SET PRINTER='%s' BAUDRATE='%d' PAPER='%d' FONT='%d'/>",
		s.Model, s.BaudRate, s.Paper, s.Font,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetMoneyDrawerSettings (4.7)
func (d *mitsuDriver) SetMoneyDrawerSettings(s DrawerSettings) error {
	cmd := fmt.Sprintf(
		"<SET CD='%d' RISE='%d' FALL='%d'/>",
		s.Pin, s.Rise, s.Fall,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetHeader (4.8)
// Установка клише и подвала.
// headerNum:
// 1 - клише №1, печатается в заголовке в самом верху документа
// 2 - клише №2, печатается после строк с наименованием пользователя и адреса расчетов
// 3 - клише №3, печатается внизу чека перед QR кодом и реквизитами
// 4 - клише №4, печатается в самом конце чека
// Суммарная длина всех строк каждого клише до 1000 символов
// lineNum: 0..9
// format: "xxxxxx" (6 цифр).
// 1 цифра: инверсия (0 – нет инверсии: черный текст на белом фоне, 1 – инверсия: белый текст на черном фоне)
// 2 цифра: размер текста по горизонтали (ширина) (0 – размер, установленный в настройках, 1 - обычный, 2-8 масштаб)
// 3 цифра: размер текста по вертикали (высота) (0 – размер, установленный в настройках, 1 - обычный, 2-8 масштаб)
// 4 цифра: шрифт (0-шрифт, установленный настройкой <Setup><Font>, 1-А, 2-B)
// 5 цифра: подчеркивание (0-нет, 1-текст, 2-строка)
// 6 цифра: выравнивание (0-лево, 1-центр, 2-право)
// Строки каждого клише надо программировать по одной, подряд без пропуска. Например, если задать строки L0 и L2, то установися только строка L0.
// Установка каждой строки стирает все последующие внутри клише. Например, если сначала задать строки с L0 по L3, а затем повторно задать строки L0 и L1, то строки L2 и L3 сотрутся
func (d *mitsuDriver) SetHeader(headerNum int, lines []ClicheLineData) error {
	var sb strings.Builder

	// Открываем тег команды
	sb.WriteString(fmt.Sprintf("<SET HEADER='%d'>", headerNum))

	// Добавляем строки L0..Ln
	for i, line := range lines {
		// Ограничение: максимум 10 строк (0-9)
		if i > 9 {
			break
		}

		format := line.Format
		if format == "" {
			format = "000000"
		}

		// Экранируем текст
		safeText := escapeXMLText(line.Text)

		// Используем атрибут FORM
		sb.WriteString(fmt.Sprintf("<L%d FORM='%s'>%s</L%d>", i, format, safeText, i))
	}

	sb.WriteString("</SET>")

	_, err := d.sendCommand(sb.String())
	return err
}

func (d *mitsuDriver) SetHeaderLine(headerNum int, lineNum int, text string, format string) error {
	if format == "" {
		format = "000000"
	}
	safeText := escapeXMLText(text)

	// Пример: <SET HEADER='1'><L0 FORM='000011'>Текст</L0></SET>
	cmd := fmt.Sprintf(
		"<SET HEADER='%d'><L%d FORM='%s'>%s</L%d></SET>",
		headerNum, lineNum, format, safeText, lineNum,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetLanSettings (4.9)
func (d *mitsuDriver) SetLanSettings(s LanSettings) error {
	// Все параметры кроме LAN (IP) необязательны, но передаем структуру целиком
	cmd := fmt.Sprintf(
		"<SET LAN='%s' MASK='%s' PORT='%d' DNS='%s' GW='%s'/>",
		s.Addr, s.Mask, s.Port, s.Dns, s.Gw,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetOfdSettings (4.10)
func (d *mitsuDriver) SetOfdSettings(s OfdSettings) error {
	cmd := fmt.Sprintf(
		"<SET OFD='%s' PORT='%d' CLIENT='%s' TimerFN='%d' TimerOFD='%d'/>",
		s.Addr, s.Port, s.Client, s.TimerFN, s.TimerOFD,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetOismSettings (4.11)
func (d *mitsuDriver) SetOismSettings(s OismSettings) error {
	cmd := fmt.Sprintf(
		"<SET OISM='%s' PORT='%d'/>",
		s.Addr, s.Port,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetOkpSettings (4.12)
func (d *mitsuDriver) SetOkpSettings(s ServerSettings) error {
	// Для OKP используется атрибут OKP вместо ADDR
	addr := s.Addr
	if addr == "" && s.Okp != "" {
		addr = s.Okp
	}
	cmd := fmt.Sprintf(
		"<SET OKP='%s' PORT='%d'/>",
		addr, s.Port,
	)
	_, err := d.sendCommand(cmd)
	return err
}

// SetOption (4.13)
// Устанавливает одну опцию b0-b9.
// Значения опций см. в таблице на стр. 22.
// Например, b0=0 (нет разделителей), b1=0 (QR слева).
func (d *mitsuDriver) SetOption(optionNum int, value int) error {
	if optionNum < 0 || optionNum > 9 {
		return fmt.Errorf("неверный номер опции: %d", optionNum)
	}
	cmd := fmt.Sprintf("<OPTION b%d='%d'/>", optionNum, value)
	_, err := d.sendCommand(cmd)
	return err
}

// SetPowerFlag (4.14)
// Сбрасывает (1) или устанавливает (0) флаг сбоя питания.
func (d *mitsuDriver) SetPowerFlag(value int) error {
	cmd := fmt.Sprintf("<SET POWER='%d'/>", value)
	_, err := d.sendCommand(cmd)
	return err
}

// TechReset выполняет технологическое обнуление устройства.
// Команда: <SET FACTORY=”/>
func (d *mitsuDriver) TechReset() error {
	// Ответ: <OK SERIAL='...' FN_STATE='...'/>
	_, err := d.sendCommand("<SET FACTORY=''/>")
	return err
}
