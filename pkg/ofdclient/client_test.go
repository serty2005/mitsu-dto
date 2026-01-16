package ofdclient

import (
	"context"
	"testing"
	"time"
)

// MockTransport — мок-транспорт для тестирования
type MockTransport struct {
	OnSend  func(ctx context.Context, address string, message []byte) ([]byte, error)
	OnClose func() error
}

func (m *MockTransport) Send(ctx context.Context, address string, message []byte) ([]byte, error) {
	if m.OnSend != nil {
		return m.OnSend(ctx, address, message)
	}
	return nil, nil
}

func (m *MockTransport) Close() error {
	if m.OnClose != nil {
		return m.OnClose()
	}
	return nil
}

// createValidResponse создает корректный ответ согласно спецификации
func createValidResponse(bodyData []byte) []byte {
	// Заголовок 30 байт + тело
	response := make([]byte, 30+len(bodyData))

	// Сигнатура: 2A 08 41 0A
	copy(response[0:4], []byte{0x2A, 0x08, 0x41, 0x0A})

	// S-Proto: 81 A2
	copy(response[4:6], []byte{0x81, 0xA2})

	// P-Proto (FFD 1.2): 01 20
	copy(response[6:8], []byte{0x01, 0x20})

	// FN Number
	copy(response[8:24], []byte("1234567890123456"))

	// BodySize (Little Endian)
	bodySize := uint16(len(bodyData))
	response[24] = byte(bodySize & 0xFF)
	response[25] = byte(bodySize >> 8)

	// Flags (Little Endian) - FlagCRCFull | FlagHasContainer = 0x06
	response[26] = 0x06
	response[27] = 0x00

	// CRC (Little Endian) - вычисляем
	crcData := make([]byte, 28+len(bodyData))
	copy(crcData[:28], response[:28])
	copy(crcData[28:], bodyData)
	crc := calcCRC16(crcData)
	response[28] = byte(crc & 0xFF)
	response[29] = byte(crc >> 8)

	// Тело
	copy(response[30:], bodyData)

	return response
}

// TestClientSend проверяет отправку сообщения клиентом
func TestClientSend(t *testing.T) {
	mockTransport := &MockTransport{}
	client := NewWithTransport(Config{
		Timeout:    300 * time.Second,
		RetryCount: 3,
	}, mockTransport)

	t.Run("Successful send", func(t *testing.T) {
		receiptData := []byte("RECEIPT_DATA")
		mockTransport.OnSend = func(ctx context.Context, address string, message []byte) ([]byte, error) {
			return createValidResponse(receiptData), nil
		}

		req := SendRequest{
			OfdAddress: "127.0.0.1:8080",
			FnNumber:   "1234567890123456",
			FFDVersion: "1.2",
			Container:  []byte("test container"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.Send(ctx, req)
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		if string(resp.Receipt) != string(receiptData) {
			t.Errorf("Receipt mismatch: got %s, expected %s", string(resp.Receipt), string(receiptData))
		}
	})

	t.Run("Invalid FN number", func(t *testing.T) {
		req := SendRequest{
			OfdAddress: "127.0.0.1:8080",
			FnNumber:   "12345", // Слишком короткий
			FFDVersion: "1.2",
			Container:  []byte("test container"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Send(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid FN number, got nil")
		}
	})

	t.Run("Empty container", func(t *testing.T) {
		req := SendRequest{
			OfdAddress: "127.0.0.1:8080",
			FnNumber:   "1234567890123456",
			FFDVersion: "1.2",
			Container:  []byte{}, // Пустой
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Send(ctx, req)
		if err == nil {
			t.Error("Expected error for empty container, got nil")
		}
	})

	t.Run("Invalid FFD version", func(t *testing.T) {
		req := SendRequest{
			OfdAddress: "127.0.0.1:8080",
			FnNumber:   "1234567890123456",
			FFDVersion: "2.0", // Несуществующая версия
			Container:  []byte("test container"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Send(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid FFD version, got nil")
		}
	})
}

// TestClientPing проверяет проверку доступности сервера
func TestClientPing(t *testing.T) {
	mockTransport := &MockTransport{}
	client := NewWithTransport(Config{
		Timeout:    300 * time.Second,
		RetryCount: 3,
	}, mockTransport)

	t.Run("Successful ping", func(t *testing.T) {
		mockTransport.OnSend = func(ctx context.Context, address string, message []byte) ([]byte, error) {
			// Возвращаем минимальный валидный ответ (только заголовок)
			return createValidResponse(nil), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Ping(ctx, "127.0.0.1:8080", "1234567890123456")
		if err != nil {
			t.Fatalf("Failed to ping: %v", err)
		}
	})

	t.Run("Invalid FN number", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Ping(ctx, "127.0.0.1:8080", "12345")
		if err == nil {
			t.Error("Expected error for invalid FN number, got nil")
		}
	})
}

// TestMessageBytes проверяет конкретные байты отправляемого сообщения
func TestMessageBytes(t *testing.T) {
	var capturedMessage []byte

	mockTransport := &MockTransport{
		OnSend: func(ctx context.Context, address string, message []byte) ([]byte, error) {
			capturedMessage = message
			return createValidResponse([]byte("OK")), nil
		},
	}

	client := NewWithTransport(Config{}, mockTransport)

	req := SendRequest{
		OfdAddress: "127.0.0.1:8080",
		FnNumber:   "9999078900012345",
		FFDVersion: "1.2",
		Container:  []byte{0x01, 0x02, 0x03},
	}

	ctx := context.Background()
	_, err := client.Send(ctx, req)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Проверяем байты сообщения
	if len(capturedMessage) < 30 {
		t.Fatalf("Message too short: %d bytes", len(capturedMessage))
	}

	// Сигнатура: 2A 08 41 0A
	if capturedMessage[0] != 0x2A || capturedMessage[1] != 0x08 ||
		capturedMessage[2] != 0x41 || capturedMessage[3] != 0x0A {
		t.Errorf("Wrong signature: %X", capturedMessage[0:4])
	}

	// S-Proto: 81 A2
	if capturedMessage[4] != 0x81 || capturedMessage[5] != 0xA2 {
		t.Errorf("Wrong S-proto: %X", capturedMessage[4:6])
	}

	// P-Proto (1.2): 01 20
	if capturedMessage[6] != 0x01 || capturedMessage[7] != 0x20 {
		t.Errorf("Wrong P-proto: %X", capturedMessage[6:8])
	}

	// BodySize = 3 в Little Endian: 03 00
	if capturedMessage[24] != 0x03 || capturedMessage[25] != 0x00 {
		t.Errorf("Wrong BodySize: %X (expected 03 00)", capturedMessage[24:26])
	}

	// Flags = 0x06 в Little Endian: 06 00
	if capturedMessage[26] != 0x06 || capturedMessage[27] != 0x00 {
		t.Errorf("Wrong Flags: %X (expected 06 00)", capturedMessage[26:28])
	}

	t.Logf("Message bytes: %X", capturedMessage)
}
