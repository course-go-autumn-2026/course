package tabletest

import "testing"

func TestSumSquares(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    int
		wantErr error
	}{
		// TODO: добавьте случаи для обычного ввода, пустого слайса и ошибки.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SumSquares(tt.numbers)

			// TODO: проверьте got и err. Для обёрнутой ошибки используйте errors.Is.
			_, _ = got, err
		})
	}
}
