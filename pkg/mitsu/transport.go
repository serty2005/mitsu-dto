package mitsu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	stx            = 0x02
	etx            = 0x03
	etb            = 0x17 // End of Transmission Block (для TCP пакетов)
	tcpDataChunkSz = 535  // Макс. кол-во данных в одном TCP пакете перед разбивкой
)

// Config определяет параметры для подключения к ККТ.
type Config struct {
	ConnectionType int32            `json:"connectionType"`      // 0 - COM, 6 - TCP
	IPAddress      string           `json:"ipAddress,omitempty"` // TCP IP
	TCPPort        int32            `json:"tcpPort,omitempty"`   // TCP Port
	ComName        string           `json:"comName,omitempty"`   // COM Port Name
	BaudRate       int32            `json:"baudRate,omitempty"`  // COM Speed
	Timeout        int              `json:"timeout,omitempty"`   // Timeout ms
	Logger         func(msg string) `json:"-"`
}

// Transport инкапсулирует логику работы с соединением (COM или TCP)
type Transport struct {
	config Config
	mu     sync.Mutex
	port   io.ReadWriteCloser // Используется только для COM
}

// NewTransport создаёт новый транспортный слой с заданными конфигурациями
func NewTransport(config Config) *Transport {
	if config.Timeout == 0 {
		config.Timeout = 3000
	}
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	return &Transport{config: config}
}

// Connect устанавливает соединение с устройством
func (t *Transport) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.connectLocked()
}

// connectLocked выполняет подключение (должен вызываться только под мьютексом)
func (t *Transport) connectLocked() error {
	var err error
	switch t.config.ConnectionType {
	case 0: // COM
		if t.port != nil {
			return nil
		}
		mode := &serial.Mode{
			BaudRate: int(t.config.BaudRate),
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		}
		t.port, err = serial.Open(t.config.ComName, mode)
		if err != nil {
			return fmt.Errorf("ошибка открытия COM-порта: %w", err)
		}
		if p, ok := t.port.(serial.Port); ok {
			p.SetReadTimeout(time.Duration(t.config.Timeout) * time.Millisecond)
		}

	case 6: // TCP/IP
		// Для TCP мы используем режим "запрос-ответ" с короткими соединениями.
		// Connect просто проверяет доступность хоста.
		addr := net.JoinHostPort(t.config.IPAddress, strconv.Itoa(int(t.config.TCPPort)))
		conn, err := net.DialTimeout("tcp", addr, time.Duration(t.config.Timeout)*time.Millisecond)
		if err != nil {
			return fmt.Errorf("ошибка подключения TCP: %w", err)
		}
		// Сразу закрываем, реальное соединение будет в performExchange
		conn.Close()
		// t.port остается nil для TCP

	default:
		return fmt.Errorf("неизвестный тип подключения: %d", t.config.ConnectionType)
	}

	return nil
}

// Disconnect разрывает соединение с устройством
func (t *Transport) Disconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.disconnectLocked()
}

// disconnectLocked разрывает соединение (должен вызываться только под мьютексом)
func (t *Transport) disconnectLocked() error {
	if t.port != nil {
		t.port.Close()
		t.port = nil
	}
	return nil
}

// Send отправляет команду устройству с поддержкой фрейминга и повторных попыток
func (t *Transport) Send(xmlCmd string, logEnabled bool) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Попытки нужны в основном для COM порта или если TCP моргнул
	attempts := 1
	if t.config.ConnectionType == 0 {
		attempts = 2
	}

	var lastErr error

	for i := 0; i < attempts; i++ {
		// 1. Проверяем состояние транспорта
		if t.config.ConnectionType == 0 && t.port == nil {
			if err := t.connectLocked(); err != nil {
				lastErr = err
				continue
			}
		}

		// 2. Обмен
		resp, err := t.performExchange(xmlCmd, logEnabled)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 3. Retry логика (только для COM)
		if t.config.ConnectionType == 0 && i < attempts-1 {
			if t.config.Logger != nil {
				t.config.Logger(fmt.Sprintf("COM Error (%v). Retrying...", err))
			}
			t.disconnectLocked()
			time.Sleep(200 * time.Millisecond)
			continue
		}
	}

	return nil, lastErr
}

// performExchange выполняет физическую отправку и прием данных (должен вызываться только под мьютексом)
func (t *Transport) performExchange(xmlCmd string, logEnabled bool) ([]byte, error) {
	if t.config.Logger != nil && logEnabled {
		t.config.Logger(fmt.Sprintf(">> TX: %s", xmlCmd))
	}

	// 1. Подготовка данных (UTF-8 -> Win1251)
	data, err := encodeCP1251(xmlCmd)
	if err != nil {
		return nil, err
	}

	var conn io.ReadWriteCloser

	// Инициализация соединения
	if t.config.ConnectionType == 0 {
		// --- COM ---
		if t.port == nil {
			return nil, errors.New("port is closed")
		}
		conn = t.port
	} else {
		// --- TCP: Transactional Mode ---
		// Открываем сокет на КАЖДЫЙ запрос
		addr := net.JoinHostPort(t.config.IPAddress, strconv.Itoa(int(t.config.TCPPort)))
		netConn, err := net.DialTimeout("tcp", addr, time.Duration(t.config.Timeout)*time.Millisecond)
		if err != nil {
			return nil, err
		}
		defer netConn.Close() // Гарантированно закрываем после обмена

		netConn.SetDeadline(time.Now().Add(time.Duration(t.config.Timeout) * time.Millisecond))
		conn = netConn
	}

	// 2. Отправка (Framing)
	if t.config.ConnectionType == 0 {
		// --- COM Framing (STX...ETX) ---
		packet := make([]byte, 0, len(data)+5)
		packet = append(packet, stx)
		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(data)))
		packet = append(packet, lenBuf...)
		packet = append(packet, data...)
		packet = append(packet, etx)
		lrc := byte(0)
		for _, b := range packet {
			lrc ^= b
		}
		packet = append(packet, lrc)

		if _, err := conn.Write(packet); err != nil {
			return nil, err
		}
	} else {
		// --- TCP Framing (Chunked + ETB) ---
		offset := 0
		totalLen := len(data)

		if totalLen == 0 {
			return nil, errors.New("empty command")
		}

		for offset < totalLen {
			remaining := totalLen - offset
			chunkSize := remaining
			if chunkSize > tcpDataChunkSz {
				chunkSize = tcpDataChunkSz
			}
			chunk := data[offset : offset+chunkSize]

			// Нужно ли слать ETB? (если это НЕ последний пакет)
			isLastPacket := (offset + chunkSize) >= totalLen

			if _, err := conn.Write(chunk); err != nil {
				return nil, err
			}

			if !isLastPacket {
				if _, err := conn.Write([]byte{etb}); err != nil {
					return nil, err
				}
			}
			offset += chunkSize
		}
	}

	// 3. Чтение ответа
	var responseData []byte

	if t.config.ConnectionType == 0 {
		// --- COM Reading ---
		buf := make([]byte, 1)
		readBuf := make([]byte, 0, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				continue
			}
			readBuf = append(readBuf, buf[0])
			if buf[0] == etx {
				lrcBuf := make([]byte, 1)
				_, err := io.ReadFull(conn, lrcBuf)
				if err != nil {
					return nil, err
				}
				readBuf = append(readBuf, lrcBuf[0])
				break
			}
		}
		if len(readBuf) < 2 {
			return nil, errors.New("short response")
		}
		responseData = readBuf[:len(readBuf)-2]

	} else {
		// --- TCP Reading ---
		accumulated := make([]byte, 0, 4096)
		tempBuf := make([]byte, 1024)

		for {
			n, err := conn.Read(tempBuf)
			if err != nil {
				// EOF при TCP Transactional mode - это НОРМАЛЬНОЕ завершение,
				// если мы уже получили данные. Устройство закрыло соединение после ответа.
				if err == io.EOF && len(accumulated) > 0 {
					break
				}
				return nil, err
			}
			if n == 0 {
				continue
			}

			chunk := tempBuf[:n]

			// Обработка ETB (признак продолжения)
			hasEtb := false
			if len(chunk) > 0 && chunk[len(chunk)-1] == etb {
				hasEtb = true
				chunk = chunk[:len(chunk)-1]
			}

			accumulated = append(accumulated, chunk...)

			// Если ETB нет, проверяем, не конец ли это XML
			if !hasEtb {
				tailLen := 50
				if len(accumulated) < tailLen {
					tailLen = len(accumulated)
				}
				tail := string(accumulated[len(accumulated)-tailLen:])

				// Если видим закрывающий тег, считаем ответ полным и выходим,
				// не дожидаясь таймаута или EOF.
				if strings.Contains(tail, "/>") ||
					strings.Contains(tail, "</OK>") ||
					strings.Contains(tail, "</ERROR>") ||
					strings.Contains(tail, "</ANS>") ||
					strings.Contains(tail, "</Do>") ||
					strings.Contains(tail, "</REG>") {
					break
				}
			}
		}
		responseData = accumulated
	}

	// 4. Проверка на логические ошибки
	if bytes.Contains(responseData, []byte("ERROR")) {
		if t.config.Logger != nil {
			decodedLog, _ := toUTF8(responseData)
			t.config.Logger(fmt.Sprintf("<< RX (ERR): %s", string(decodedLog)))
		}
		return nil, ParseError(responseData)
	}

	if t.config.Logger != nil && logEnabled {
		decodedLog, _ := toUTF8(responseData)
		t.config.Logger(fmt.Sprintf("<< RX: %s", string(decodedLog)))
	}

	return responseData, nil
}

// encodeCP1251 конвертирует строку из UTF-8 в Windows-1251
func encodeCP1251(s string) ([]byte, error) {
	encoder := charmap.Windows1251.NewEncoder()
	res, _, err := transform.Bytes(encoder, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("ошибка кодирования в WIN-1251: %w", err)
	}
	return res, nil
}
