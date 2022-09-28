package pkg

import "time"

const DatetimeFormatter = "2006-01-02 15:04:05"

const (
	OperateRecordTargetType1 = "intern"
	OperateRecordTargetType2 = "company"
	OperateRecordTargetType3 = "resume"
)

const (
	CacheORMigratorExpired  = 24 * time.Hour
	CKORMigratorLatestRowId = "operate-record-migrator-go"
	CKORMigratorException   = "bd-operate-record:%d"

	CacheDUMigratorExpired     = 24 * time.Hour
	CacheDUMigratorLatestRowId = "deliver-uncheck-migrator-go"
	CKDUMigratorException      = "bd-deliver-uncheck:%d"

	CacheBDMigratorExpired = 48 * time.Hour
	CKBDMigratorException  = "bad-data-go"
)
