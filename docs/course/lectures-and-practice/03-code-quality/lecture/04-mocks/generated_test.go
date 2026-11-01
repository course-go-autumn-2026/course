package mocksdemo_test

import (
	"context"
	"errors"
	"testing"

	mocksdemo "code-quality-examples/04-mocks"
	"code-quality-examples/04-mocks/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_GetUserName_GeneratedMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUserRepository(ctrl)
	repository.EXPECT().
		GetUser(gomock.Any(), 1).
		Return(mocksdemo.User{ID: 1, Name: "Alice"}, nil)

	service := mocksdemo.NewService(repository)
	name, err := service.GetUserName(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, "Alice", name)
}

func TestService_GetUserName_RepositoryError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUserRepository(ctrl)
	repository.EXPECT().
		GetUser(gomock.Any(), 7).
		Return(mocksdemo.User{}, databaseErr)

	service := mocksdemo.NewService(repository)
	_, err := service.GetUserName(context.Background(), 7)

	require.Error(t, err)
	assert.ErrorIs(t, err, databaseErr)
}
