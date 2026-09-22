package mocksdemo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	getUser func(context.Context, int) (User, error)
}

func (s repositoryStub) GetUser(ctx context.Context, id int) (User, error) {
	return s.getUser(ctx, id)
}

func TestService_GetUserName_ManualStub(t *testing.T) {
	repository := repositoryStub{
		getUser: func(_ context.Context, id int) (User, error) {
			require.Equal(t, 1, id)
			return User{ID: id, Name: "Alice"}, nil
		},
	}

	service := NewService(repository)
	name, err := service.GetUserName(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, "Alice", name)
}
