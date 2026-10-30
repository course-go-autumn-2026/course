package clients

import "errors"

var ErrLocalRoomUnavailable = errors.New("LocalStudio: option unavailable")
var ErrPartnerRoomUnavailable = errors.New("PartnerStudio: option unavailable")

type LocalStudio struct{}

func (*LocalStudio) Estimate(room string, hours int) (int, int, error) {
	if room == "" || hours <= 0 {
		return 0, 0, errors.New("LocalStudio: invalid request")
	}
	if room != "Студия A" && room != "Студия B" {
		return 0, 0, ErrLocalRoomUnavailable
	}
	return 200 + 800*hours, 20, nil
}

func (*LocalStudio) ListPhotographers() []string {
	return []string{"Первый", "Второй"}
}

type SessionRequest struct {
	RoomName      string
	DurationHours int
}

type SessionResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type PartnerStudio struct{}

func (*PartnerStudio) Calculate(request SessionRequest) (SessionResponse, error) {
	if request.RoomName == "" || request.DurationHours <= 0 {
		return SessionResponse{}, errors.New("PartnerStudio: invalid request")
	}
	if request.RoomName != "Студия A" && request.RoomName != "Студия B" {
		return SessionResponse{}, ErrPartnerRoomUnavailable
	}
	return SessionResponse{
		CostKopecks:     (300+750*request.DurationHours)*100 + 50,
		DurationSeconds: 855,
	}, nil
}

func (*PartnerStudio) Cancel(id string) string {
	return "Съёмка отменена: " + id
}
