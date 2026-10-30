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
	ID          string
	Items       []Item
	ServingKind string
	Area        string
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

type Serving interface {
	Cost(area string) (int, error)
}

type Takeaway struct{}

func (Takeaway) Cost(string) (int, error) { return 0, nil }

type TableService struct{}

func (TableService) Cost(area string) (int, error) {
	if area == "" {
		return 0, errors.New("area is required")
	}
	return 100, nil
}

type TerraceService struct{}

func (TerraceService) Cost(area string) (int, error) {
	if area != "Терраса" {
		panic("premium option is unavailable")
	}
	return 250, nil
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

	var option Serving
	switch record.ServingKind {
	case "takeaway":
		option = Takeaway{}
	case "table":
		option = TableService{}
	case "terrace":
		option = TerraceService{}
	default:
		return "", errors.New("unknown option")
	}
	cost, err := option.Cost(record.Area)
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
