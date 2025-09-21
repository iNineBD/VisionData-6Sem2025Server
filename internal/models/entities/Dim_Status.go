package entities

type Dim_Status struct {
	StatusKey   int64  `gorm:"column:StatusKey;primaryKey;autoIncrement"`
	Name        string `gorm:"column:Name;size:60"`
	StatusId_BK int64  `gorm:"column:StatusId_BK;size:4"`
}
