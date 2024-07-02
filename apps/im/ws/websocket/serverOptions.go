package websocket

type Options func(opt *option)

type option struct {
	Authentication // 嵌入接口
	pattern        string
}

func newOption(opt ...Options) option {
	o := option{
		Authentication: new(authentication),
		pattern:        "/ws",
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
