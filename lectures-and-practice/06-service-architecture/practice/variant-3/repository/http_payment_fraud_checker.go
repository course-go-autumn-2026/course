package repository

import (
	"context"
	"net/http"
)

type HTTPPaymentFraudChecker struct {
	url    string
	client *http.Client
}

func NewHTTPPaymentFraudChecker(url string, client *http.Client) *HTTPPaymentFraudChecker {
	return &HTTPPaymentFraudChecker{url: url, client: client}
}

func (c *HTTPPaymentFraudChecker) Check(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusForbidden, nil
}
