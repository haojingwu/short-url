package models

import "time"

//URL对于MySQL中的urls表
type URL struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"` //短链码 唯一索引
	OriginalURL string    `gorm:"type:text;not null" json:"original_url"`            //原始长链
	ClickCount  int       `gorm:"default:0" json:"click_count"`                      //点击次数
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

//TableName指定表名(GORM默认加复数,这里显式指定)
func (URL) TableName() string {
	return "urls"
}
