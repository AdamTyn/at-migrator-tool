package entity

const TableNameTmpDeliver = "tmp_deliver_202209"

type TmpDeliver struct {
	ID         int64  `db:"id"`
	DeliverID  int64  `db:"deliver_id"`
	InternUUID string `db:"intern_uuid"`
}

func (*TmpDeliver) TableName() string {
	return TableNameTmpDeliver
}
