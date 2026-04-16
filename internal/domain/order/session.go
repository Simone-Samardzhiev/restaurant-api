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

func (t SessionTable) Number() int {
	return t.number
}

var (
	StatusOpen   = SessionStatus{"open"}
	StatusClosed = SessionStatus{"closed"}
	StatusPaid   = SessionStatus{"paid"}
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
	case StatusOpen.raw, StatusClosed.raw, StatusPaid.raw:
		return SessionStatus{raw: status}, nil
	default:
		return SessionStatus{}, &domain.ErrorDetail{
			Code:     domain.ErrorCodeInvalidSessionStatus,
			Message:  "invalid session status",
			Metadata: map[string]any{"actual": status, "supported": []string{StatusOpen.raw, StatusClosed.raw, StatusPaid.raw}},
		}
	}
}

// Equal checks if the session is equals to string.
func (s SessionStatus) Equal(status SessionStatus) bool {
	return s.raw == status.raw
}

func (s SessionStatus) String() string {
	return s.raw
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
func ParseSession(id uuid.UUID, table int, status string) (*Session, error) {
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
		return nil, domain.NewValidationError("invalid session", domain.ErrorCodeInvalidSession, errs...)
	}

	return &Session{
		Id:     id,
		Table:  parsedTable,
		Status: parsedStatus,
	}, nil
}

// AddSessionRequest represents a request for adding a new order session.
type AddSessionRequest struct {
	Table  SessionTable
	Status SessionStatus
}

// ParseAddSessionRequest parses [AddSessionRequest] for table number and status.
//
// If any of the fields are invalid the error will be of type [domain.Error].
func ParseAddSessionRequest(table int, status string) (*AddSessionRequest, error) {
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
		return nil, domain.NewValidationError("invalid session", domain.ErrorCodeInvalidSession, errs...)
	}

	return &AddSessionRequest{
		Table:  parsedTable,
		Status: parsedStatus,
	}, nil
}

// UpdateSessionRequest represents a request for updating a session.
type UpdateSessionRequest struct {
	Id     uuid.UUID
	Table  *SessionTable
	Status *SessionStatus
}

// ParseUpdateSessionRequest parses [UpdateSessionRequest] from id, table and status.
//
// If any of the fields is invalid or the update is empty the error will be of type [domain.Error].
func ParseUpdateSessionRequest(id uuid.UUID, table *int, status *string) (*UpdateSessionRequest, error) {
	if table == nil && status == nil {
		return nil, domain.NewBadRequestError("update does not have data", domain.ErrorCodeNoData, nil)
	}

	errs := make([]domain.ErrorDetail, 0)
	var update = &UpdateSessionRequest{
		Id: id,
	}

	if table != nil {
		parsedTable, err := ParseSessionTable(*table)
		if err != nil {
			if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
				errs = append(errs, *detailErr)
			} else {
				return nil, err
			}
		}
		update.Table = &parsedTable
	}

	if status != nil {
		parsedStatus, err := ParseSessionStatus(*status)
		if err != nil {
			if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
				errs = append(errs, *detailErr)
			} else {
				return nil, err
			}
		}
		update.Status = &parsedStatus
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid session update", domain.ErrorCodeInvalidSessionUpdate, errs...)
	}

	return update, nil
}
