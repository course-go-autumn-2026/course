package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type NotificationClient struct {
	url    string
	client *http.Client
}

func NewNotificationClient(url string) *NotificationClient {
	return &NotificationClient{
		url:    url,
		client: http.DefaultClient,
	}
}

func (c *NotificationClient) BookingCreated(ctx context.Context, id int64) error {
	body := fmt.Sprintf(`{"booking_id":%d}`, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	return nil
}
