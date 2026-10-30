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

type Request struct {
	ID              string
	Items           []Item
	DiagnosticsKind string
	Device          string
}

type Store interface {
	Save(Request) error
	Load(id string) (Request, error)
	Delete(id string) error
	Count() int
}

type MemoryStore struct {
	records map[string]Request
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Request)}
}

func (s *MemoryStore) Save(record Request) error {
	s.records[record.ID] = record
	return nil
}

func (s *MemoryStore) Load(id string) (Request, error) {
	record, ok := s.records[id]
	if !ok {
		return Request{}, errors.New("record not found")
	}
	return record, nil
}

func (s *MemoryStore) Delete(id string) error {
	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Count() int { return len(s.records) }

type Diagnostics interface {
	Cost(device string) (int, error)
}

type NoDiagnostics struct{}

func (NoDiagnostics) Cost(string) (int, error) { return 0, nil }

type StandardDiagnostics struct{}

func (StandardDiagnostics) Cost(device string) (int, error) {
	if device == "" {
		return 0, errors.New("device is required")
	}
	return 200, nil
}

type AdvancedDiagnostics struct{}

func (AdvancedDiagnostics) Cost(device string) (int, error) {
	if device != "Ноутбук" {
		panic("premium option is unavailable")
	}
	return 500, nil
}

type ConsoleNotifier struct{}

func (ConsoleNotifier) Notify(id string) error {
	fmt.Printf("Заявка %s: принято\n", id)
	return nil
}

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Accept(record Request) (string, error) {
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

	var option Diagnostics
	switch record.DiagnosticsKind {
	case "none":
		option = NoDiagnostics{}
	case "standard":
		option = StandardDiagnostics{}
	case "advanced":
		option = AdvancedDiagnostics{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Device)
	if err != nil {
		return "", err
	}
	total += cost

	receipt := fmt.Sprintf("Заявка %s: %d руб.", record.ID, total)
	if err := s.store.Save(record); err != nil {
		return "", err
	}
	notifier := ConsoleNotifier{}
	if err := notifier.Notify(record.ID); err != nil {
		return "", err
	}
	return receipt, nil
}
