package repository

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type HTTPModerator struct {
	URL    string
	Client *http.Client
}

func (m *HTTPModerator) Check(ctx context.Context, id int64) (bool, error) {
	client := m.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		m.URL, strings.NewReader(fmt.Sprintf(`{"listing_id":%d}`, id)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err // ошибка наружу, сервис решит pending/review
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

type HTTPAntifraud struct {
	URL    string
	Client *http.Client
}

func (a *HTTPAntifraud) Blocked(ctx context.Context, sellerID int64, price float64) (bool, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		a.URL, strings.NewReader(fmt.Sprintf(`{"seller_id":%d,"price":%f}`, sellerID, price)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err // fail-open на уровне сервиса
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusForbidden, nil
}
