package mitsu

import (
	"context"
	"fmt"
)

// ItemPosition представляет позицию в чеке
type ItemPosition struct {
	Name     string
	Quantity float64
	Price    float64
	Tax      int
}

// PaymentInfo представляет информацию об оплате
type PaymentInfo struct {
	Type int     // 0 - наличные, 1 - безналичные, 2 - аванс, 3 - кредит, 4 - иная
	Sum  float64 // сумма оплаты
}

// OpenCheck открывает чек
func (c *mitsuClient) OpenCheck(ctx context.Context, checkType int, taxSystem int) error {
	cmd := fmt.Sprintf("<Do CHECK='OPEN' TYPE='%d' TAX='%d' MERGE='0'/>", checkType, taxSystem)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// AddPosition добавляет позицию в чек
func (c *mitsuClient) AddPosition(ctx context.Context, pos ItemPosition) error {
	taxMap := map[int]int{
		0: 6, // Без НДС
		1: 1, // 20%
		2: 2, // 10%
		3: 3, // 20/120
		4: 4, // 10/110
		5: 5, // 0%
		6: 6, // Без НДС
	}
	tax := taxMap[pos.Tax]
	if tax == 0 {
		tax = 6 // по умолчанию Без НДС
	}

	total := pos.Price * pos.Quantity
	safeName := escapeXMLText(pos.Name)

	cmd := fmt.Sprintf("<ADD ITEM='%.3f' TAX='%d' UNIT='0' PRICE='%.2f' TOTAL='%.2f' TYPE='1' MODE='4'><NAME>%s</NAME></ADD>",
		pos.Quantity, tax, pos.Price, total, safeName)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// Subtotal рассчитывает промежуточный итог
func (c *mitsuClient) Subtotal(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<Do CHECK='TOTAL'/>")
	return err
}

// Payment производит оплату
func (c *mitsuClient) Payment(ctx context.Context, pay PaymentInfo) error {
	var pa, pb, pc, pd, pe float64
	switch pay.Type {
	case 0: // наличные
		pa = pay.Sum
	case 1: // безналичные
		pb = pay.Sum
	case 2: // аванс
		pc = pay.Sum
	case 3: // кредит
		pd = pay.Sum
	case 4: // иная
		pe = pay.Sum
	default:
		pb = pay.Sum // по умолчанию безналичные
	}

	cmd := fmt.Sprintf("<Do CHECK='PAY' PA='%.2f' PB='%.2f' PC='%.2f' PD='%.2f' PE='%.2f'/></Do>",
		pa, pb, pc, pd, pe)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// CloseCheck закрывает чек
func (c *mitsuClient) CloseCheck(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<Do CHECK='END'/>")
	if err != nil {
		return err
	}

	_, err = c.SendCommand(ctx, "<Do CHECK='CLOSE'/>")
	if err != nil {
		return err
	}

	_, err = c.SendCommand(ctx, "<PRINT/>")
	return err
}

// CancelCheck отменяет чек
func (c *mitsuClient) CancelCheck(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<Do CHECK='CANCEL'/>")
	return err
}

// OpenCorrectionCheck открывает чек коррекции
func (c *mitsuClient) OpenCorrectionCheck(ctx context.Context, checkType int, taxSystem int) error {
	cmd := fmt.Sprintf("<Do CHECK='CORR' TYPE='%d' TAX='%d'/>", checkType, taxSystem)
	_, err := c.SendCommand(ctx, cmd)
	return err
}
