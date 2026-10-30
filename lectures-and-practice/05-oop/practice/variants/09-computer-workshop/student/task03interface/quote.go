package task03interface

import "example.com/oop-practice/09-computer-workshop/student/task03interface/clients"

type ServiceQuote struct {
	PriceRubles int
	Minutes     int
}

type Reception struct {
	client *clients.LocalTechnician
}

func NewReception(client *clients.LocalTechnician) *Reception {
	return &Reception{client: client}
}

func (s *Reception) QuoteService(device string, count int) (ServiceQuote, error) {
	price, minutes, err := s.client.Estimate(device, count)
	if err != nil {
		return ServiceQuote{}, err
	}
	return ServiceQuote{PriceRubles: price, Minutes: minutes}, nil
}
