package ofdclient

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Transport определяет интерфейс транспорта для связи с ОФД
type Transport interface {
	// Send отправляет сообщение и возвращает ответ
	Send(ctx context.Context, address string, message []byte) ([]byte, error)

	// Close закрывает соединение
	Close() error
}

// TCPTransport реализует транспорт на основе TCP
type TCPTransport struct {
	conn       net.Conn
	timeout    time.Duration
	retryCount int
	retryDelay time.Duration
	logger     func(string)
}

// NewTCPTransport создает новый TCP транспорт
func NewTCPTransport(timeout time.Duration, retryCount int, retryDelay time.Duration, logger func(string)) *TCPTransport {
	return &TCPTransport{
		timeout:    timeout,
		retryCount: retryCount,
		retryDelay: retryDelay,
		logger:     logger,
	}
}

// Send реализует отправку сообщения через TCP
func (t *TCPTransport) Send(ctx context.Context, address string, message []byte) ([]byte, error) {
	var err error

	// Логируем попытку соединения
	if t.logger != nil {
		t.logger(fmt.Sprintf("Connecting to %s", address))
	}

	// Устанавливаем соединение с повторными попытками
	for i := 0; i <= t.retryCount; i++ {
		// Проверяем контекст на отмену
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Устанавливаем соединение
		dialer := &net.Dialer{
			Timeout: t.timeout,
		}

		t.conn, err = dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			break // Успешно подключились
		}

		// Логируем ошибку подключения
		if t.logger != nil {
			t.logger(fmt.Sprintf("Connection attempt %d failed: %v", i+1, err))
		}

		// Если это не последняя попытка, ждем перед повторной попыткой
		if i < t.retryCount {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(t.retryDelay):
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer t.conn.Close()

	// Логируем успешное подключение
	if t.logger != nil {
		t.logger(fmt.Sprintf("Connected to %s", address))
	}

	// Устанавливаем таймаут для соединения
	if err := t.conn.SetDeadline(time.Now().Add(t.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Отправляем сообщение
	if t.logger != nil {
		t.logger(fmt.Sprintf("Sending %d bytes", len(message)))
	}

	n, err := t.conn.Write(message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	if n != len(message) {
		return nil, errors.New("failed to send complete message")
	}

	// Логируем успешную отправку
	if t.logger != nil {
		t.logger("Message sent successfully")
	}

	// === ИСПРАВЛЕНО: Читаем ответ согласно спецификации ===
	// Шаг 1: Читаем заголовок ответа (30 байт)
	headerBuf := make([]byte, MessageHeaderSize)
	if _, err := io.ReadFull(t.conn, headerBuf); err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	// Шаг 2: Извлекаем размер тела из байтов 24-25 (Little Endian!)
	bodySize := binary.LittleEndian.Uint16(headerBuf[24:26])

	if t.logger != nil {
		t.logger(fmt.Sprintf("Response header received, body size: %d bytes", bodySize))
	}

	// Шаг 3: Читаем тело если есть
	var response []byte
	if bodySize > 0 {
		bodyBuf := make([]byte, bodySize)
		if _, err := io.ReadFull(t.conn, bodyBuf); err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		// Собираем полное сообщение: заголовок + тело
		response = make([]byte, MessageHeaderSize+int(bodySize))
		copy(response[:MessageHeaderSize], headerBuf)
		copy(response[MessageHeaderSize:], bodyBuf)
	} else {
		// Только заголовок
		response = headerBuf
	}

	// Логируем успешное получение ответа
	if t.logger != nil {
		t.logger(fmt.Sprintf("Received complete response: %d bytes", len(response)))
	}

	return response, nil
}

// Close закрывает соединение
func (t *TCPTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// Ping проверяет доступность сервера ОФД
func (t *TCPTransport) Ping(ctx context.Context, address string, fnNumber string) error {
	// Создаем тестовое сообщение (только заголовок без тела)
	header, err := CreateMessageHeader(fnNumber, "1.2", FlagCRCHeader, 0)
	if err != nil {
		return err
	}

	// Сериализуем сообщение
	message, err := SerializeMessage(header, nil)
	if err != nil {
		return err
	}

	// Отправляем сообщение
	_, err = t.Send(ctx, address, message)
	return err
}
