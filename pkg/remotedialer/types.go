package remotedialer

import (
	"time"
)

var (
	PingWaitDuration  = time.Duration(time.Minute)
	PingWriteInterval = time.Duration(10 * time.Second)
)
