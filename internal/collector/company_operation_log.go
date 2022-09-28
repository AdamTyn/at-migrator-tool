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

const CompanyOLogCollectorName = "CompanyOLog"

/**
 * CompanyOLogCollector
 * @author ctc
 * @description 采集器-企业操作日志
 */
type CompanyOLogCollector struct {
	closed bool
	data   []*entity.CompanyOperationLog
	tunnel chan *entity.CompanyOperationLog
	db     *sql.DB
	size   int // 采集器最大缓存数
}

func NewCompanyOLogCollector(size int, db *sql.DB) *CompanyOLogCollector {
	return &CompanyOLogCollector{
		tunnel: make(chan *entity.CompanyOperationLog, size+1),
		db:     db,
		size:   size,
		data:   make([]*entity.CompanyOperationLog, 0, size),
	}
}

func (c CompanyOLogCollector) Name() string {
	return CompanyOLogCollectorName
}

func (c *CompanyOLogCollector) Listen() {
	// 每d秒落库1次
	d := 5 * time.Second
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

func (c *CompanyOLogCollector) flush() {
	if c.closed || len(c.data) < 1 {
		return
	}
	fields := "company_uuid,operator_uuid,operator_name,action_type,platform,ip,content,created_time"
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", entity.TableNameCompanyOperationLog, fields)
	var buf bytes.Buffer
	buf.WriteString(sqlStr)
	for k := range c.data {
		if k == len(c.data)-1 {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s');",
				c.data[k].CompanyUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		} else {
			buf.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s'),",
				c.data[k].CompanyUUID, c.data[k].OperatorUUID, c.data[k].OperatorName, c.data[k].ActionType,
				c.data[k].Platform, c.data[k].IP, c.data[k].Content, c.data[k].CreatedTime.Format(pkg.DatetimeFormatter)))
		}
	}
	_, err := c.db.Exec(buf.String())
	if err != nil {
		log.ExceptionF("CompanyOLogCollector->flush: %s", err.Error())
	}
	c.data = c.data[0:0:c.size]
}

func (c *CompanyOLogCollector) Put(in interface{}) error {
	if c.closed {
		return pkg.ErrCollectorClosed
	}
	if en, ok := in.(*entity.OperateRecord); ok {
		companyOperationLog := en.CompanyOperationLog()
		c.tunnel <- companyOperationLog
		return nil
	}
	return pkg.ErrCollectorUnSupportType
}

func (c *CompanyOLogCollector) Close() {
	var old bool
	old, c.closed = c.closed, true
	if !old {
		close(c.tunnel)
	}
}
