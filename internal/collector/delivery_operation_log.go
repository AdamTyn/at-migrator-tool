package collector

import (
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

const DeliveryOLogCollectorName = "DeliveryOLog"

/**
 * DeliveryOLogCollector
 * @author ctc
 * @description 采集器-投递操作日志
 */
type DeliveryOLogCollector struct {
	closed bool
	data   []*entity.DeliveryOperationLog
	tunnel chan *entity.DeliveryOperationLog
	db     *sql.DB
	size   int // 采集器最大缓存数
	logger *log.Logger
}

func NewDeliveryOLogCollector(size int, db *sql.DB, logger *log.Logger) *DeliveryOLogCollector {
	return &DeliveryOLogCollector{
		tunnel: make(chan *entity.DeliveryOperationLog, size+1),
		db:     db,
		logger: logger,
		size:   size,
		data:   make([]*entity.DeliveryOperationLog, 0, size),
	}
}

func (c DeliveryOLogCollector) Name() string {
	return DeliveryOLogCollectorName
}

func (c *DeliveryOLogCollector) Listen() {
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

func (c *DeliveryOLogCollector) flush() {
	if c.closed || len(c.data) < 1 {
		return
	}
	fields := "company_uuid,delivery_uuid,operator_uuid,operator_name,action_type,platform,ip,content,created_time"
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", entity.TableNameDeliveryOperationLog, fields)
	var buf bytes.Buffer
	buf.WriteString(sqlStr)
	for k := range c.data {
		if k == len(c.data)-1 {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s');",
				c.data[k].CompanyUUID, c.data[k].DeliveryUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		} else {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s'),",
				c.data[k].CompanyUUID, c.data[k].DeliveryUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		}
	}
	_, err := c.db.Exec(buf.String())
	fmt.Println(buf.String())
	if err != nil {
		c.logger.Printf("[Exception] DeliveryOLogCollector->flush: %s\n", err.Error())
		os.Exit(1)
	}
	c.data = c.data[0:0:c.size]
}

func (c *DeliveryOLogCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(*entity.OperateRecord); ok {
		deliveryOperationLog := en.DeliveryOperationLog()
		c.tunnel <- deliveryOperationLog
		return nil
	}
	return pkg.ErrCollectorUnSupportType
}

func (c *DeliveryOLogCollector) Close() {
	if !c.closed {
		close(c.tunnel)
	}
	c.closed = true
}
