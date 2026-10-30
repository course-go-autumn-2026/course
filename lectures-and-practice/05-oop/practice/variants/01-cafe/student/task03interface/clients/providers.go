package clients

import "errors"

var ErrBaristaDrinkUnavailable = errors.New("Barista: option unavailable")
var ErrAutoRecipeUnavailable = errors.New("AutoBar: option unavailable")

type Barista struct{}

func (*Barista) Estimate(drink string, cups int) (int, int, error) {
	if drink == "" || cups <= 0 {
		return 0, 0, errors.New("Barista: invalid request")
	}
	if drink != "Американо" && drink != "Латте" {
		return 0, 0, ErrBaristaDrinkUnavailable
	}
	return 100 + 50*cups, 10, nil
}

func (*Barista) ListBaristas() []string {
	return []string{"Первый", "Второй"}
}

type BrewRequest struct {
	Recipe   string
	CupCount int
}

type BrewResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type AutoBar struct{}

func (*AutoBar) Calculate(request BrewRequest) (BrewResponse, error) {
	if request.Recipe == "" || request.CupCount <= 0 {
		return BrewResponse{}, errors.New("AutoBar: invalid request")
	}
	if request.Recipe != "Американо" && request.Recipe != "Латте" {
		return BrewResponse{}, ErrAutoRecipeUnavailable
	}
	return BrewResponse{
		CostKopecks:     (120+40*request.CupCount)*100 + 50,
		DurationSeconds: 375,
	}, nil
}

func (*AutoBar) Clean(id string) string {
	return "Очистка завершена: " + id
}
