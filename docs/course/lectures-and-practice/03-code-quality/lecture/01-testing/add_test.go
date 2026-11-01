package testingdemo

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	want := 5

	if result != want {
		t.Errorf("Add(2, 3) = %d; want %d", result, want)
	}

	t.Log("тест сложения завершён")
}
