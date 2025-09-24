package entities

type Dim_Users struct {
	UserKey   int64  `gorm:"column:UserKey;primaryKey;autoIncrement"`
	FullName  string `gorm:"column:FullName;size:120"`
	IsVIP     bool   `gorm:"column:IsVIP;"`
	UserId_BK int64  `gorm:"column:UserId_BK;size:4"`
}
