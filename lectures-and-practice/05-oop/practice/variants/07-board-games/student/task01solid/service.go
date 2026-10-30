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
	ID           string
	Items        []Item
	HandoverKind string
	Point        string
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

type Handover interface {
	Cost(point string) (int, error)
}

type ClubPickup struct{}

func (ClubPickup) Cost(string) (int, error) { return 0, nil }

type LockerPickup struct{}

func (LockerPickup) Cost(point string) (int, error) {
	if point == "" {
		return 0, errors.New("point is required")
	}
	return 200, nil
}

type HomeHandover struct{}

func (HomeHandover) Cost(point string) (int, error) {
	if point != "Центр" {
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

	var option Handover
	switch record.HandoverKind {
	case "club":
		option = ClubPickup{}
	case "locker":
		option = LockerPickup{}
	case "home":
		option = HomeHandover{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Point)
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
