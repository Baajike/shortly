package models

// DailyClick is returned by the clicks-by-day aggregate query.
type DailyClick struct {
	Date  string `gorm:"column:date"  json:"date"`
	Count int64  `gorm:"column:count" json:"count"`
}

// GroupedStat is a generic label→count pair used for device / country / browser breakdowns.
type GroupedStat struct {
	Label string `gorm:"column:label" json:"label"`
	Count int64  `gorm:"column:count" json:"count"`
}

// Analytics is the full analytics payload returned to the client.
type Analytics struct {
	TotalClicks     int64         `json:"total_clicks"`
	ClicksByDay     []DailyClick  `json:"clicks_by_day"`
	ClicksByDevice  []GroupedStat `json:"clicks_by_device"`
	ClicksByCountry []GroupedStat `json:"clicks_by_country"`
	ClicksByBrowser []GroupedStat `json:"clicks_by_browser"`
}
