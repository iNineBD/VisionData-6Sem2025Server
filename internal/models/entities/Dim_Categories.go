package entities

type Dim_Categories struct {
	CompanyKey      int64  `gorm:"column:CompanyKey;primaryKey;autoIncrement"`
	CategoryName    string `gorm:"column:CategoryName;size:100"`
	SubCategoryName string `gorm:"column:SubCategoryName;size:100"`
	CategoryId_BK   int64  `gorm:"column:CategoryId_BK;size:4"`
}
