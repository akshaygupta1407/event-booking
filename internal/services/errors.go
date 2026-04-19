package services

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyUsed   = errors.New("email already in use")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidEventTime   = errors.New("event start_time must be before end_time")
	ErrEventCapacity      = errors.New("insufficient available tickets")
	ErrCapacityTooLow     = errors.New("total tickets cannot be lower than already booked tickets")
)
