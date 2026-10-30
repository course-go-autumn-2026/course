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

type Order struct {
	ID         string
	Items      []Item
	FinishKind string
	Box        string
}

type Store interface {
	Save(Order) error
	Load(id string) (Order, error)
	Delete(id string) error
	Count() int
}

type MemoryStore struct {
	records map[string]Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Order)}
}

func (s *MemoryStore) Save(record Order) error {
	s.records[record.ID] = record
	return nil
}

func (s *MemoryStore) Load(id string) (Order, error) {
	record, ok := s.records[id]
	if !ok {
		return Order{}, errors.New("record not found")
	}
	return record, nil
}

func (s *MemoryStore) Delete(id string) error {
	delete(s.records, id)
	return nil
}

func (s *MemoryStore) Count() int { return len(s.records) }

type Finish interface {
	Cost(box string) (int, error)
}

type NoFinish struct{}

func (NoFinish) Cost(string) (int, error) { return 0, nil }

type WaxFinish struct{}

func (WaxFinish) Cost(box string) (int, error) {
	if box == "" {
		return 0, errors.New("box is required")
	}
	return 200, nil
}

type CeramicFinish struct{}

func (CeramicFinish) Cost(box string) (int, error) {
	if box != "Бокс 1" {
		panic("premium option is unavailable")
	}
	return 500, nil
}

type ConsoleNotifier struct{}

func (ConsoleNotifier) Notify(id string) error {
	fmt.Printf("Заказ %s: принято\n", id)
	return nil
}

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) PlaceOrder(record Order) (string, error) {
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

	var option Finish
	switch record.FinishKind {
	case "none":
		option = NoFinish{}
	case "wax":
		option = WaxFinish{}
	case "ceramic":
		option = CeramicFinish{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Box)
	if err != nil {
		return "", err
	}
	total += cost

	receipt := fmt.Sprintf("Заказ %s: %d руб.", record.ID, total)
	if err := s.store.Save(record); err != nil {
		return "", err
	}
	notifier := ConsoleNotifier{}
	if err := notifier.Notify(record.ID); err != nil {
		return "", err
	}
	return receipt, nil
}
