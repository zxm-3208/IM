package websocket

import "time"

type Options func(opt *option)

type option struct {
	Authentication    // 嵌入接口
	pattern           string
	maxConnectionIdle time.Duration

	ack          AckType
	ackTimeout   time.Duration
	sendErrCount int
	concurrency  int
	discover     Discover
}

func newOption(opt ...Options) option {
	o := option{
		Authentication:    new(authentication),
		pattern:           "/ws",
		maxConnectionIdle: defaultMaxConnectionIdle,
		ackTimeout:        defaultAckTimeout,
		sendErrCount:      defaultSendErrCount,
		concurrency:       defaultConcurrency,
	}

	for _, opt := range opt {
		opt(&o)
	}

	return o
}

func WithAuthentication(authentication Authentication) Options { // 因为Options是函数类型，所以返回一个匿名函数
	return func(opt *option) {
		opt.Authentication = authentication
	}
}

func WithHandlerPattern(pattern string) Options {
	return func(opt *option) {
		opt.pattern = pattern
	}
}

func WithMaxConnectionIdle(maxConnectionIdle time.Duration) Options {
	return func(opt *option) {
		if maxConnectionIdle > 0 {
			opt.maxConnectionIdle = maxConnectionIdle
		}
	}
}

func WithServerAck(ack AckType) Options {
	return func(opt *option) {
		opt.ack = ack
	}
}

func WithConcurrency(concurrency int) Options {
	return func(opt *option) {
		opt.concurrency = concurrency
	}
}

func WithServerDiscover(discover Discover) Options {
	return func(opt *option) {
		opt.discover = discover
	}
}
