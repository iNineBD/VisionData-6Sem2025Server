package sqlserver

import (
	"fmt"
	"orderstreamrest/internal/models/entities"
	"os"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

//total tickets
// total tickets by category -
// total tickets by priority -
// total tickets by channel -
// total tickets by tag -
// total tickets by department PERGUNTAR PRO ANDRÉ

// SQLServerInternal is a struct that contains a SQL Server database connection
type Internal struct {
	db *gorm.DB
}

// NewSQLServerInternal is a function that returns a new SQLServerInternal struct
func NewSQLServerInternal() (*Internal, error) {

	sqlServerUsername := os.Getenv("SQLSERVER_USERNAME")
	sqlServerPassword := os.Getenv("SQLSERVER_PASSWORD")
	sqlServerHost := os.Getenv("SQLSERVER_HOST")
	sqlServerPort := os.Getenv("SQLSERVER_PORT")
	sqlServerDatabase := os.Getenv("SQLSERVER_DATABASE")

	dsn := "sqlserver://" + sqlServerUsername + ":" + sqlServerPassword + "@" + sqlServerHost + ":" + sqlServerPort + "?database=" + sqlServerDatabase
	fmt.Println("DSN SQLSERVER:", dsn)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &Internal{
		db: db,
	}, nil
}

// Retorna o total de tickets
func (s *Internal) GetTotalTickets() (int64, error) {
	var total int64
	err := s.db.Table("dbo.Fact_Tickets").
		Select("SUM(QtTickets)").
		Scan(&total).Error
	return total, err
}

// Retorna o total de tickets agrupados por categoria
func (s *Internal) GetTicketsByCategory() ([]struct {
	entities.Dim_Categories
	Total int64
}, error) {
	var results []struct {
		entities.Dim_Categories
		Total int64
	}
	err := s.db.Table("dbo.Fact_Tickets ft").
		Select("dc.CategoryName, SUM(ft.QtTickets) as Total").
		Joins("INNER JOIN dbo.Dim_Categories dc ON ft.CategoryKey = dc.CategoryKey").
		Group("dc.CategoryName").
		Order("Total DESC").
		Scan(&results).Error
	return results, err
}

// Retorna o total de tickets agrupados por prioridade
func (s *Internal) GetTicketsByPriority() ([]struct {
	entities.Dim_Priorities
	Total int64
}, error) {
	var results []struct {
		entities.Dim_Priorities
		Total int64
	}
	err := s.db.Table("dbo.Fact_Tickets ft").
		Select("dp.Name, SUM(ft.QtTickets) as Total").
		Joins("INNER JOIN dbo.Dim_Priorities dp ON ft.PriorityKey = dp.PriorityKey").
		Group("dp.Name").
		Order("Total DESC").
		Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por channel
func (s *Internal) GetTicketsByChannel() ([]struct {
	entities.Dim_Channel
	Total int64
}, error) {
	var results []struct {
		entities.Dim_Channel
		Total int64
	}
	err := s.db.Table("dbo.Fact_Tickets ft").
		Select("dc.ChannelName, SUM(ft.QtTickets) as Total").
		Joins("INNER JOIN dbo.Dim_Channel dc ON ft.ChannelKey = dc.ChannelKey").
		Group("dc.ChannelName").
		Order("Total DESC").
		Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por tag
func (s *Internal) GetTicketsByTag() ([]struct {
	entities.Dim_Tags
	Total int64
}, error) {
	var results []struct {
		entities.Dim_Tags
		Total int64
	}
	err := s.db.Table("dbo.Fact_Tickets ft").
		Select("dt.Name, SUM(ft.QtTickets) as Total").
		Joins("INNER JOIN dbo.Dim_Tags dt ON ft.TagKey = dt.TagKey").
		Group("dt.Name").
		Order("Total DESC").
		Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por departamento
func (s *Internal) GetTicketsByDepartment() ([]struct {
	entities.Dim_Companies
	Total int64
}, error) {
	var results []struct {
		entities.Dim_Companies
		Total int64
	}
	err := s.db.Table("dbo.Fact_Tickets ft").
		Select("dc.Name, SUM(ft.QtTickets) as Total").
		Joins("INNER JOIN dbo.Dim_Companies dc ON ft.CompanyKey = dc.CompanyKey").
		Group("dc.Name").
		Order("Total DESC").
		Scan(&results).Error
	return results, err
}
