package entities

type Dim_Tags struct {
	TagKey   int64  `gorm:"column:TagKey;primaryKey;autoIncrement"`
	Name     string `gorm:"column:Name;size:60"`
	TagId_BK int64  `gorm:"column:TagId_BK;size:4"`
}
