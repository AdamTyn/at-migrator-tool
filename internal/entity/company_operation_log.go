package entity

import (
	"time"
)

const TableNameCompanyOperationLog = "company_operation_log"

type CompanyOperationLog struct {
	ID           int64     `db:"id"`
	CompanyUUID  string    `db:"company_uuid"`  // 企业uuid
	OperatorUUID string    `db:"operator_uuid"` // 操作人uuid
	OperatorName string    `db:"operator_name"` // 操作人名称
	ActionType   string    `db:"action_type"`   // 操作类型
	Platform     string    `db:"platform"`      // 平台
	IP           string    `db:"ip"`            // ip地址
	Content      string    `db:"content"`       // 操作内容
	CreatedTime  time.Time `db:"created_time"`  // 创建时间
}

func DefaultCompanyOperationLog() CompanyOperationLog {
	return CompanyOperationLog{
		OperatorName: "",
		IP:           "0.0.0.0",
		Platform:     "",
		Content:      "{}",
	}
}

func (*CompanyOperationLog) TableName() string {
	return TableNameCompanyOperationLog
}
