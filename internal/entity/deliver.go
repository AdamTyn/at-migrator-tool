package entity

import "database/sql"

const TableNameDeliver = "deliver"

type Deliver struct {
	ID            int64          `db:"id"`
	UUID          sql.NullString `db:"uuid"`
	InternUUID    sql.NullString `db:"intern_uuid"`
	DeliverStatus sql.NullString `db:"deliver_status"`
	IsSub         sql.NullBool   `db:"is_sub"`
}

func (*Deliver) TableName() string {
	return TableNameDeliver
}
