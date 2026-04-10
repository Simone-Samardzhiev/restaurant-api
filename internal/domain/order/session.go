package order

import (
	"errors"
	"restaurant/internal/domain"

	"github.com/google/uuid"
)

// SessionTable represents a valid table number that is larger than 0.
type SessionTable struct {
	number int
}

// ParseSessionTable parses a [SessionTable] from a number.
//
// If the number is less than or equals to 0, the returned error will be of type
// [domain.ErrorDetail].
func ParseSessionTable(number int) (SessionTable, error) {
	if number <= 0 {
		return SessionTable{}, &domain.ErrorDetail{
			Code:     domain.ErrorCodeInvalidSessionTable,
			Message:  "invalid session table",
			Metadata: map[string]any{"actual": number, "min": 1},
		}
	}

	return SessionTable{number: number}, nil
}

const (
	OpenSession  string = "open"
	CloseSession string = "closed"
	PaidSession  string = "paid"
)

// SessionStatus represents a valid table status that is either
// open, closed or paid.
type SessionStatus struct {
	raw string
}

// ParseSessionStatus parses [SessionStatus] from string.
//
// If the status is invalid the error will be of type [domain.ErrorDetail].
func ParseSessionStatus(status string) (SessionStatus, error) {
	switch status {
	case OpenSession, CloseSession, PaidSession:
		return SessionStatus{raw: status}, nil
	default:
		return SessionStatus{}, &domain.ErrorDetail{
			Code:     domain.ErrorCodeInvalidSessionStatus,
			Message:  "invalid session status",
			Metadata: map[string]any{"actual": status, "supported": []string{OpenSession, CloseSession, PaidSession}},
		}
	}
}

// Equal checks if the session is equals to string.
func (s SessionStatus) Equal(session string) bool {
	return s.raw == session
}

// Session represents a valid session entity.
type Session struct {
	Id     uuid.UUID
	Table  SessionTable
	Status SessionStatus
}

// ParseSession parses [Session] from table number and status.
//
// If the session is invalid the error will be of type [domain.Error].
func ParseSession(id uuid.UUID, table int, status string) (Session, error) {
	errs := make([]domain.ErrorDetail, 0)
	parsedTable, err := ParseSessionTable(table)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			errs = append(errs, domain.ErrorDetail{})
		}
	}

	parsedStatus, err := ParseSessionStatus(status)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			errs = append(errs, domain.ErrorDetail{})
		}
	}

	if len(errs) > 0 {
		return Session{}, domain.NewValidationError("invalid session", domain.ErrorCodeInvalidSession, errs...)
	}

	return Session{
		Id:     id,
		Table:  parsedTable,
		Status: parsedStatus,
	}, nil
}
