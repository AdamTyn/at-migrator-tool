package collector

import (
	"at-migrator-tool/internal/pkg"
	"github.com/go-redis/redis"
	"log"
	"time"
)

const BadDataCollectorName = "BadData"

/**
 * BadDataCollector
 * @author ctc
 * @description 采集器-异常数据
 */
type BadDataCollector struct {
	closed   bool
	data     []int64
	tunnel   chan int64
	size     int // 采集器最大缓存数
	logger   *log.Logger
	redisCli *redis.Client
}

func NewBadDataCollector(size int, redisCli *redis.Client, logger *log.Logger) *BadDataCollector {
	return &BadDataCollector{
		tunnel:   make(chan int64, size+1),
		redisCli: redisCli,
		logger:   logger,
		size:     size,
		data:     make([]int64, 0, size),
	}
}

func (c BadDataCollector) Name() string {
	return BadDataCollectorName
}

func (c *BadDataCollector) Listen() {
	// 每d秒落缓存1次
	d := 1 * time.Second
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case in := <-c.tunnel:
			if len(c.data) == c.size-1 {
				c.data = append(c.data, in)
				c.flush()
			} else {
				if len(c.data) >= c.size {
					c.flush()
				}
				c.data = append(c.data, in)
			}
		case <-ticker.C:
			c.flush()
		default:
			if c.closed {
				return
			}
		}
	}
}

func (c *BadDataCollector) flush() {
	if c.closed || len(c.data) < 1 {
		return
	}
	pipeline := c.redisCli.Pipeline()
	for k := range c.data {
		pipeline.SAdd(pkg.CKORMigratorException, c.data[k])
	}
	pipeline.Expire(pkg.CKORMigratorException, pkg.CacheORMigratorExpired)
	_, err := pipeline.Exec()
	if err != nil {
		c.logger.Printf("[Exception] BadDataCollector->flush: %s\n", err.Error())
	}
	c.data = c.data[0:0:c.size]
}

func (c *BadDataCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(int64); ok {
		c.tunnel <- en
		return nil
	}
	return pkg.ErrCollectorUnSupportType
}

func (c *BadDataCollector) Close() {
	if !c.closed {
		close(c.tunnel)
	}
	c.closed = true
}
