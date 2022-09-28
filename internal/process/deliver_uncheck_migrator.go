package process

import (
	"at-migrator-tool/internal"
	col "at-migrator-tool/internal/collector"
	"at-migrator-tool/internal/contract"
	"at-migrator-tool/internal/entity"
	"at-migrator-tool/internal/pkg"
	"at-migrator-tool/internal/pkg/log"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

var uncheckedDeliverStatus = pkg.ArrayString{"delivered", "checked"}

const DeliverUncheckMigratorName = "deliver-uncheck-migrator"

// DeliverUncheckMigrator
type DeliverUncheckMigrator struct {
	Cs            map[string]contract.Collector // 不用考虑并发
	dbSource      *sql.DB                       // 迁移前的数据源
	dbTarget      *sql.DB                       // 迁移后的数据库
	redisCli      *redis.Client
	app           *internal.Application
	closed        bool
	emptyFetchNum int64 // 空查询次数，不用考虑并发
	latestRowId   int64 // 最后操作的id，不用考虑并发
	latestFsRobot int64 // 最后飞书通知毫秒时间戳，不用考虑并发
}

func NewDeliverUncheckMigrator(app *internal.Application, dbSource *sql.DB, dbTarget *sql.DB, redisCli *redis.Client) *DeliverUncheckMigrator {
	m := &DeliverUncheckMigrator{
		app:      app,
		dbSource: dbSource,
		dbTarget: dbTarget,
		redisCli: redisCli,
		Cs:       make(map[string]contract.Collector),
	}
	if !app.Conf.Migrator.DeliverUncheck.Enable {
		log.WarnF("process [%s] not enable", DeliverUncheckMigratorName)
		m.closed = true
	}
	return m
}

func (m DeliverUncheckMigrator) Name() string {
	return DeliverUncheckMigratorName
}

func (m *DeliverUncheckMigrator) Add(c contract.Collector) *DeliverUncheckMigrator {
	if !m.closed {
		m.Cs[c.Name()] = c
	}
	return m
}

func (m *DeliverUncheckMigrator) Start() {
	if m.closed {
		return
	}
	m.collectors()
	m.loadLatestRowId()
	for k := range m.Cs {
		go m.Cs[k].Listen()
	}
	if m.app.Conf.Migrator.DeliverUncheck.TruncateFirst {
		log.Warn("DeliverUncheckMigrator->truncate first !!!")
		sqlStr := fmt.Sprintf("TRUNCATE %s;", entity.TableNameTmpDeliver)
		_, err := m.dbTarget.Exec(sqlStr)
		if err != nil {
			log.ExceptionF("DeliverUncheckMigrator->truncate: %s", err.Error())
		} else {
			m.latestRowId = 1
		}
	}
	// 查询1次数据库
	m.fetch()
	// 之后每d秒查询1次数据库
	d := time.Duration(3)
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

func (m *DeliverUncheckMigrator) Shutdown() {
	var old bool
	old, m.closed = m.closed, true
	if !old {
		for k := range m.Cs {
			m.Cs[k].Close()
		}
	}
	log.NoticeF("process [%s] shutdown", DeliverUncheckMigratorName)
}

/**
 * DeliverUncheckMigrator.collectors
 * @author ctc
 * @description 步骤1：初始化采集器
 */
func (m *DeliverUncheckMigrator) collectors() {
	// 在外部使用Add方法添加
}

func (m DeliverUncheckMigrator) collector(name string) contract.Collector {
	if c, ok := m.Cs[name]; ok {
		return c
	}
	return &col.DefaultCollector{}
}

/**
 * DeliverUncheckMigrator.loadLatestRowId
 * @author ctc
 * @description 步骤2：加载最后操作的id
 */
func (m *DeliverUncheckMigrator) loadLatestRowId() {
	if m.closed {
		return
	}
	defer func() {
		log.InfoF("m.latestRowId=%d", m.latestRowId)
	}()
	// a.0. 先尝试命中redis缓
	// a.1. 同时延长缓存过期时间
	cached, _ := m.redisCli.Get(pkg.CacheDUMigratorLatestRowId).Int64()
	m.redisCli.Expire(pkg.CacheDUMigratorLatestRowId, pkg.CacheDUMigratorExpired)
	if cached > 0 {
		m.latestRowId = cached
		return
	}
	// b.0. 没有命中redis缓存
	// b.1. 兜底最小id是1
	if m.latestRowId < 1 {
		m.latestRowId = 1
	}
}

/**
 * DeliverUncheckMigrator.fetch
 * @author ctc
 * @description 步骤3：从数据源获取数据，分发到不同到采集器
 */
func (m *DeliverUncheckMigrator) fetch() {
	if m.closed {
		return
	}
	// 心跳查询每次id的递增步长
	fetchStep := m.app.Conf.Migrator.DeliverUncheck.FetchStep
	newId := m.latestRowId + fetchStep
	fields := "id,deliver_status,is_sub,intern_uuid"
	// 设置步长，只用主键索引查询
	sqlStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id>=%d AND id<%d LIMIT %d;",
		fields, entity.TableNameDeliver, m.latestRowId, newId, fetchStep,
	)
	log.InfoF("m.dbSource.QueryRow(sqlStr)=%s", sqlStr)
	rows, err0 := m.dbSource.Query(sqlStr)
	if err0 != nil {
		log.ExceptionF("m.dbSource.Query(sqlStr): %s", err0.Error())
		return
	}
	if rows != nil {
		// 查询结果是否为空
		empty := true
		for rows.Next() {
			empty = false
			ent := entity.Deliver{}
			err := rows.Scan(&ent.ID, &ent.DeliverStatus, &ent.IsSub, &ent.InternUUID)
			if err != nil {
				log.ExceptionF("DeliverUncheckMigrator->fetch: %s", err.Error())
				continue
			}
			// 只保留uncheckedDeliverStatus存在的数据
			if uncheckedDeliverStatus.HasNot(ent.DeliverStatus.String) {
				log.NoticeF("DeliverUncheckMigrator->fetch: id=%d discard", ent.ID)
				continue
			}
			// 只保留is_sub=false的数据
			if !ent.IsSub.Valid || ent.IsSub.Bool {
				log.NoticeF("DeliverUncheckMigrator->fetch: id=%d discard", ent.ID)
				continue
			}
			// 只保留intern_uuid不为空的数据
			if ent.InternUUID.String == "" {
				log.NoticeF("DeliverUncheckMigrator->fetch: id=%d discard", ent.ID)
				continue
			}
			err = m.collector(col.DeliverUncheckCollectorName).Put(&ent)
			if err != nil {
				// 需要记录这一部分id到redis
				log.ExceptionF("DeliverUncheckMigrator->fetch->Put: %s", err.Error())
				_ = m.collector(col.BadDataCollectorName).Put(&entity.BadData{K: pkg.CKDUMigratorException, V: ent.ID})
				continue
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

// DeliverUncheckMigrator.isNotEmptyFetch
func (m *DeliverUncheckMigrator) isNotEmptyFetch(empty bool) bool {
	if empty {
		m.emptyFetchNum++
		max := m.app.Conf.Migrator.DeliverUncheck.MaxEmptyFetchNum
		// 超过最大空查询次数，发送飞书通知
		if m.emptyFetchNum >= max {
			// 每6个小时提醒1次
			ms := time.Now().UnixNano() / 1e6
			if ms >= m.latestFsRobot {
				m.latestFsRobot = ms + 21600*1e3
				data := make(map[string]interface{})
				data["msg_type"] = "text"
				data["content"] = map[string]string{
					"text": fmt.Sprintf("[%s] DeliverUncheckMigrator 已经空查询 %d 次了。通知发出时间为:%d", m.app.Conf.Name, m.emptyFetchNum, ms),
				}
				_ = m.collector(col.FsRobotCollectorName).Put(data)
			}
			m.emptyFetchNum = 0
		}
		return false
	}
	m.emptyFetchNum = 0
	return true
}

/**
 * DeliverUncheckMigrator.setLatestRowId
 * @author ctc
 * @description 步骤4：更新最后操作的id
 */
func (m *DeliverUncheckMigrator) setLatestRowId(newId int64) {
	if m.closed {
		return
	}
	m.redisCli.Set(pkg.CacheDUMigratorLatestRowId, newId, pkg.CacheDUMigratorExpired)
	m.latestRowId = newId
}
