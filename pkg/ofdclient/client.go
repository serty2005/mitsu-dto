package ofdclient

import (
	"context"
	"fmt"
	"time"
)

// Client определяет интерфейс OFD-клиента
type Client interface {
	// Send отправляет контейнер с документом в ОФД и возвращает квитанцию
	Send(ctx context.Context, req SendRequest) (*SendResponse, error)

	// Ping проверяет доступность сервера ОФД (служебный режим без контейнера)
	Ping(ctx context.Context, ofdAddress string, fnNumber string) error

	// Close освобождает ресурсы клиента
	Close() error

	// SendRaw отправляет готовое сообщение (уже сформированное ККТ) и возвращает квитанцию
	SendRaw(ctx context.Context, address string, rawMessage []byte) (*SendResponse, error)
}

// New создает новый OFD-клиент с заданной конфигурацией
func New(cfg Config) Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 300 * time.Second // По спецификации ФНС
	}
	return &ofdClient{
		cfg:       cfg,
		transport: NewTCPTransport(cfg.Timeout, cfg.RetryCount, cfg.RetryInterval, cfg.Logger),
	}
}

// NewWithTransport создает клиент с пользовательским транспортом (для тестов)
func NewWithTransport(cfg Config, transport Transport) Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 300 * time.Second
	}
	return &ofdClient{
		cfg:       cfg,
		transport: transport,
	}
}

type ofdClient struct {
	cfg       Config
	transport Transport
}

// Send реализует отправку контейнера с документом в ОФД
func (c *ofdClient) Send(ctx context.Context, req SendRequest) (*SendResponse, error) {
	// Валидация
	if req.FnNumber == "" || len(req.FnNumber) != 16 {
		return nil, ErrInvalidFnNumber
	}
	if req.FFDVersion == "" {
		return nil, ErrInvalidFFDVersion
	}
	if len(req.Container) == 0 {
		return nil, ErrEmptyContainer
	}

	// Добавлен флаг FlagExpectResponse (0x10),
	// чтобы сервер ОФД знал, что мы ждем квитанцию.
	flags := FlagCRCFull | FlagHasContainer | FlagExpectResponse

	// Создаем заголовок сообщения с версией ФФД из запроса
	header, err := CreateMessageHeader(
		req.FnNumber,
		req.FFDVersion,
		flags,
		uint16(len(req.Container)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message header: %w", err)
	}

	// Сериализуем сообщение
	message, err := SerializeMessage(header, req.Container)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Логируем отправку
	if c.cfg.Logger != nil {
		c.cfg.Logger(fmt.Sprintf("Sending document to OFD, FN: %s, FFD: %s, size: %d bytes",
			req.FnNumber, req.FFDVersion, len(req.Container)))
	}

	// Отправляем сообщение через транспорт
	response, err := c.transport.Send(ctx, req.OfdAddress, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Десериализуем ответ
	_, respBody, err := DeserializeMessage(response)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	// Проверяем, что ответ содержит контейнер (квитанцию)
	if len(respBody) == 0 {
		return nil, ErrNoContainer
	}

	// Возвращаем квитанцию
	return &SendResponse{
		Receipt:    respBody,
		RawMessage: response,
	}, nil
}

// SendRaw отправляет готовое бинарное сообщение в ОФД
func (c *ofdClient) SendRaw(ctx context.Context, address string, rawMessage []byte) (*SendResponse, error) {
	if len(rawMessage) == 0 {
		return nil, ErrEmptyContainer
	}

	// Логируем отправку
	if c.cfg.Logger != nil {
		c.cfg.Logger(fmt.Sprintf("Sending RAW message to OFD, size: %d bytes", len(rawMessage)))
	}

	// Отправляем сообщение через транспорт
	response, err := c.transport.Send(ctx, address, rawMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to send raw message: %w", err)
	}

	// Десериализуем ответ, чтобы извлечь квитанцию (тело сообщения)
	_, respBody, err := DeserializeMessage(response)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	// Проверяем, что ответ содержит контейнер (квитанцию)
	if len(respBody) == 0 {
		return nil, ErrNoContainer
	}

	return &SendResponse{
		Receipt:    respBody,
		RawMessage: response,
	}, nil
}

// Ping проверяет доступность сервера ОФД
func (c *ofdClient) Ping(ctx context.Context, ofdAddress string, fnNumber string) error {
	if len(fnNumber) != 16 {
		return ErrInvalidFnNumber
	}

	// Создаем тестовое сообщение (только заголовок без тела)
	header, err := CreateMessageHeader(fnNumber, "1.2", FlagCRCHeader, 0)
	if err != nil {
		return fmt.Errorf("failed to create message header: %w", err)
	}

	// Сериализуем сообщение
	message, err := SerializeMessage(header, nil)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Отправляем сообщение
	_, err = c.transport.Send(ctx, ofdAddress, message)
	return err
}

// Close освобождает ресурсы клиента
func (c *ofdClient) Close() error {
	return c.transport.Close()
}
