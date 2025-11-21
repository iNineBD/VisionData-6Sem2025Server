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
	db     *gorm.DB
	db_bkp *gorm.DB
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

	dsn = "sqlserver://" + sqlServerUsername + ":" + sqlServerPassword + "@" + sqlServerHost + ":" + sqlServerPort + "?database=LGPD"
	fmt.Println("DSN SQLSERVER:", dsn)

	db2, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB2, err := db2.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB2.Ping(); err != nil {
		return nil, err
	}

	return &Internal{
		db:     db,
		db_bkp: db2,
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

// Retorna o tempo médio de resolução de tickets por prioridade
func (s *Internal) GetAverageResolutionTime() ([]struct {
	NomePrioridade      string  `gorm:"column:nome_prioridade"`
	MediaResolucaoHoras float64 `gorm:"column:media_resolucao_horas"`
	MediaResolucaoDias  float64 `gorm:"column:media_resolucao_dias"`
}, error) {
	var results []struct {
		NomePrioridade      string  `gorm:"column:nome_prioridade"`
		MediaResolucaoHoras float64 `gorm:"column:media_resolucao_horas"`
		MediaResolucaoDias  float64 `gorm:"column:media_resolucao_dias"`
	}
	query := `
    SELECT
        dp.Name as nome_prioridade,
        AVG(CAST(DATEDIFF(SECOND,
            DATETIMEFROMPARTS(de.Year, de.Month, de.Day, de.Hour, de.Minute, 0,0),
            DATETIMEFROMPARTS(dc.Year, dc.Month, dc.Day, dc.Hour, dc.Minute, 0,0)
        ) AS FLOAT) / 3600.0) AS "media_resolucao_horas",
        AVG(CAST(DATEDIFF(SECOND,
            DATETIMEFROMPARTS(de.Year, de.Month, de.Day, de.Hour, de.Minute, 0,0),
            DATETIMEFROMPARTS(dc.Year, dc.Month, dc.Day, dc.Hour, dc.Minute, 0,0)
        ) AS FLOAT) / 86400.0) AS "media_resolucao_dias"
    FROM dbo.Fact_Tickets ft
    JOIN Dim_Priorities dp
        ON ft.PriorityKey = dp.PriorityKey
    JOIN DW.dbo.Dim_Dates de
        ON ft.EntryDateKey = de.DateKey
    JOIN DW.dbo.Dim_Dates dc
        ON ft.ClosedDateKey = dc.DateKey
    WHERE ft.ClosedDateKey IS NOT NULL
    GROUP BY dp.Name
    ORDER BY nome_prioridade;
    `
	err := s.db.Raw(query).Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por status e mês
func (s *Internal) GetTicketsByStatusAndMonth() ([]struct {
	NomeStatus string `gorm:"column:nome_status"`
	Ano        int    `gorm:"column:ano"`
	Janeiro    int    `gorm:"column:janeiro"`
	Fevereiro  int    `gorm:"column:fevereiro"`
	Marco      int    `gorm:"column:marco"`
	Abril      int    `gorm:"column:abril"`
	Maio       int    `gorm:"column:maio"`
	Junho      int    `gorm:"column:junho"`
	Julho      int    `gorm:"column:julho"`
	Agosto     int    `gorm:"column:agosto"`
	Setembro   int    `gorm:"column:setembro"`
	Outubro    int    `gorm:"column:outubro"`
	Novembro   int    `gorm:"column:novembro"`
	Dezembro   int    `gorm:"column:dezembro"`
}, error) {
	var results []struct {
		NomeStatus string `gorm:"column:nome_status"`
		Ano        int    `gorm:"column:ano"`
		Janeiro    int    `gorm:"column:janeiro"`
		Fevereiro  int    `gorm:"column:fevereiro"`
		Marco      int    `gorm:"column:marco"`
		Abril      int    `gorm:"column:abril"`
		Maio       int    `gorm:"column:maio"`
		Junho      int    `gorm:"column:junho"`
		Julho      int    `gorm:"column:julho"`
		Agosto     int    `gorm:"column:agosto"`
		Setembro   int    `gorm:"column:setembro"`
		Outubro    int    `gorm:"column:outubro"`
		Novembro   int    `gorm:"column:novembro"`
		Dezembro   int    `gorm:"column:dezembro"`
	}

	query := `
    WITH Counts AS (
        SELECT
            ds.Name AS status,
            dd.Year,
            dd.Month AS monthnum,
            COUNT(*) AS cnt
        FROM dbo.Fact_Tickets ft
        JOIN DW.dbo.Dim_Dates dd
            ON ft.EntryDateKey = dd.DateKey
        JOIN DW.dbo.Dim_Status ds
            ON ft.StatusKey = ds.StatusKey
        GROUP BY ds.Name, dd.Year, dd.Month
    ),
    Pivoted AS (
        SELECT
            status,
            [Year],
            ISNULL(MAX(CASE WHEN monthnum = 1 THEN cnt END), 0) AS janeiro,
            ISNULL(MAX(CASE WHEN monthnum = 2 THEN cnt END), 0) AS fevereiro,
            ISNULL(MAX(CASE WHEN monthnum = 3 THEN cnt END), 0) AS marco,
            ISNULL(MAX(CASE WHEN monthnum = 4 THEN cnt END), 0) AS abril,
            ISNULL(MAX(CASE WHEN monthnum = 5 THEN cnt END), 0) AS maio,
            ISNULL(MAX(CASE WHEN monthnum = 6 THEN cnt END), 0) AS junho,
            ISNULL(MAX(CASE WHEN monthnum = 7 THEN cnt END), 0) AS julho,
            ISNULL(MAX(CASE WHEN monthnum = 8 THEN cnt END), 0) AS agosto,
            ISNULL(MAX(CASE WHEN monthnum = 9 THEN cnt END), 0) AS setembro,
            ISNULL(MAX(CASE WHEN monthnum = 10 THEN cnt END), 0) AS outubro,
            ISNULL(MAX(CASE WHEN monthnum = 11 THEN cnt END), 0) AS novembro,
            ISNULL(MAX(CASE WHEN monthnum = 12 THEN cnt END), 0) AS dezembro
        FROM Counts
        GROUP BY status, [Year]
    )
    SELECT
        status AS nome_status,
        [Year] AS ano,
        janeiro, fevereiro, marco, abril, maio, junho, julho, agosto, setembro, outubro, novembro, dezembro
    FROM Pivoted
    ORDER BY status, [Year];
    `

	err := s.db.Raw(query).Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por mês e ano
func (s *Internal) GetTicketsByMonth() ([]struct {
	Ano          int `gorm:"column:ano"`
	Mes          int `gorm:"column:mes"`
	TotalTickets int `gorm:"column:total_tickets"`
}, error) {
	var results []struct {
		Ano          int `gorm:"column:ano"`
		Mes          int `gorm:"column:mes"`
		TotalTickets int `gorm:"column:total_tickets"`
	}

	query := `
    SELECT
        dd.Year AS ano,
        dd.Month AS mes,
        COUNT(*) AS total_tickets
    FROM dbo.Fact_Tickets ft
    JOIN DW.dbo.Dim_Dates dd
        ON ft.EntryDateKey = dd.DateKey
    GROUP BY dd.Year, dd.Month
    ORDER BY ano, mes;
    `

	err := s.db.Raw(query).Scan(&results).Error
	return results, err
}

// Retorna o total de tickets por prioridade e mês
func (s *Internal) GetTicketsByPriorityAndMonth() ([]struct {
	NomePrioridades string `gorm:"column:nome_prioridades"`
	Ano             int    `gorm:"column:ano"`
	Janeiro         int    `gorm:"column:janeiro"`
	Fevereiro       int    `gorm:"column:fevereiro"`
	Marco           int    `gorm:"column:marco"`
	Abril           int    `gorm:"column:abril"`
	Maio            int    `gorm:"column:maio"`
	Junho           int    `gorm:"column:junho"`
	Julho           int    `gorm:"column:julho"`
	Agosto          int    `gorm:"column:agosto"`
	Setembro        int    `gorm:"column:setembro"`
	Outubro         int    `gorm:"column:outubro"`
	Novembro        int    `gorm:"column:novembro"`
	Dezembro        int    `gorm:"column:dezembro"`
}, error) {
	var results []struct {
		NomePrioridades string `gorm:"column:nome_prioridades"`
		Ano             int    `gorm:"column:ano"`
		Janeiro         int    `gorm:"column:janeiro"`
		Fevereiro       int    `gorm:"column:fevereiro"`
		Marco           int    `gorm:"column:marco"`
		Abril           int    `gorm:"column:abril"`
		Maio            int    `gorm:"column:maio"`
		Junho           int    `gorm:"column:junho"`
		Julho           int    `gorm:"column:julho"`
		Agosto          int    `gorm:"column:agosto"`
		Setembro        int    `gorm:"column:setembro"`
		Outubro         int    `gorm:"column:outubro"`
		Novembro        int    `gorm:"column:novembro"`
		Dezembro        int    `gorm:"column:dezembro"`
	}

	query := `
    WITH Counts AS (
        SELECT
            dp.Name AS prioridades,
            dd.Year,
            dd.Month AS monthnum,
            COUNT(*) AS cnt
        FROM dbo.Fact_Tickets ft
        JOIN DW.dbo.Dim_Dates dd
            ON ft.EntryDateKey = dd.DateKey
        JOIN DW.dbo.Dim_Priorities dp
            ON ft.PriorityKey = dp.PriorityKey
        GROUP BY dp.Name, dd.Year, dd.Month
    ),
    Pivoted AS (
        SELECT
            prioridades,
            [Year],
            ISNULL(MAX(CASE WHEN monthnum = 1 THEN cnt END), 0) AS janeiro,
            ISNULL(MAX(CASE WHEN monthnum = 2 THEN cnt END), 0) AS fevereiro,
            ISNULL(MAX(CASE WHEN monthnum = 3 THEN cnt END), 0) AS marco,
            ISNULL(MAX(CASE WHEN monthnum = 4 THEN cnt END), 0) AS abril,
            ISNULL(MAX(CASE WHEN monthnum = 5 THEN cnt END), 0) AS maio,
            ISNULL(MAX(CASE WHEN monthnum = 6 THEN cnt END), 0) AS junho,
            ISNULL(MAX(CASE WHEN monthnum = 7 THEN cnt END), 0) AS julho,
            ISNULL(MAX(CASE WHEN monthnum = 8 THEN cnt END), 0) AS agosto,
            ISNULL(MAX(CASE WHEN monthnum = 9 THEN cnt END), 0) AS setembro,
            ISNULL(MAX(CASE WHEN monthnum = 10 THEN cnt END), 0) AS outubro,
            ISNULL(MAX(CASE WHEN monthnum = 11 THEN cnt END), 0) AS novembro,
            ISNULL(MAX(CASE WHEN monthnum = 12 THEN cnt END), 0) AS dezembro
        FROM Counts
        GROUP BY prioridades, [Year]
    )
    SELECT
        prioridades AS nome_prioridades,
        [Year] AS ano,
        janeiro, fevereiro, marco, abril, maio, junho, julho, agosto, setembro, outubro, novembro, dezembro
    FROM Pivoted
    ORDER BY prioridades, [Year];
    `

	err := s.db.Raw(query).Scan(&results).Error
	return results, err
}
