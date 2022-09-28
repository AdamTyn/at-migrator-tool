package collector

import (
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/pkg/log"
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

const DeliverUncheckCollectorName = "DeliverUncheck"

/**
 * DeliverUncheckCollector
 * @author ctc
 * @description 采集器-未处理的投递临时表
 */
type DeliverUncheckCollector struct {
	closed bool
	data   []*entity.Deliver
	tunnel chan *entity.Deliver
	db     *sql.DB
	size   int // 采集器最大缓存数
}

func NewDeliverUncheckCollector(size int, db *sql.DB) *DeliverUncheckCollector {
	return &DeliverUncheckCollector{
		tunnel: make(chan *entity.Deliver, size+1),
		db:     db,
		size:   size,
		data:   make([]*entity.Deliver, 0, size),
	}
}

func (c DeliverUncheckCollector) Name() string {
	return DeliverUncheckCollectorName
}

func (c *DeliverUncheckCollector) Listen() {
	// 每d秒落库1次
	d := 2 * time.Second
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

func (c *DeliverUncheckCollector) flush() {
	if c.closed || len(c.data) < 1 {
		return
	}
	sqlStr := fmt.Sprintf("INSERT INTO %s (deliver_id,intern_uuid) VALUES ", entity.TableNameTmpDeliver)
	var buf bytes.Buffer
	buf.WriteString(sqlStr)
	for k := range c.data {
		if k == len(c.data)-1 {
			buf.WriteString(fmt.Sprintf("(%d,'%s');", c.data[k].ID, c.data[k].InternUUID.String))
		} else {
			buf.WriteString(fmt.Sprintf("(%d,'%s'),", c.data[k].ID, c.data[k].InternUUID.String))
		}
	}
	_, err := c.db.Exec(buf.String())
	if err != nil {
		log.ExceptionF("DeliverUncheckCollector->flush: %s", err.Error())
	}
	c.data = c.data[0:0:c.size]
}

func (c *DeliverUncheckCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(*entity.Deliver); ok {
		c.tunnel <- en
		return nil
	}
	return pkg.ErrCollectorUnSupportType
}

func (c *DeliverUncheckCollector) Close() {
	var old bool
	old, c.closed = c.closed, true
	if !old {
		close(c.tunnel)
	}
}
