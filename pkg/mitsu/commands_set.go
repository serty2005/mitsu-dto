package mitsu

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SetDateTime устанавливает дату и время ККТ (См. п. 4.3)
// ВНИМАНИЕ! Если часы кассы установлены по ошибке на будущую дату/время и оформлен
// хотя бы один фискальный документ, вернуть назад часы невозможно.
// В этом случае потребуется замена ФН.
func (c *mitsuClient) SetDateTime(ctx context.Context, t time.Time) error {
	dateStr := t.Format("2006-01-02")
	timeStr := t.Format("15:04:05")
	cmd := fmt.Sprintf("<SET DATE='%s' TIME='%s' />", dateStr, timeStr)

	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetCashier устанавливает данные кассира (См. п. 4.4)
// name: Идентификатор (ФИО, должность), макс 64 символа.
// inn: ИНН кассира (необязательно).
// Необходимо устанавливать кассира перед открытием каждого чека.
func (c *mitsuClient) SetCashier(ctx context.Context, name string, inn string) error {
	safeName := escapeXMLText(name)
	cmd := fmt.Sprintf("<SET CASHIER='%s' INN='%s'/>", safeName, inn)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetComSettings устанавливает настройки COM-порта (См. п. 4.5)
func (c *mitsuClient) SetComSettings(ctx context.Context, speed int32) error {
	cmd := fmt.Sprintf("<SET COM='%d'/>", speed)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetPrinterSettings устанавливает настройки принтера (См. п. 4.6)
func (c *mitsuClient) SetPrinterSettings(ctx context.Context, s PrinterSettings) error {
	cmd := fmt.Sprintf(
		"<SET PRINTER='%s' BAUDRATE='%d' PAPER='%d' FONT='%d'/>",
		s.Model, s.BaudRate, s.Paper, s.Font,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetMoneyDrawerSettings устанавливает настройки денежного ящика (См. п. 4.7)
func (c *mitsuClient) SetMoneyDrawerSettings(ctx context.Context, s DrawerSettings) error {
	cmd := fmt.Sprintf(
		"<SET CD='%d' RISE='%d' FALL='%d'/>",
		s.Pin, s.Rise, s.Fall,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// ClicheLineData содержит данные одной строки клише
type ClicheLineData struct {
	Text   string
	Format string
}

// SetHeader устанавливает клише и подвала (См. п. 4.8)
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
func (c *mitsuClient) SetHeader(ctx context.Context, headerNum int, lines []ClicheLineData) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<SET HEADER='%d'>", headerNum))

	for i, line := range lines {
		if i > 9 {
			break
		}

		format := line.Format
		if format == "" {
			format = "000000"
		}

		safeText := escapeXMLText(line.Text)
		sb.WriteString(fmt.Sprintf("<L%d FORM='%s'>%s</L%d>", i, format, safeText, i))
	}

	sb.WriteString("</SET>")

	_, err := c.SendCommand(ctx, sb.String())
	return err
}

// SetHeaderLine устанавливает одну строку клише (См. п. 4.8)
func (c *mitsuClient) SetHeaderLine(ctx context.Context, headerNum int, lineNum int, text string, format string) error {
	if format == "" {
		format = "000000"
	}
	safeText := escapeXMLText(text)

	cmd := fmt.Sprintf(
		"<SET HEADER='%d'><L%d FORM='%s'>%s</L%d></SET>",
		headerNum, lineNum, format, safeText, lineNum,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetLanSettings устанавливает настройки LAN (См. п. 4.9)
func (c *mitsuClient) SetLanSettings(ctx context.Context, s LanSettings) error {
	cmd := fmt.Sprintf(
		"<SET LAN='%s' MASK='%s' PORT='%d' DNS='%s' GW='%s'/>",
		s.Addr, s.Mask, s.Port, s.Dns, s.Gw,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetOfdSettings устанавливает настройки ОФД (См. п. 4.10)
func (c *mitsuClient) SetOfdSettings(ctx context.Context, s OfdSettings) error {
	cmd := fmt.Sprintf(
		"<SET OFD='%s' PORT='%d' CLIENT='%s' TimerFN='%d' TimerOFD='%d'/>",
		s.Addr, s.Port, s.Client, s.TimerFN, s.TimerOFD,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetOismSettings устанавливает настройки OISM (См. п. 4.11)
func (c *mitsuClient) SetOismSettings(ctx context.Context, s ServerSettings) error {
	cmd := fmt.Sprintf(
		"<SET OISM='%s' PORT='%d'/>",
		s.Addr, s.Port,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetOkpSettings устанавливает настройки OKP (См. п. 4.12)
func (c *mitsuClient) SetOkpSettings(ctx context.Context, s ServerSettings) error {
	addr := s.Addr
	if addr == "" && s.Okp != "" {
		addr = s.Okp
	}
	cmd := fmt.Sprintf(
		"<SET OKP='%s' PORT='%d'/>",
		addr, s.Port,
	)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetOption устанавливает одну опцию устройства (См. п. 4.13)
// Устанавливает одну опцию b0-b9.
// Значения опций см. в таблице на стр. 22.
// Например, b0=0 (нет разделителей), b1=0 (QR слева).
func (c *mitsuClient) SetOption(ctx context.Context, optionNum int, value int) error {
	if optionNum < 0 || optionNum > 9 {
		return fmt.Errorf("неверный номер опции: %d", optionNum)
	}
	cmd := fmt.Sprintf("<OPTION b%d='%d'/>", optionNum, value)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetPowerFlag устанавливает флаг питания (См. п. 4.14)
// Сбрасывает (1) или устанавливает (0) флаг сбоя питания.
func (c *mitsuClient) SetPowerFlag(ctx context.Context, value int) error {
	cmd := fmt.Sprintf("<SET POWER='%d'/>", value)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// SetTimezone устанавливает часовой пояс (Добавлено в FW 1.2.18)
func (c *mitsuClient) SetTimezone(ctx context.Context, value int) error {
	cmd := fmt.Sprintf("<SET TIMEZONE='%d'/>", value)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// TechReset выполняет технологическое обнуление устройства
// Команда: <SET FACTORY=”/>
func (c *mitsuClient) TechReset(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<SET FACTORY=''/>")
	return err
}

// escapeXMLText экранирует специальные символы XML
func escapeXMLText(s string) string {
	result := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			result = append(result, '&', 'a', 'm', 'p', ';')
		case '<':
			result = append(result, '&', 'l', 't', ';')
		case '>':
			result = append(result, '&', 'g', 't', ';')
		case '\'':
			result = append(result, '&', 'a', 'p', 'o', 's', ';')
		case '"':
			result = append(result, '&', 'q', 'u', 'o', 't', ';')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}
