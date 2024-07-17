package websocket

import (
	"github.com/zeromicro/go-zero/core/netx"
	"os"
	"strings"
)

const (
	allEths  = "0.0.0.0"
	envPodIp = "POD_IP"
)

func FigureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
