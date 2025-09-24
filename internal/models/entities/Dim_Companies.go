package entities

type Dim_Companies struct {
	CompanyKey   int64  `gorm:"column:CompanyKey;primaryKey;autoIncrement"`
	Name         string `gorm:"column:Name;size:120"`
	Segmento     string `gorm:"column:Segmento;size:60"`
	CNPJ         string `gorm:"column:CNPJ;size:32"`
	CompanyId_BK int64  `gorm:"column:CompanyId_BK;size:4"`
}
