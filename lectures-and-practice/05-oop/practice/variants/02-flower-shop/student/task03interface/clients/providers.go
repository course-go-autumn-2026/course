package clients

import "errors"

var ErrLocalSizeUnavailable = errors.New("LocalFlorist: option unavailable")
var ErrQuickSizeUnavailable = errors.New("QuickFlorist: option unavailable")

type LocalFlorist struct{}

func (*LocalFlorist) Estimate(size string, count int) (int, int, error) {
	if size == "" || count <= 0 {
		return 0, 0, errors.New("LocalFlorist: invalid request")
	}
	if size != "Маленький" && size != "Большой" {
		return 0, 0, ErrLocalSizeUnavailable
	}
	return 200 + 50*count, 20, nil
}

func (*LocalFlorist) ListFlorists() []string {
	return []string{"Первый", "Второй"}
}

type AssemblyRequest struct {
	BouquetSize  string
	BouquetCount int
}

type AssemblyResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type QuickFlorist struct{}

func (*QuickFlorist) Calculate(request AssemblyRequest) (AssemblyResponse, error) {
	if request.BouquetSize == "" || request.BouquetCount <= 0 {
		return AssemblyResponse{}, errors.New("QuickFlorist: invalid request")
	}
	if request.BouquetSize != "Маленький" && request.BouquetSize != "Большой" {
		return AssemblyResponse{}, ErrQuickSizeUnavailable
	}
	return AssemblyResponse{
		CostKopecks:     (250+40*request.BouquetCount)*100 + 50,
		DurationSeconds: 915,
	}, nil
}

func (*QuickFlorist) Track(id string) string {
	return "Букеты собираются: " + id
}
