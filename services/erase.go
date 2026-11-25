package services

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/repositories"
	"9itchi-tracker-api/utils"

	"go.uber.org/zap"
)

// EraseService removes what this service keeps about a user whose account is being deleted.
type EraseService interface {
	// DELETE /internal/users/:id (internal): called by account-api before it deletes the account.
	EraseUser(ctx context.Context, userID string) *models.ApiError
}

type eraseService struct {
	repo repositories.EraseRepository
}

func NewEraseService(repo repositories.EraseRepository) EraseService {
	return &eraseService{repo: repo}
}

var eraseUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// EraseUser is safe to call again: account-api retries until every service has answered. Only the number of
// rows is logged, never whose they were.
func (s *eraseService) EraseUser(ctx context.Context, userID string) *models.ApiError {
	if !eraseUUID.MatchString(userID) {
		return &models.ApiError{Code: http.StatusBadRequest, Error: errors.New("invalid user id")}
	}
	if _, err := s.repo.EraseUser(ctx, userID); err != nil {
		utils.Logger.Error("erase user", zap.Error(err))
		return &models.ApiError{Code: http.StatusInternalServerError, Error: errors.New("internal error")}
	}
	return nil
}
