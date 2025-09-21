package entities

type Dim_Priorities struct {
	PriorityKey   int64  `gorm:"column:PriorityKey;primaryKey;autoIncrement"`
	Name          string `gorm:"column:Name;size:50"`
	Wheight       int64  `gorm:"column:Wheight;size:4"`
	PrioriryId_BK int64  `gorm:"column:PriorityId_BK;size:4"`
}
