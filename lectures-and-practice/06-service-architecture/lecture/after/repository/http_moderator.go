package repository

import (
	"context"
	"fmt"
	"net/http"
)

// HTTPModerator - адаптер поверх внешнего сервиса модерации.
// Реализует service.Moderator. Если завтра модерация переедет
// за очередь сообщений - появится новый адаптер с тем же интерфейсом,
// а service.ListingService не изменится ни на строку.
type HTTPModerator struct {
	baseURL string
	client  *http.Client
}

func NewHTTPModerator(baseURL string) *HTTPModerator {
	return &HTTPModerator{baseURL: baseURL, client: http.DefaultClient}
}

func (m *HTTPModerator) Check(ctx context.Context, listingID int64) (bool, error) {
	url := fmt.Sprintf("%s/check?listing_id=%d", m.baseURL, listingID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
