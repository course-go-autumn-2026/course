package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/api"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
	"github.com/google/uuid"
)

func decodeCreateTrip(r *http.Request) (api.TripData, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return api.TripData{}, fmt.Errorf("read request body: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := decodeStrictJSON(body, &raw); err != nil {
		return api.TripData{}, err
	}

	for _, field := range []string{"user_id", "driver_id", "price"} {
		if _, err := requiredField(raw, field); err != nil {
			return api.TripData{}, err
		}
	}

	startPoint, err := requiredField(raw, "start_point")
	if err != nil {
		return api.TripData{}, err
	}
	if err := validatePointFields(startPoint, "start_point"); err != nil {
		return api.TripData{}, err
	}

	endPoint, err := requiredField(raw, "end_point")
	if err != nil {
		return api.TripData{}, err
	}
	if err := validatePointFields(endPoint, "end_point"); err != nil {
		return api.TripData{}, err
	}

	var value api.TripData
	if err := decodeStrictJSON(body, &value); err != nil {
		return api.TripData{}, err
	}

	return value, nil
}

func decodeStrictJSON(body []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(value); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON object")
	}

	return nil
}

func requiredField(
	object map[string]json.RawMessage,
	name string,
) (json.RawMessage, error) {
	value, ok := object[name]
	if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, fmt.Errorf("%s is required", name)
	}

	return value, nil
}

func validatePointFields(raw json.RawMessage, name string) error {
	var point map[string]json.RawMessage
	if err := decodeStrictJSON(raw, &point); err != nil {
		return fmt.Errorf("%s must be an object: %w", name, err)
	}

	for _, field := range []string{"latitude", "longitude"} {
		if _, err := requiredField(point, field); err != nil {
			return fmt.Errorf("%s.%w", name, err)
		}
	}

	return nil
}

func createTripInput(value api.TripData) (trip.CreateTripInput, error) {
	if value.UserId == uuid.Nil {
		return trip.CreateTripInput{}, errors.New("user_id must be a non-empty UUID")
	}

	if value.DriverId == uuid.Nil {
		return trip.CreateTripInput{}, errors.New("driver_id must be a non-empty UUID")
	}

	if err := validateCoordinates("start_point", value.StartPoint); err != nil {
		return trip.CreateTripInput{}, err
	}

	if err := validateCoordinates("end_point", value.EndPoint); err != nil {
		return trip.CreateTripInput{}, err
	}

	if value.Price < 0 {
		return trip.CreateTripInput{}, errors.New("price cannot be negative")
	}

	return trip.CreateTripInput{
		UserID:   value.UserId,
		DriverID: value.DriverId,
		StartPoint: trip.Coordinates{
			Latitude:  value.StartPoint.Latitude,
			Longitude: value.StartPoint.Longitude,
		},
		EndPoint: trip.Coordinates{
			Latitude:  value.EndPoint.Latitude,
			Longitude: value.EndPoint.Longitude,
		},
		Price: value.Price,
	}, nil
}

func validateCoordinates(name string, value api.Coordinates) error {
	if value.Latitude < -90 || value.Latitude > 90 {
		return fmt.Errorf("%s.latitude must be between -90 and 90", name)
	}

	if value.Longitude < -180 || value.Longitude > 180 {
		return fmt.Errorf("%s.longitude must be between -180 and 180", name)
	}

	return nil
}
