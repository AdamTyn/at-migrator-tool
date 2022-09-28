package collector

import (
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/pkg/log"
	"fmt"
	"github.com/go-redis/redis"
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
	data     []*entity.BadData
	tunnel   chan *entity.BadData
	size     int // 采集器最大缓存数
	redisCli *redis.Client
}

func NewBadDataCollector(size int, redisCli *redis.Client) *BadDataCollector {
	return &BadDataCollector{
		tunnel:   make(chan *entity.BadData, size+1),
		redisCli: redisCli,
		size:     size,
		data:     make([]*entity.BadData, 0, size),
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
	p := c.redisCli.Pipeline()
	defer p.Close()
	for k := range c.data {
		val := fmt.Sprintf(c.data[k].K, c.data[k].V)
		p.SAdd(pkg.CKBDMigratorException, val)
	}
	p.Expire(pkg.CKBDMigratorException, pkg.CacheBDMigratorExpired)
	_, err := p.Exec()
	if err != nil {
		log.ExceptionF("BadDataCollector->flush: %s", err.Error())
	}
	c.data = c.data[0:0:c.size]
}

func (c *BadDataCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(*entity.BadData); ok {
		c.tunnel <- en
		return nil
	}
	return pkg.ErrCollectorUnSupportType
}

func (c *BadDataCollector) Close() {
	var old bool
	old, c.closed = c.closed, true
	if !old {
		close(c.tunnel)
	}
}
