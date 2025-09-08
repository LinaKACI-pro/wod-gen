package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/lib/pq"
)

type WodRepositoryInterface interface {
	SaveWod(ctx context.Context, w models.Wod) (models.Wod, error)
	ListWods(ctx context.Context, limit, offset int) ([]models.Wod, error)
}

type WodRepository struct {
	db *sql.DB
}

func NewWodRepository(db *sql.DB) *WodRepository {
	return &WodRepository{db: db}
}

func (r *WodRepository) SaveWod(ctx context.Context, w models.Wod) (models.Wod, error) {
	blocks, err := json.Marshal(w.Blocks)
	if err != nil {
		return models.Wod{}, fmt.Errorf("json.Marshal: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO wods (id, seed, created_at, level, duration_min, equipment, blocks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, w.ID, w.Seed, w.CreatedAt, w.Level, w.DurationMin,
		pq.Array(w.Equipment), blocks,
	)
	if err != nil {
		return models.Wod{}, fmt.Errorf("db.ExecContext: %w", err)
	}

	return w, err
}

func (r *WodRepository) ListWods(ctx context.Context, limit, offset int) ([]models.Wod, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, seed, created_at, level, duration_min, equipment, blocks
		FROM wods
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer func() {
		if err = rows.Close(); err != nil {
			slog.Warn("failed to close rows: ", "err: ", err)
		}
	}()

	var wods []models.Wod
	for rows.Next() {
		var w models.Wod
		var rawBlocks []byte
		err := rows.Scan(
			&w.ID,
			&w.Seed,
			&w.CreatedAt,
			&w.Level,
			&w.DurationMin,
			pq.Array(&w.Equipment),
			&rawBlocks,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		if err := json.Unmarshal(rawBlocks, &w.Blocks); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		wods = append(wods, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return wods, nil
}
