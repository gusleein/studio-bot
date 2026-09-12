package converter

import (
	"github.com/lib/pq"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/pack/model"
)

// ToDomain преобразует модель БД в доменный объект Pack.
func ToDomain(m *model.PackModel) *domain.Pack {
	tags := []string(m.Tags)
	if tags == nil {
		tags = []string{}
	}

	return &domain.Pack{
		ID:             m.ID,
		Title:          m.Title,
		Description:    m.Description,
		Tags:           tags,
		FilePath:       m.FilePath,
		PackType:       domain.PackType(m.PackType),
		TelegramFileID: m.TelegramFileID,
		Notified:       m.Notified,
		AddedAt:        m.AddedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// ToModel преобразует доменный объект Pack в модель БД.
func ToModel(p *domain.Pack) *model.PackModel {
	return &model.PackModel{
		ID:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		Tags:           pq.StringArray(p.Tags),
		FilePath:       p.FilePath,
		PackType:       string(p.PackType),
		TelegramFileID: p.TelegramFileID,
		Notified:       p.Notified,
		AddedAt:        p.AddedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
