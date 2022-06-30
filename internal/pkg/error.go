package pkg

import "errors"

var (
	ErrCollectorUnSupportType = errors.New("collector unsupported type")
	ErrCollectorClosed        = errors.New("collector closed")
)
