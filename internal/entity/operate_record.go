package entity

import (
	"database/sql"
	"strings"
)

const TableNameOperateRecord = "operate_record"

type OperateRecord struct {
	ID          int64          `db:"id"`
	UUID        string         `db:"uuid;not null"` // 总长16字节，前三个前缀后12位随机小写字母加数字
	UserUUID    sql.NullString `db:"user_uuid"`
	CompanyUUID sql.NullString `db:"company_uuid"`
	OpType      sql.NullString `db:"op_type"` // 操作类型
	TargetType  sql.NullString `db:"target_type"`
	TargetUUID  sql.NullString `db:"target_uuid"` // 操作目标uuid，根据类型不同而不同
	BuildTime   sql.NullTime   `db:"build_time"`
	Status      sql.NullString `db:"status"` // 状态字符串
	UpdateTime  sql.NullTime   `db:"update_time"`
	Content     sql.NullString `db:"content"` // 附加内容
}

func (*OperateRecord) TableName() string {
	return TableNameOperateRecord
}

func (ent *OperateRecord) DeliveryOperationLog() *DeliveryOperationLog {
	log := DefaultDeliveryOperationLog()
	log.CompanyUUID = ent.CompanyUUID.String
	log.DeliveryUUID = ent.TargetUUID.String
	log.OperatorUUID = ent.UserUUID.String
	log.ActionType = ent.OpType.String
	log.CreatedTime = ent.BuildTime.Time
	if ent.Content.Valid {
		log.Content = ent.Content.String
	}
	// 特殊处理JSON字符串里面的单引号
	log.Content = strings.Replace(log.Content, "'", "''", -1)
	return &log
}

func (ent *OperateRecord) CompanyOperationLog() *CompanyOperationLog {
	log := DefaultCompanyOperationLog()
	log.CompanyUUID = ent.CompanyUUID.String
	log.OperatorUUID = ent.UserUUID.String
	log.ActionType = ent.OpType.String
	log.CreatedTime = ent.BuildTime.Time
	if ent.Content.Valid {
		log.Content = ent.Content.String
	}
	// 特殊处理JSON字符串里面的单引号
	log.Content = strings.Replace(log.Content, "'", "''", -1)
	return &log
}

func (ent *OperateRecord) InternOperationLog() *InternOperationLog {
	log := DefaultInternOperationLog()
	log.CompanyUUID = ent.CompanyUUID.String
	log.InternUUID = ent.TargetUUID.String
	log.OperatorUUID = ent.UserUUID.String
	log.ActionType = ent.OpType.String
	log.CreatedTime = ent.BuildTime.Time
	if ent.Content.Valid {
		log.Content = ent.Content.String
	}
	// 特殊处理JSON字符串里面的单引号
	log.Content = strings.Replace(log.Content, "'", "''", -1)
	return &log
}
