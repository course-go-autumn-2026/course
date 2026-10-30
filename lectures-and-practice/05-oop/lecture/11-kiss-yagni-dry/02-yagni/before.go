package main

import "fmt"

// Реальное требование: единственный способ доставки, цена 300 рублей.
// Интерфейс, фабрики и реестр добавлены на случай будущих способов доставки.
type BeforeDelivery interface {
	Fee() int
}

type BeforeCourier struct{}

func (BeforeCourier) Fee() int { return 300 }

type DeliveryFactory func() BeforeDelivery

type DeliveryRegistry struct {
	factories map[string]DeliveryFactory
}

func NewDeliveryRegistry() *DeliveryRegistry {
	return &DeliveryRegistry{factories: make(map[string]DeliveryFactory)}
}

func (r *DeliveryRegistry) Register(name string, factory DeliveryFactory) {
	r.factories[name] = factory
}

func (r *DeliveryRegistry) Create(name string) (BeforeDelivery, error) {
	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("неизвестный способ доставки: %s", name)
	}
	return factory(), nil
}
