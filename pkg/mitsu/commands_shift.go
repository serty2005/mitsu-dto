package mitsu

import (
	"context"
	"fmt"
)

// OpenShift открывает смену
func (c *mitsuClient) OpenShift(ctx context.Context, operator string) error {
	if operator != "" {
		if err := c.SetCashier(ctx, operator, ""); err != nil {
			return fmt.Errorf("ошибка установки кассира: %w", err)
		}
	}

	_, err := c.SendCommand(ctx, "<Do SHIFT='OPEN'/>")
	return err
}

// CloseShift закрывает смену
func (c *mitsuClient) CloseShift(ctx context.Context, operator string) error {
	if operator != "" {
		if err := c.SetCashier(ctx, operator, ""); err != nil {
			return fmt.Errorf("ошибка установки кассира: %w", err)
		}
	}

	_, err := c.SendCommand(ctx, "<Do SHIFT='CLOSE'/>")
	return err
}

// PrintXReport печатает X-отчет
func (c *mitsuClient) PrintXReport(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<MAKE REPORT='X'/>")
	if err != nil {
		return err
	}

	_, err = c.SendCommand(ctx, "<PRINT/>")
	return err
}

// PrintZReport печатает отчет по расчетам (Этот отчет не закрывает смену!)
func (c *mitsuClient) PrintZReport(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<MAKE REPORT='Z'/>")
	if err != nil {
		return err
	}

	_, err = c.SendCommand(ctx, "<PRINT/>")
	return err
}
