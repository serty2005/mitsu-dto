package ofdclient

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidFnNumber   = errors.New("ofdclient: invalid FN number (must be 16 digits)")
	ErrInvalidFFDVersion = errors.New("ofdclient: invalid FFD version")
	ErrConnectionFailed  = errors.New("ofdclient: connection to OFD failed")
	ErrTimeout           = errors.New("ofdclient: timeout waiting for response")
	ErrInvalidResponse   = errors.New("ofdclient: invalid response from OFD")
	ErrCRCMismatch       = errors.New("ofdclient: CRC mismatch in response")
	ErrNoContainer       = errors.New("ofdclient: response contains no container")
	ErrEmptyContainer    = errors.New("ofdclient: container is empty")
	ErrServerRejected    = errors.New("ofdclient: server rejected message")
)

// OfdError представляет ошибку от сервера ОФД
type OfdError struct {
	Code    int
	Message string
}

func (e *OfdError) Error() string {
	return fmt.Sprintf("ofdclient: OFD error %d: %s", e.Code, e.Message)
}
