package task01solid

import (
	"errors"
	"fmt"
)

type Item struct {
	Name     string
	Price    int
	Quantity int
}

type Rental struct {
	ID         string
	Items      []Item
	ReturnKind string
	Station    string
}

type Store interface {
	Save(Rental) error
	Load(id string) (Rental, error)
	Delete(id string) error
	Count() int
}

type MemoryStore struct {
	records map[string]Rental
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Rental)}
}

func (s *MemoryStore) Save(record Rental) error {
	s.records[record.ID] = record
	return nil
}

func (s *MemoryStore) Load(id string) (Rental, error) {
	record, ok := s.records[id]
	if !ok {
		return Rental{}, errors.New("record not found")
	}
	return record, nil
}

func (s *MemoryStore) Delete(id string) error {
	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Count() int { return len(s.records) }

type ReturnOption interface {
	Cost(station string) (int, error)
}

type SameStation struct{}

func (SameStation) Cost(string) (int, error) { return 0, nil }

type OtherStation struct{}

func (OtherStation) Cost(station string) (int, error) {
	if station == "" {
		return 0, errors.New("station is required")
	}
	return 200, nil
}

type CourierReturn struct{}

func (CourierReturn) Cost(station string) (int, error) {
	if station != "Центр" {
		panic("premium option is unavailable")
	}
	return 500, nil
}

type ConsoleNotifier struct{}

func (ConsoleNotifier) Notify(id string) error {
	fmt.Printf("Аренда %s: принято\n", id)
	return nil
}

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Rent(record Rental) (string, error) {
	if record.ID == "" || len(record.Items) == 0 {
		return "", errors.New("ID and items are required")
	}

	total := 0
	for _, item := range record.Items {
		if item.Price < 0 || item.Quantity <= 0 {
			return "", errors.New("invalid item")
		}
		total += item.Price * item.Quantity
	}

	var option ReturnOption
	switch record.ReturnKind {
	case "same":
		option = SameStation{}
	case "other":
		option = OtherStation{}
	case "courier":
		option = CourierReturn{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Station)
	if err != nil {
		return "", err
	}
	total += cost

	receipt := fmt.Sprintf("Аренда %s: %d руб.", record.ID, total)
	if err := s.store.Save(record); err != nil {
		return "", err
	}
	notifier := ConsoleNotifier{}
	if err := notifier.Notify(record.ID); err != nil {
		return "", err
	}
	return receipt, nil
}
