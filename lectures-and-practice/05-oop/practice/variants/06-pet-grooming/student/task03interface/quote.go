package task03interface

import "example.com/oop-practice/06-pet-grooming/student/task03interface/clients"

type CareQuote struct {
	PriceRubles int
	Minutes     int
}

type Reception struct {
	client *clients.LocalGroomer
}

func NewReception(client *clients.LocalGroomer) *Reception {
	return &Reception{client: client}
}

func (s *Reception) QuoteCare(pet string, weightKg int) (CareQuote, error) {
	price, minutes, err := s.client.Estimate(pet, weightKg)
	if err != nil {
		return CareQuote{}, err
	}
	return CareQuote{PriceRubles: price, Minutes: minutes}, nil
}
