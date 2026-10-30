package clients

import "errors"

var ErrBoxOfficeShowUnavailable = errors.New("BoxOffice: option unavailable")
var ErrOnlineShowUnavailable = errors.New("OnlineTickets: option unavailable")

type BoxOffice struct{}

func (*BoxOffice) Estimate(show string, seats int) (int, int, error) {
	if show == "" || seats <= 0 {
		return 0, 0, errors.New("BoxOffice: invalid request")
	}
	if show != "Дневной" && show != "Вечерний" {
		return 0, 0, ErrBoxOfficeShowUnavailable
	}
	return 100 + 300*seats, 5, nil
}

func (*BoxOffice) ListCashiers() []string {
	return []string{"Первый", "Второй"}
}

type TicketRequest struct {
	ShowType  string
	SeatCount int
}

type TicketResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type OnlineTickets struct{}

func (*OnlineTickets) Calculate(request TicketRequest) (TicketResponse, error) {
	if request.ShowType == "" || request.SeatCount <= 0 {
		return TicketResponse{}, errors.New("OnlineTickets: invalid request")
	}
	if request.ShowType != "Дневной" && request.ShowType != "Вечерний" {
		return TicketResponse{}, ErrOnlineShowUnavailable
	}
	return TicketResponse{
		CostKopecks:     (150+280*request.SeatCount)*100 + 50,
		DurationSeconds: 135,
	}, nil
}

func (*OnlineTickets) Cancel(id string) string {
	return "Бронь отменена: " + id
}
