package create

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/google/uuid"
)

const maxBodyBytes = 1 << 20

func decodeCreateRequest(w http.ResponseWriter, r *http.Request, request *api.TripData) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	fields, err := requiredFields(body, "user_id", "driver_id", "start_point", "end_point", "price")
	if err != nil {
		return err
	}
	for _, point := range []string{"start_point", "end_point"} {
		if _, err := requiredFields(fields[point], "latitude", "longitude"); err != nil {
			return fmt.Errorf("%s: %w", point, err)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	if request.UserId == uuid.Nil || request.DriverId == uuid.Nil || request.Price < 0 ||
		!validPoint(request.StartPoint) || !validPoint(request.EndPoint) {
		return errors.New("request values are outside the allowed range")
	}
	return nil
}

func requiredFields(body []byte, names ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return nil, errors.New("expected a JSON object")
	}
	for _, name := range names {
		value, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("%s is required", name)
		}
	}
	return fields, nil
}

func validPoint(point api.Coordinates) bool {
	return point.Latitude >= -90 && point.Latitude <= 90 && point.Longitude >= -180 && point.Longitude <= 180
}
