// Package mitsu предоставляет интерфейс для взаимодействия с фискальными
// регистраторами Mitsu 1-F через прямой протокол обмена (XML over COM/TCP).
package mitsu

import (
	"fmt"
)

// ItemPosition содержит параметры позиции чека.
type ItemPosition struct {
	Name     string  `json:"name"`     // Наименование товара
	Price    float64 `json:"price"`    // Цена
	Quantity float64 `json:"quantity"` // Количество
	TaxRate  int     `json:"taxRate"`  // Ставка НДС (0-6)
}

// PaymentInfo содержит параметры оплаты.
type PaymentInfo struct {
	Type   int     `json:"type"`   // Тип оплаты (0 - наличные, 1 - безналичные, ...)
	Amount float64 `json:"amount"` // Сумма оплаты
}

// OpenShift открывает смену.
func (d *mitsuDriver) OpenShift(operator string) error {
	// Устанавливаем кассира перед открытием смены
	if operator != "" {
		if err := d.SetCashier(operator, ""); err != nil {
			return fmt.Errorf("ошибка установки кассира: %w", err)
		}
	}

	// Открываем смену
	_, err := d.sendCommand("<Do SHIFT='OPEN'/>")
	return err
}

// CloseShift закрывает смену.
func (d *mitsuDriver) CloseShift(operator string) error {
	// Устанавливаем кассира перед закрытием смены
	if operator != "" {
		if err := d.SetCashier(operator, ""); err != nil {
			return fmt.Errorf("ошибка установки кассира: %w", err)
		}
	}

	// Закрываем смену
	_, err := d.sendCommand("<Do SHIFT='CLOSE'/>")
	return err
}

// PrintXReport печатает X-отчет.
func (d *mitsuDriver) PrintXReport() error {
	// Формируем отчет
	_, err := d.sendCommand("<MAKE REPORT='X'/>")
	if err != nil {
		return err
	}

	// Печатаем отчет
	_, err = d.sendCommand("<PRINT/>")
	return err
}

// PrintZReport печатает отчет по расчетам (Этот отчет не закрывает смену!).
func (d *mitsuDriver) PrintZReport() error {
	// Формируем отчет
	_, err := d.sendCommand("<MAKE REPORT='Z'/>")
	if err != nil {
		return err
	}

	// Печатаем отчет
	_, err = d.sendCommand("<PRINT/>")
	return err
}

// OpenCheck открывает чек.
func (d *mitsuDriver) OpenCheck(checkType int, taxSystem int) error {
	cmd := fmt.Sprintf("<Do CHECK='OPEN' TYPE='%d' TAX='%d' MERGE='0'/>", checkType, taxSystem)
	_, err := d.sendCommand(cmd)
	return err
}

// AddPosition добавляет позицию в чек.
func (d *mitsuDriver) AddPosition(pos ItemPosition) error {
	// Маппинг TaxRate: 0->6 (Без НДС), 1->1 (20%), 2->2 (10%), 3->3 (20/120), 4->4 (10/110), 5->5 (0%), 6->6 (Без НДС)
	taxMap := map[int]int{
		0: 6, // Без НДС
		1: 1, // 20%
		2: 2, // 10%
		3: 3, // 20/120
		4: 4, // 10/110
		5: 5, // 0%
		6: 6, // Без НДС
	}
	tax := taxMap[pos.TaxRate]
	if tax == 0 {
		tax = 6 // по умолчанию Без НДС
	}

	total := pos.Price * pos.Quantity
	safeName := escapeXML(pos.Name)

	cmd := fmt.Sprintf("<ADD ITEM='%.3f' TAX='%d' UNIT='0' PRICE='%.2f' TOTAL='%.2f' TYPE='1' MODE='4'><NAME>%s</NAME></ADD>",
		pos.Quantity, tax, pos.Price, total, safeName)
	_, err := d.sendCommand(cmd)
	return err
}

// Subtotal рассчитывает промежуточный итог.
func (d *mitsuDriver) Subtotal() error {
	_, err := d.sendCommand("<Do CHECK='TOTAL'/>")
	return err
}

// Payment производит оплату.
func (d *mitsuDriver) Payment(pay PaymentInfo) error {
	var pa, pb, pc, pd, pe float64
	switch pay.Type {
	case 0: // наличные
		pa = pay.Amount
	case 1: // безналичные
		pb = pay.Amount
	case 2: // аванс
		pc = pay.Amount
	case 3: // кредит
		pd = pay.Amount
	case 4: // иная
		pe = pay.Amount
	default:
		pb = pay.Amount // по умолчанию безналичные
	}

	cmd := fmt.Sprintf("<Do CHECK='PAY' PA='%.2f' PB='%.2f' PC='%.2f' PD='%.2f' PE='%.2f'/></Do>",
		pa, pb, pc, pd, pe)
	_, err := d.sendCommand(cmd)
	return err
}

// CloseCheck закрывает чек.
func (d *mitsuDriver) CloseCheck() error {
	// Завершаем формирование чека
	_, err := d.sendCommand("<Do CHECK='END'/>")
	if err != nil {
		return err
	}

	// Закрываем чек и печатаем
	_, err = d.sendCommand("<Do CHECK='CLOSE'/>")
	if err != nil {
		return err
	}

	// Печатаем чек
	_, err = d.sendCommand("<PRINT/>")
	return err
}

// CancelCheck отменяет чек.
func (d *mitsuDriver) CancelCheck() error {
	_, err := d.sendCommand("<Do CHECK='CANCEL'/>")
	return err
}

// OpenCorrectionCheck открывает чек коррекции.
func (d *mitsuDriver) OpenCorrectionCheck(checkType int, taxSystem int) error {
	cmd := fmt.Sprintf("<Do CHECK='CORR' TYPE='%d' TAX='%d'/>", checkType, taxSystem)
	_, err := d.sendCommand(cmd)
	return err
}

// RebootDevice перезапускает устройство.
func (d *mitsuDriver) RebootDevice() error {
	_, err := d.sendCommand("<DEVICE JOB='0'/>")
	return err
}

// PrintDiagnostics печатает диагностическую информацию.
func (d *mitsuDriver) PrintDiagnostics() error {
	// Формируем отчет
	_, err := d.sendCommand("<MAKE REPORT='X'/>")
	if err != nil {
		return err
	}

	// Печатаем отчет
	_, err = d.sendCommand("<PRINT/>")
	return err
}

// DeviceJob выполняет задачу устройства.
func (d *mitsuDriver) DeviceJob(job int) error {
	cmd := fmt.Sprintf("<DEVICE JOB='%d'/>", job)
	_, err := d.sendCommand(cmd)
	return err
}
