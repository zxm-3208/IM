package websocket

import (
	"math"
	"time"
)

const (
	defaultMaxConnectionIdle = time.Duration(math.MaxInt64)
	defaultAckTimeout        = 30 * time.Second
	defaultSendErrCount      = 5
	defaultConcurrency       = 10
)
