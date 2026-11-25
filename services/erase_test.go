package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeErase struct {
	ids []string
	err error
}

func (f *fakeErase) EraseUser(_ context.Context, id string) (int64, error) {
	f.ids = append(f.ids, id)
	return 3, f.err
}

func TestEraseUser(t *testing.T) {
	const id = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
	repo := &fakeErase{}
	svc := NewEraseService(repo)
	require.Nil(t, svc.EraseUser(context.Background(), id))
	require.Nil(t, svc.EraseUser(context.Background(), id), "erasing again is not an error")
	assert.Equal(t, []string{id, id}, repo.ids)

	apiErr := svc.EraseUser(context.Background(), "nope")
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.Code)
	assert.Len(t, repo.ids, 2, "a bad id never reaches the database")

	repo.err = assert.AnError
	apiErr = svc.EraseUser(context.Background(), id)
	require.NotNil(t, apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.Code)
	assert.EqualError(t, apiErr.Error, "internal error", "the cause is not sent back")
}
