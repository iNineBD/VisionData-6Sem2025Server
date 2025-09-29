package entities

// Fact_Tickets represents the structure of the Fact_Tickets table in the database
type Fact_Tickets struct {
	TicketKey   int64 `gorm:"column:TicketKey;primaryKey;autoIncrement"`
	UserKey     int64 `gorm:"column:UserKey;size:4"`
	AgentKey    int64 `gorm:"column:AgentKey;size:4"`
	CompanyKey  int64 `gorm:"column:CompanyKey;size:4"`
	CategoryKey int64 `gorm:"column:CategoryKey;size:4"`
	PriorityKey int64 `gorm:"column:PriorityKey;size:4"`
	StatusKey   int64 `gorm:"column:StatusKey;size:4"`
	ProductKey  int64 `gorm:"column:ProductKey;size:4"`
	TagKey      int64 `gorm:"column:TagKey;size:4"`
	QtTickets   int64 `gorm:"column:QtTickets;size:4"`
	ChannelKey  int64 `gorm:"column:ChannelKey;size:4"`
}
