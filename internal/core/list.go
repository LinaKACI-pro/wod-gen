package core

import (
	"context"
	"fmt"

	"github.com/LinaKACI-pro/wod-gen/internal/core/catalog"
	"github.com/LinaKACI-pro/wod-gen/internal/models"
	"github.com/LinaKACI-pro/wod-gen/internal/repository"
)

type WodListInterface interface {
	List(ctx context.Context, limit, offset int) ([]models.Wod, error)
}

type WodList struct {
	wodRepository repository.WodRepositoryInterface
	catalog       *catalog.Catalog
}

func NewWodList(catalog *catalog.Catalog, wodRepository repository.WodRepositoryInterface) *WodList {
	return &WodList{catalog: catalog, wodRepository: wodRepository}
}

func (w *WodList) List(ctx context.Context, limit, offset int) ([]models.Wod, error) {
	wods, err := w.wodRepository.ListWods(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("wodRepository.ListWods(): %w", err)
	}
	return wods, nil
}
