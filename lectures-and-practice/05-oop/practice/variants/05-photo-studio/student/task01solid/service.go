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

type Booking struct {
	ID           string
	Items        []Item
	BackdropKind string
	Room         string
}

type Store interface {
	Save(Booking) error
	Load(id string) (Booking, error)
	Delete(id string) error
	Count() int
}

type MemoryStore struct {
	records map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Booking)}
}

func (s *MemoryStore) Save(record Booking) error {
	s.records[record.ID] = record
	return nil
}

func (s *MemoryStore) Load(id string) (Booking, error) {
	record, ok := s.records[id]
	if !ok {
		return Booking{}, errors.New("record not found")
	}
	return record, nil
}

func (s *MemoryStore) Delete(id string) error {
	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Count() int { return len(s.records) }

type Backdrop interface {
	Cost(room string) (int, error)
}

type PlainBackdrop struct{}

func (PlainBackdrop) Cost(string) (int, error) { return 0, nil }

type ColorBackdrop struct{}

func (ColorBackdrop) Cost(room string) (int, error) {
	if room == "" {
		return 0, errors.New("room is required")
	}
	return 200, nil
}

type ChromakeyBackdrop struct{}

func (ChromakeyBackdrop) Cost(room string) (int, error) {
	if room != "Студия A" {
		panic("premium option is unavailable")
	}
	return 500, nil
}

type ConsoleNotifier struct{}

func (ConsoleNotifier) Notify(id string) error {
	fmt.Printf("Бронь %s: принято\n", id)
	return nil
}

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Book(record Booking) (string, error) {
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

	var option Backdrop
	switch record.BackdropKind {
	case "plain":
		option = PlainBackdrop{}
	case "color":
		option = ColorBackdrop{}
	case "chroma":
		option = ChromakeyBackdrop{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Room)
	if err != nil {
		return "", err
	}
	total += cost

	receipt := fmt.Sprintf("Бронь %s: %d руб.", record.ID, total)
	if err := s.store.Save(record); err != nil {
		return "", err
	}
	notifier := ConsoleNotifier{}
	if err := notifier.Notify(record.ID); err != nil {
		return "", err
	}
	return receipt, nil
}
