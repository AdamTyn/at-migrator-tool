package collector

import (
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const InternOLogCollectorName = "InternOLog"

/**
 * InternOLogCollector
 * @author ctc
 * @description 采集器-职位操作日志
 */
type InternOLogCollector struct {
	closed bool
	data   []*entity.InternOperationLog
	tunnel chan *entity.InternOperationLog
	db     *sql.DB
	size   int // 采集器最大缓存数
	logger *log.Logger
}

func NewInternOLogCollector(size int, db *sql.DB, logger *log.Logger) *InternOLogCollector {
	return &InternOLogCollector{
		tunnel: make(chan *entity.InternOperationLog, size+1),
		db:     db,
		logger: logger,
		size:   size,
		data:   make([]*entity.InternOperationLog, 0, size),
	}
}

func (c InternOLogCollector) Name() string {
	return InternOLogCollectorName
}

func (c *InternOLogCollector) Listen() {
	// 每d秒落库1次
	d := 3 * time.Second
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

func (c *InternOLogCollector) flush() {
	if c.closed || len(c.data) < 1 {
		return
	}
	fields := "company_uuid,intern_uuid,operator_uuid,operator_name,action_type,platform,ip,content,created_time"
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", entity.TableNameInternOperationLog, fields)
	var buf bytes.Buffer
	buf.WriteString(sqlStr)
	for k := range c.data {
		if k == len(c.data)-1 {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s');",
				c.data[k].CompanyUUID, c.data[k].InternUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		} else {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s'),",
				c.data[k].CompanyUUID, c.data[k].InternUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		}
	}
	_, err := c.db.Exec(buf.String())
	if err != nil {
		c.logger.Printf("[Exception] InternOLogCollector->flush: %s\n", err.Error())
	}
	c.data = c.data[0:0:c.size]
}

func (c *InternOLogCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(*entity.OperateRecord); ok {
		internOperationLog := en.InternOperationLog()
		c.tunnel <- internOperationLog
		return nil
	}
	return pkg.ErrCollectorUnSupportType

}

func (c *InternOLogCollector) Close() {
	if !c.closed {
		close(c.tunnel)
	}
	c.closed = true
}
