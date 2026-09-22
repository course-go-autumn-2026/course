package handler

import (
	"context"
	"errors"
)

type GreetingService struct{}

func (s *GreetingService) Greet(_ context.Context, name string) (string, error) {
	if name == "blocked" {
		return "", errors.New("name is blocked")
	}
	return "Hello, " + name + "!", nil
}
