package product

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/product/converter"
	"github.com/yourstudio/studio-bot/internal/repository/product/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	createQuery = `
		INSERT INTO products
			(id, name, description, price, created_at, updated_at)
		VALUES
			(:id, :name, :description, :price, :created_at, :updated_at)`

	getByIDQuery = `
		SELECT id, name, description, price, created_at, updated_at
		FROM products
		WHERE id = $1`

	listQuery = `
		SELECT id, name, description, price, created_at, updated_at
		FROM products
		ORDER BY name ASC`
)

type Repo struct {
	db *sqlx.DB
}

var _ repository.ProductRepository = (*Repo)(nil)

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	db := database.GetDB(ctx, r.db)

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	if _, err := db.NamedExecContext(ctx, createQuery, converter.ToModel(p)); err != nil {
		return nil, errors.Wrap(err, "create product")
	}
	return p, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	db := database.GetDB(ctx, r.db)

	var m model.ProductModel
	if err := db.GetContext(ctx, &m, getByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get product by id")
	}
	return converter.ToDomain(&m), nil
}

func (r *Repo) List(ctx context.Context) ([]*domain.Product, error) {
	db := database.GetDB(ctx, r.db)

	var rows []model.ProductModel
	if err := db.SelectContext(ctx, &rows, listQuery); err != nil {
		return nil, errors.Wrap(err, "list products")
	}

	result := make([]*domain.Product, len(rows))
	for i := range rows {
		result[i] = converter.ToDomain(&rows[i])
	}
	return result, nil
}
