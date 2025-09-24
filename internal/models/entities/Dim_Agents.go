package entities

type Dim_Agents struct {
	AgentKey       int64  `gorm:"column:AgentKey;primaryKey;autoIncrement"`
	FullName       string `gorm:"column:FullName;size:120"`
	DepartmentName string `gorm:"column:DepartmentName;size:100"`
	IsActive       bool   `gorm:"column:IsActive;"`
	AgentId_BK     int64  `gorm:"column:AgentId_BK;size:4"`
}
