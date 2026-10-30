package clients

import "errors"

var ErrClubCategoryUnavailable = errors.New("ClubLibrary: option unavailable")
var ErrPartnerCategoryUnavailable = errors.New("PartnerLibrary: option unavailable")

type ClubLibrary struct{}

func (*ClubLibrary) Estimate(category string, days int) (int, int, error) {
	if category == "" || days <= 0 {
		return 0, 0, errors.New("ClubLibrary: invalid request")
	}
	if category != "Семейная" && category != "Стратегия" {
		return 0, 0, ErrClubCategoryUnavailable
	}
	return 100 + 100*days, 10, nil
}

func (*ClubLibrary) ListGames() []string {
	return []string{"Первый", "Второй"}
}

type RentalRequest struct {
	GameCategory string
	RentalDays   int
}

type RentalResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type PartnerLibrary struct{}

func (*PartnerLibrary) Calculate(request RentalRequest) (RentalResponse, error) {
	if request.GameCategory == "" || request.RentalDays <= 0 {
		return RentalResponse{}, errors.New("PartnerLibrary: invalid request")
	}
	if request.GameCategory != "Семейная" && request.GameCategory != "Стратегия" {
		return RentalResponse{}, ErrPartnerCategoryUnavailable
	}
	return RentalResponse{
		CostKopecks:     (150+80*request.RentalDays)*100 + 50,
		DurationSeconds: 255,
	}, nil
}

func (*PartnerLibrary) Extend(id string) string {
	return "Аренда продлена: " + id
}
