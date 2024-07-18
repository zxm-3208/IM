package job

import "time"

const (
	// 默认重试的间隔时间
	DefaultRetryJetLag  = time.Second
	DefaultRetryTimeout = 2 * time.Second
	DefaultRetryNums    = 5
)
