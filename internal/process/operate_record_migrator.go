package process

import (
	"at-migrator-tool/internal"
	col "at-migrator-tool/internal/collector"
	"at-migrator-tool/internal/contract"
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const OperateRecordMigratorName = "operate-record-migrator"

// OperateRecordMigrator
type OperateRecordMigrator struct {
	Cs            map[string]contract.Collector // 不用考虑并发
	dbSource      *sql.DB                       // 迁移前的数据源
	redisCli      *redis.Client
	app           *internal.Application
	closed        bool
	emptyFetchNum int64 // 空查询次数，不用考虑并发
	latestRowId   int64 // 最后操作的id，不用考虑并发
	latestFsRobot int64 // 最后飞书通知毫秒时间戳，不用考虑并发
}

func NewOperateRecordMigrator(app *internal.Application, dbSource *sql.DB, redisCli *redis.Client) *OperateRecordMigrator {
	return &OperateRecordMigrator{
		app:      app,
		dbSource: dbSource,
		redisCli: redisCli,
		Cs:       make(map[string]contract.Collector),
	}
}

func (m OperateRecordMigrator) Name() string {
	return OperateRecordMigratorName
}

func (m *OperateRecordMigrator) Add(c contract.Collector) *OperateRecordMigrator {
	m.Cs[c.Name()] = c
	return m
}

func (m *OperateRecordMigrator) Start() {
	m.collectors()
	m.loadLatestRowId()
	for k := range m.Cs {
		go m.Cs[k].Listen()
	}
	// 查询1次数据库
	m.fetch()
	// 之后每d秒查询1次数据库
	d := time.Duration(10)
	ticker := time.NewTicker(d * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.fetch()
		default:
			if m.closed {
				return
			}
		}
	}
}

func (m *OperateRecordMigrator) Shutdown() {
	if !m.closed {
		for k := range m.Cs {
			m.Cs[k].Close()
		}
	}
	m.closed = true
	m.app.Logf("[Notice] process %s shutdown\n", OperateRecordMigratorName)
}

/**
 * OperateRecordMigrator.collectors
 * @author ctc
 * @description 步骤1：初始化采集器
 */
func (m *OperateRecordMigrator) collectors() {
	// 在外部使用Add方法添加
}

/**
 * OperateRecordMigrator.loadLatestRowId
 * @author ctc
 * @description 步骤2：加载最后操作的id
 */
func (m *OperateRecordMigrator) loadLatestRowId() {
	if m.closed {
		return
	}
	// a.0. 先尝试命中redis缓
	// a.1. 同时延长缓存过期时间
	cached, _ := m.redisCli.Get(pkg.CKORMigratorLatestRowId).Int64()
	m.redisCli.Expire(pkg.CKORMigratorException, pkg.CacheORMigratorExpired)
	if cached > 0 {
		m.latestRowId = cached
		return
	}
	// b.0. 没有命中redis缓存，直接通过数据源查询最小id
	// b.1. 只获取leftDatetime之后的数据
	leftDatetime := "2021-01-01 00:00:00"
	sqlStr := fmt.Sprintf(
		"SELECT MIN(id) AS min_id FROM %s WHERE build_time >= '%s';",
		entity.TableNameOperateRecord, leftDatetime,
	)
	m.app.Logf("[Info] m.dbSource.QueryRow(sqlStr)=%s\n", sqlStr)
	row := m.dbSource.QueryRow(sqlStr)
	if row != nil {
		var minId int64
		_ = row.Scan(&minId)
		if minId > 0 {
			m.latestRowId = minId
		}
	}
	// 兜底最小id是1
	if m.latestRowId < 1 {
		m.latestRowId = 1
	}
	m.app.Logf("[Info] m.latestRowId=%d\n", m.latestRowId)
}

/**
 * OperateRecordMigrator.fetch
 * @author ctc
 * @description 步骤3：从数据源获取数据，分发到不同到采集器
 */
func (m *OperateRecordMigrator) fetch() {
	if m.closed {
		return
	}
	// 心跳查询每次id的递增步长
	fetchStep := m.app.Conf.Migrator.OperateRecord.FetchStep
	newId := m.latestRowId + fetchStep
	fields := "id,user_uuid,company_uuid,op_type,target_type,target_uuid,build_time,status,content"
	// 设置步长，只用主键索引查询
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id >= %d AND id < %d LIMIT %d;",
		fields, entity.TableNameOperateRecord, m.latestRowId, newId, fetchStep,
	)
	m.app.Logf("[Info] m.dbSource.QueryRow(sqlStr)=%s\n", sqlStr)
	rows, err0 := m.dbSource.Query(sqlStr)
	if err0 != nil {
		m.app.Logf("[Exception] m.dbSource.Query(sqlStr):%s\n", err0.Error())
		return
	}
	if rows != nil {
		// 查询结果是否为空
		empty := true
		for rows.Next() {
			empty = false
			ent := entity.OperateRecord{}
			err := rows.Scan(&ent.ID, &ent.UserUUID, &ent.CompanyUUID, &ent.OpType, &ent.TargetType, &ent.TargetUUID, &ent.BuildTime, &ent.Status, &ent.Content)
			if err != nil {
				m.app.Logf("[Exception] OperateRecordMigrator->fetch: %s\n", err.Error())
				continue
			}
			tt := ent.TargetType.String
			switch tt {
			case pkg.OperateRecordTargetType1:
				err = m.Cs[col.InternOLogCollectorName].Put(&ent)
			case pkg.OperateRecordTargetType2:
				err = m.Cs[col.CompanyOLogCollectorName].Put(&ent)
			case pkg.OperateRecordTargetType3:
				err = m.Cs[col.DeliveryOLogCollectorName].Put(&ent)
			default:
				// 需要记录这一部分id到redis
				err = m.Cs[col.BadDataCollectorName].Put(&ent.ID)
			}
			if err != nil {
				// 需要记录这一部分id到redis
				m.app.Logf("[Exception] OperateRecordMigrator->fetch->Put: %s\n", err.Error())
				_ = m.Cs[col.BadDataCollectorName].Put(&ent.ID)
			}
		}
		// 判断是否空查询
		if m.isNotEmptyFetch(empty) {
			// 更新最后操作的id到redis
			m.setLatestRowId(newId)
		}
	}
	_ = rows.Close()
}

// OperateRecordMigrator.isNotEmptyFetch
func (m *OperateRecordMigrator) isNotEmptyFetch(empty bool) bool {
	if empty {
		m.emptyFetchNum++
		max := m.app.Conf.Migrator.OperateRecord.MaxEmptyFetchNum
		// 超过最大空查询次数，发送飞书通知
		if m.emptyFetchNum >= max {
			// 每6个小时提醒1次
			ms := time.Now().UnixNano() / 1e6
			if ms >= m.latestFsRobot {
				m.latestFsRobot = ms + 21600*1e3
				data := make(map[string]interface{})
				data["msg_type"] = "text"
				data["content"] = map[string]string{
					"text": fmt.Sprintf("[at-migrator-tool] OperateRecordMigrator 已经空查询 %d 次了。通知发出时间为:%d", m.emptyFetchNum, ms),
				}
				_ = m.Cs[col.FsRobotCollectorName].Put(data)
			}
			m.emptyFetchNum = 0
		}
		return false
	}
	m.emptyFetchNum = 0
	return true
}

/**
 * OperateRecordMigrator.setLatestRowId
 * @author ctc
 * @description 步骤4：更新最后操作的id
 */
func (m *OperateRecordMigrator) setLatestRowId(newId int64) {
	if m.closed {
		return
	}
	m.redisCli.Set(pkg.CKORMigratorLatestRowId, newId, pkg.CacheORMigratorExpired)
	m.latestRowId = newId
}
