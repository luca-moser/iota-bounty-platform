package controllers

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type Controller interface {
	Init() error
}

var ErrInvalidID = errors.New("invalid object id")
var ErrInvalidModel = errors.New("invalid model")
var ErrInvalidQuery = errors.New("invalid query")
var ErrAlreadyConfirmed = errors.New("user is already confirmed")
var ErrInvalidConfirmationCode = errors.New("invalid confirmation code")
var ErrInvalidModelUpdate = errors.New("invalid model update")
var ErrInvalidModelDeletion = errors.New("invalid model deletion")
var ErrInternalError = errors.New("internal error occurred")
var ErrIssuesDeactivated = errors.New("repository has issues deactivated")
var ErrIssueIsClosed = errors.New("issue is closed")
var ErrIssueDoesntExist = errors.New("issue doesn't exist")
var ErrRepositoryNotInPlatform = errors.New("repository not added to platform")

var DefaultTimeout = time.Duration(5) * time.Second

func DefaultCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}
