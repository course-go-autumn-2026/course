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

type Appointment struct {
	ID       string
	Items    []Item
	CareKind string
	Size     string
}

type Store interface {
	Save(Appointment) error
	Load(id string) (Appointment, error)
	Delete(id string) error
	Count() int
}

type MemoryStore struct {
	records map[string]Appointment
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Appointment)}
}

func (s *MemoryStore) Save(record Appointment) error {
	s.records[record.ID] = record
	return nil
}

func (s *MemoryStore) Load(id string) (Appointment, error) {
	record, ok := s.records[id]
	if !ok {
		return Appointment{}, errors.New("record not found")
	}
	return record, nil
}

func (s *MemoryStore) Delete(id string) error {
	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Count() int { return len(s.records) }

type CareOption interface {
	Cost(size string) (int, error)
}

type BasicCare struct{}

func (BasicCare) Cost(string) (int, error) { return 0, nil }

type CoatCare struct{}

func (CoatCare) Cost(size string) (int, error) {
	if size == "" {
		return 0, errors.New("size is required")
	}
	return 200, nil
}

type SpaCare struct{}

func (SpaCare) Cost(size string) (int, error) {
	if size != "Крупный" {
		panic("premium option is unavailable")
	}
	return 500, nil
}

type ConsoleNotifier struct{}

func (ConsoleNotifier) Notify(id string) error {
	fmt.Printf("Запись %s: принято\n", id)
	return nil
}

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Book(record Appointment) (string, error) {
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

	var option CareOption
	switch record.CareKind {
	case "basic":
		option = BasicCare{}
	case "coat":
		option = CoatCare{}
	case "spa":
		option = SpaCare{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Size)
	if err != nil {
		return "", err
	}
	total += cost

	receipt := fmt.Sprintf("Запись %s: %d руб.", record.ID, total)
	if err := s.store.Save(record); err != nil {
		return "", err
	}
	notifier := ConsoleNotifier{}
	if err := notifier.Notify(record.ID); err != nil {
		return "", err
	}
	return receipt, nil
}
