package entities

type Dim_Products struct {
	ProductKey   int64  `gorm:"column:ProductKey;primaryKey;autoIncrement"`
	Name         string `gorm:"column:Name;size:150"`
	Code         string `gorm:"column:Code;size:50"`
	IsActive     bool   `gorm:"column:IsActive"`
	ProductId_BK int64  `gorm:"column:ProductId_BK;size:4"`
}
