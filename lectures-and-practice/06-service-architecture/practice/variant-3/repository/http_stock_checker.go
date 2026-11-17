package repository

import (
	"context"
	"net/http"
)

type HTTPStockChecker struct {
	url    string
	client *http.Client
}

func NewHTTPStockChecker(url string, client *http.Client) *HTTPStockChecker {
	return &HTTPStockChecker{url: url, client: client}
}

func (c *HTTPStockChecker) Check(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
