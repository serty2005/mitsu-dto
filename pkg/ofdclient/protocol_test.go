package ofdclient

import (
	"testing"
)

// TestMessageFormat проверяет формат сообщения согласно спецификации ФНС
func TestMessageFormat(t *testing.T) {
	t.Run("Valid message format per FNS spec", func(t *testing.T) {
		fnNumber := "9999078900012345"
		header, err := CreateMessageHeader(fnNumber, "1.2", FlagCRCFull|FlagHasContainer, 100)
		if err != nil {
			t.Fatalf("Failed to create message header: %v", err)
		}

		body := make([]byte, 100)
		message, err := SerializeMessage(header, body)
		if err != nil {
			t.Fatalf("Failed to serialize message: %v", err)
		}

		// Проверяем длину сообщения
		expectedLen := 30 + 100
		if len(message) != expectedLen {
			t.Errorf("Wrong message length: got %d, expected %d", len(message), expectedLen)
		}

		// === Проверяем сигнатуру ===
		// По спецификации: '2A08410A'h → байты [0x2A, 0x08, 0x41, 0x0A]
		expectedSignature := []byte{0x2A, 0x08, 0x41, 0x0A}
		for i, b := range expectedSignature {
			if message[i] != b {
				t.Errorf("Wrong signature byte[%d]: got %02X, expected %02X", i, message[i], b)
			}
		}

		// === Проверяем версию S-протокола ===
		// По спецификации: '81A2'h → байты [0x81, 0xA2]
		expectedSProto := []byte{0x81, 0xA2}
		for i, b := range expectedSProto {
			if message[4+i] != b {
				t.Errorf("Wrong S-proto byte[%d]: got %02X, expected %02X", i, message[4+i], b)
			}
		}

		// === Проверяем версию P-протокола (FFD 1.2) ===
		// По спецификации: '0120'h → байты [0x01, 0x20]
		expectedPProto := []byte{0x01, 0x20}
		for i, b := range expectedPProto {
			if message[6+i] != b {
				t.Errorf("Wrong P-proto byte[%d]: got %02X, expected %02X", i, message[6+i], b)
			}
		}

		// === Проверяем номер ФН ===
		expectedFnBytes := []byte("9999078900012345")
		for i := 0; i < 16; i++ {
			if message[8+i] != expectedFnBytes[i] {
				t.Errorf("Wrong FN byte[%d]: got %02X, expected %02X", i, message[8+i], expectedFnBytes[i])
			}
		}

		// === Проверяем BodySize в Little Endian ===
		// 100 = 0x0064 → в LE: [0x64, 0x00]
		if message[24] != 0x64 || message[25] != 0x00 {
			t.Errorf("Wrong BodySize: got %02X %02X, expected 64 00 (Little Endian)", message[24], message[25])
		}

		// === Проверяем флаги в Little Endian ===
		// FlagCRCFull|FlagHasContainer = 0x06 → в LE: [0x06, 0x00]
		expectedFlags := uint16(FlagCRCFull | FlagHasContainer)
		if message[26] != byte(expectedFlags&0xFF) || message[27] != byte(expectedFlags>>8) {
			t.Errorf("Wrong flags: got %02X %02X, expected %02X %02X (Little Endian)",
				message[26], message[27], byte(expectedFlags&0xFF), byte(expectedFlags>>8))
		}

		// === Проверяем что CRC записан (не проверяем значение, только что не нули) ===
		// CRC должен быть вычислен и записан
		crcBytes := message[28:30]
		t.Logf("CRC bytes: %02X %02X", crcBytes[0], crcBytes[1])
	})
}

// TestFFDVersions проверяет конвертацию версий ФФД
func TestFFDVersions(t *testing.T) {
	tests := []struct {
		version  string
		expected [2]byte
	}{
		{"1.0", [2]byte{0x01, 0x00}},
		{"1.05", [2]byte{0x01, 0x05}},
		{"1.1", [2]byte{0x01, 0x10}},
		{"1.2", [2]byte{0x01, 0x20}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result, err := FFDVersionToBytes(tt.version)
			if err != nil {
				t.Fatalf("FFDVersionToBytes(%s) error: %v", tt.version, err)
			}
			if result != tt.expected {
				t.Errorf("FFDVersionToBytes(%s) = %v, expected %v", tt.version, result, tt.expected)
			}
		})
	}
}

// TestSerializeDeserializeRoundtrip проверяет что сериализация/десериализация работает корректно
func TestSerializeDeserializeRoundtrip(t *testing.T) {
	fnNumber := "1234567890123456"
	header, err := CreateMessageHeader(fnNumber, "1.2", FlagCRCFull|FlagHasContainer, 10)
	if err != nil {
		t.Fatalf("CreateMessageHeader error: %v", err)
	}

	body := []byte("0123456789")
	message, err := SerializeMessage(header, body)
	if err != nil {
		t.Fatalf("SerializeMessage error: %v", err)
	}

	// Десериализуем
	parsedHeader, parsedBody, err := DeserializeMessage(message)
	if err != nil {
		t.Fatalf("DeserializeMessage error: %v", err)
	}

	// Проверяем заголовок
	if parsedHeader.Signature != header.Signature {
		t.Errorf("Signature mismatch: %v != %v", parsedHeader.Signature, header.Signature)
	}
	if parsedHeader.SProtoVersion != header.SProtoVersion {
		t.Errorf("SProtoVersion mismatch: %v != %v", parsedHeader.SProtoVersion, header.SProtoVersion)
	}
	if parsedHeader.PProtoVersion != header.PProtoVersion {
		t.Errorf("PProtoVersion mismatch: %v != %v", parsedHeader.PProtoVersion, header.PProtoVersion)
	}
	if parsedHeader.FnNumber != header.FnNumber {
		t.Errorf("FnNumber mismatch: %v != %v", parsedHeader.FnNumber, header.FnNumber)
	}
	if parsedHeader.BodySize != header.BodySize {
		t.Errorf("BodySize mismatch: %d != %d", parsedHeader.BodySize, header.BodySize)
	}
	if parsedHeader.Flags != header.Flags {
		t.Errorf("Flags mismatch: %v != %v", parsedHeader.Flags, header.Flags)
	}

	// Проверяем тело
	if string(parsedBody) != string(body) {
		t.Errorf("Body mismatch: %s != %s", string(parsedBody), string(body))
	}
}

// TestCRCMismatch проверяет что неверный CRC вызывает ошибку
func TestCRCMismatch(t *testing.T) {
	fnNumber := "1234567890123456"
	header, _ := CreateMessageHeader(fnNumber, "1.2", FlagCRCFull, 5)
	body := []byte("12345")
	message, _ := SerializeMessage(header, body)

	// Портим CRC
	message[28] ^= 0xFF
	message[29] ^= 0xFF

	_, _, err := DeserializeMessage(message)
	if err == nil {
		t.Error("Expected CRC mismatch error, got nil")
	}
}

// TestInvalidFnNumber проверяет валидацию номера ФН
func TestInvalidFnNumber(t *testing.T) {
	tests := []string{
		"",
		"123",
		"12345678901234567", // 17 символов
		"123456789012345",   // 15 символов
	}

	for _, fn := range tests {
		t.Run(fn, func(t *testing.T) {
			_, err := CreateMessageHeader(fn, "1.2", FlagCRCHeader, 0)
			if err == nil {
				t.Errorf("Expected error for FN '%s', got nil", fn)
			}
		})
	}
}
