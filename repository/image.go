package repository

import (
	"context"
	"time"

	"Backup_image/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImageRepository interface {
	UpsertAnalyzedImage(ctx context.Context, path string, tagNames []string) (*model.Image, error)
	SearchImages(ctx context.Context, keyword, tag string) ([]model.Image, error)
	GetTagCloud(ctx context.Context) ([]model.TagCount, error)
	DeleteImageRecord(ctx context.Context, id uint) error
}

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) UpsertAnalyzedImage(ctx context.Context, path string, tagNames []string) (*model.Image, error) {
	if r.db == nil {
		return nil, ErrDatabaseUnavailable
	}
	image := &model.Image{Path: path}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newImage := model.Image{Path: path}
		conflict := clause.OnConflict{
			Columns: []clause.Column{{Name: "path"}},
			DoUpdates: clause.Assignments(map[string]any{
				"updated_at": time.Now(),
			}),
		}
		if err := tx.Clauses(conflict).Create(&newImage).Error; err != nil {
			return err
		}
		if err := tx.Where("path = ?", path).First(image).Error; err != nil {
			return err
		}

		tags := make([]model.Tag, 0, len(tagNames))
		for _, name := range tagNames {
			tag := model.Tag{}
			if err := tx.Where("name = ?", name).FirstOrCreate(&tag, model.Tag{Name: name}).Error; err != nil {
				return err
			}
			tags = append(tags, tag)
		}
		if err := tx.Model(image).Association("Tags").Replace(&tags); err != nil {
			return err
		}
		return tx.Preload("Tags").First(image, image.ID).Error
	})
	return image, err
}

func (r *imageRepository) SearchImages(ctx context.Context, keyword, tag string) ([]model.Image, error) {
	if r.db == nil {
		return nil, ErrDatabaseUnavailable
	}
	query := r.db.WithContext(ctx).Model(&model.Image{}).Preload("Tags")
	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where(`images.path LIKE ? OR EXISTS (
			SELECT 1 FROM image_tags
			JOIN tags ON tags.id = image_tags.tag_id
			WHERE image_tags.image_id = images.id AND tags.name LIKE ?
		)`, pattern, pattern)
	}
	if tag != "" {
		query = query.Where(`EXISTS (
			SELECT 1 FROM image_tags
			JOIN tags ON tags.id = image_tags.tag_id
			WHERE image_tags.image_id = images.id AND tags.name = ?
		)`, tag)
	}

	var images []model.Image
	err := query.Order("images.updated_at DESC, images.id DESC").Find(&images).Error
	return images, err
}

func (r *imageRepository) GetTagCloud(ctx context.Context) ([]model.TagCount, error) {
	if r.db == nil {
		return nil, ErrDatabaseUnavailable
	}
	counts := make([]model.TagCount, 0)
	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.name, COUNT(DISTINCT image_tags.image_id) AS count").
		Joins("JOIN image_tags ON image_tags.tag_id = tags.id").
		Group("tags.id, tags.name").
		Order("count DESC, tags.name ASC").
		Scan(&counts).Error
	return counts, err
}

func (r *imageRepository) DeleteImageRecord(ctx context.Context, id uint) error {
	if r.db == nil {
		return ErrDatabaseUnavailable
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var image model.Image
		if err := tx.First(&image, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&image).Association("Tags").Clear(); err != nil {
			return err
		}
		if err := tx.Delete(&image).Error; err != nil {
			return err
		}
		return tx.Where(`NOT EXISTS (
			SELECT 1 FROM image_tags WHERE image_tags.tag_id = tags.id
		)`).Delete(&model.Tag{}).Error
	})
}
