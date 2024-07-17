package websocket

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

/*
 * 去中心化服务发现机制
 * 用户连接后，将用户信息与服务器IP一起绑定注册到Redis
 * 当用户发送信息的时候，根据发送的目标从记录位置中获取绑定关系，查找相应服务并发送
 */

type Discover interface {
	// 注册服务
	Register(serverAddr string) error
	// 绑定用户
	BoundUser(uid string) error
	// 解除与用户绑定
	RelieveUser(uid string) error
	// 转发
	Transpond(msg interface{}, uid ...string) error
}

// 默认的
type nopDiscover struct {
	serverAddr string
}

// 注册服务
func (d *nopDiscover) Register(serverAddr string) error {
	return nil
}

// 绑定用户
func (d *nopDiscover) BoundUser(uid string) error {
	return nil
}

// 解除绑定
func (d *nopDiscover) RelieveUser(uid string) error {
	return nil
}

// 转发消息
func (d *nopDiscover) Transpond(msg interface{}, uid ...string) error {
	return nil
}

// 默认的
type redisDiscover struct {
	serverAddr   string
	auth         http.Header
	srvKey       string
	boundUserKey string
	redis        *redis.Redis
	clients      map[string]Client
}

func NewRedisDiscover(auth http.Header, srvKey string, redisCfg redis.RedisConf) *redisDiscover {
	return &redisDiscover{
		auth:         auth,
		srvKey:       fmt.Sprintf("%s:%s", srvKey, "discover"),
		boundUserKey: fmt.Sprintf("%s:%s", srvKey, "boundUserKey"),
		redis:        redis.MustNewRedis(redisCfg),
		clients:      make(map[string]Client),
	}
}

// 注册服务
func (d *redisDiscover) Register(serverAddr string) (err error) {
	d.serverAddr = serverAddr
	// 服务列表: redis存储用set
	go d.redis.Set(d.srvKey, serverAddr)
	return
}

// 绑定用户
func (d *redisDiscover) BoundUser(uid string) (err error) {
	// 用户绑定
	exists, err := d.redis.Hexists(d.boundUserKey, uid)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	// 绑定
	return d.redis.Hset(d.boundUserKey, uid, d.serverAddr)
}

// 解除绑定
func (d *redisDiscover) RelieveUser(uid string) (err error) {
	_, err = d.redis.Hdel(d.boundUserKey, uid)
	return
}

// 转发消息
func (d *redisDiscover) Transpond(msg interface{}, uid ...string) (err error) {
	for _, uid := range uid {
		srvAddr, err := d.redis.Hget(d.srvKey, uid)
		if err != nil {
			return err
		}
		srvClient, ok := d.clients[srvAddr]
		if !ok {
			srvClient = d.createClient(srvAddr)
		}

		fmt.Println("redis transpand -》 ", srvAddr, " uid ", uid)
		if err := d.send(srvClient, msg, uid); err != nil {
			return err
		}
	}
	return
}

func (d *redisDiscover) send(srvClient Client, msg interface{}, uid string) error {
	return srvClient.Send(Message{
		Type:         FrameTranspond,
		TranspondUid: uid,
		Data:         msg,
	})
}

func (d *redisDiscover) createClient(serverAddr string) Client {
	return NewClient(serverAddr, WithClientHeader(d.auth))
}
