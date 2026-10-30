package clients

import "errors"

var ErrOfficeFormatUnavailable = errors.New("OfficePrinter: option unavailable")
var ErrQuickFormatUnavailable = errors.New("QuickPrinter: option unavailable")

type OfficePrinter struct{}

func (*OfficePrinter) Estimate(format string, pages int) (int, int, error) {
	if format == "" || pages <= 0 {
		return 0, 0, errors.New("OfficePrinter: invalid request")
	}
	if format != "A4" && format != "A3" {
		return 0, 0, ErrOfficeFormatUnavailable
	}
	return 100 + 10*pages, 15, nil
}

func (*OfficePrinter) ListPrinters() []string {
	return []string{"Первый", "Второй"}
}

type PrintRequest struct {
	PaperFormat string
	PageCount   int
}

type PrintResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type QuickPrinter struct{}

func (*QuickPrinter) Calculate(request PrintRequest) (PrintResponse, error) {
	if request.PaperFormat == "" || request.PageCount <= 0 {
		return PrintResponse{}, errors.New("QuickPrinter: invalid request")
	}
	if request.PaperFormat != "A4" && request.PaperFormat != "A3" {
		return PrintResponse{}, ErrQuickFormatUnavailable
	}
	return PrintResponse{
		CostKopecks:     (150+8*request.PageCount)*100 + 50,
		DurationSeconds: 555,
	}, nil
}

func (*QuickPrinter) Cancel(id string) string {
	return "Печать отменена: " + id
}
