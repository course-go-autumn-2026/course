package clients

import "errors"

var ErrLocalPetUnavailable = errors.New("LocalGroomer: option unavailable")
var ErrQuickPetUnavailable = errors.New("QuickGroomer: option unavailable")

type LocalGroomer struct{}

func (*LocalGroomer) Estimate(pet string, weightKg int) (int, int, error) {
	if pet == "" || weightKg <= 0 {
		return 0, 0, errors.New("LocalGroomer: invalid request")
	}
	if pet != "Кошка" && pet != "Собака" {
		return 0, 0, ErrLocalPetUnavailable
	}
	return 200 + 100*weightKg, 60, nil
}

func (*LocalGroomer) ListGroomers() []string {
	return []string{"Первый", "Второй"}
}

type CareRequest struct {
	PetType string
	MassKg  int
}

type CareResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type QuickGroomer struct{}

func (*QuickGroomer) Calculate(request CareRequest) (CareResponse, error) {
	if request.PetType == "" || request.MassKg <= 0 {
		return CareResponse{}, errors.New("QuickGroomer: invalid request")
	}
	if request.PetType != "Кошка" && request.PetType != "Собака" {
		return CareResponse{}, ErrQuickPetUnavailable
	}
	return CareResponse{
		CostKopecks:     (300+90*request.MassKg)*100 + 50,
		DurationSeconds: 2655,
	}, nil
}

func (*QuickGroomer) Cancel(id string) string {
	return "Запись отменена: " + id
}
