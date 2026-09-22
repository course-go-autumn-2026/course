package tabletestdemo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr error
	}{
		{
			name:  "trims, normalizes and removes duplicates",
			input: []string{" Go ", "testing", "go"},
			want:  []string{"go", "testing"},
		},
		{
			name:  "ignores blank tags",
			input: []string{"", "  ", "Go"},
			want:  []string{"go"},
		},
		{
			name:    "rejects empty input",
			input:   nil,
			wantErr: ErrEmptyTags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTags(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
