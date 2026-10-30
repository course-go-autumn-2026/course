package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RoomAvailabilityClient struct {
	url    string
	client *http.Client
}

func NewRoomAvailabilityClient(url string) *RoomAvailabilityClient {
	return &RoomAvailabilityClient{
		url:    url,
		client: http.DefaultClient,
	}
}

func (c *RoomAvailabilityClient) Check(ctx context.Context, id int64, startAt time.Time, endAt time.Time) (bool, error) {
	body := fmt.Sprintf(
		`{"room_id":%d,"start_at":%q,"end_at":%q}`,
		id,
		startAt.Format(time.RFC3339),
		endAt.Format(time.RFC3339),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, strings.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
