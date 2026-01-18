package mitsu

import (
	"context"
	"fmt"
)

// RebootDevice перезапускает устройство
func (c *mitsuClient) RebootDevice(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<DEVICE JOB='0'/>")
	return err
}

// PrintDiagnostics печатает диагностическую информацию
func (c *mitsuClient) PrintDiagnostics(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<MAKE REPORT='X'/>")
	if err != nil {
		return err
	}

	_, err = c.SendCommand(ctx, "<PRINT/>")
	return err
}

// DeviceJob выполняет задачу устройства
func (c *mitsuClient) DeviceJob(ctx context.Context, job int) error {
	cmd := fmt.Sprintf("<DEVICE JOB='%d'/>", job)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// Feed проматывает бумагу на указанное количество строк
func (c *mitsuClient) Feed(ctx context.Context, lines int) error {
	cmd := fmt.Sprintf("<FEED N='%d'/>", lines)
	_, err := c.SendCommand(ctx, cmd)
	return err
}

// Cut выполняет отрезку чека
func (c *mitsuClient) Cut(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<CUT/>")
	return err
}

// PrintLastDocument печатает последний сформированный документ (копию)
func (c *mitsuClient) PrintLastDocument(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<PRINT/>")
	return err
}

// ResetMGM сбрасывает флаг МГМ
func (c *mitsuClient) ResetMGM(ctx context.Context) error {
	_, err := c.SendCommand(ctx, "<MAKE FISCAL='RESET'/>")
	return err
}
