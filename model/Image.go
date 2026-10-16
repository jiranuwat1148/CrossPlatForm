package model

import "time"

type Image struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Path      string    `gorm:"not null;uniqueIndex" json:"path"`
	Tags      []Tag     `gorm:"many2many:image_tags;" json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null;uniqueIndex" json:"name"`
}

type TagCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
