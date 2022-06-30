package pkg

import (
	"at-migrator-tool/internal/conf"
	"github.com/go-redis/redis"
)

func NewRedis(c *conf.Data_Redis) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		DB:           int(c.Db),
		PoolSize:     20,
		MinIdleConns: 5,
	})
	if _, err := rdb.Ping().Result(); err != nil {
		panic(err)
	}
	return rdb
}
