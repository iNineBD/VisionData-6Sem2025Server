package entities

type Dim_Channel struct {
	ChannelKey  int64  `gorm:"column:ChannelKey;primaryKey;autoIncrement"`
	ChannelName string `gorm:"column:ChannelName;size:40"`
}
