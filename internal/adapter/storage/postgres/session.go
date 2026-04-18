package postgres

import (
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"

	"context"

	"github.com/google/uuid"
)

// SessionRepository implements [order.SessionRepository] using postgres.
type SessionRepository struct {
	db *sql.DB
}

var _ order.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository allocates and creates a new [SessionRepository].
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Save(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error) {
	session := &order.Session{
		Id:     uuid.New(),
		Table:  request.Table,
		Status: request.Status,
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO order_sessions(id, table_number, status) VALUES ($1, $2, $3)`,
		session.Id,
		session.Table.Number(),
		session.Status.String(),
	)

	if err != nil {
		return nil, domain.NewInternalError("error inserting session", err)
	}

	return session, nil
}

func (r *SessionRepository) Get(ctx context.Context) ([]order.Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, table_number, status FROM order_sessions`)
	if err != nil {
		return nil, domain.NewInternalError("error getting sessions", err)
	}
	defer rows.Close()

	sessions := make([]order.Session, 0)
	for rows.Next() {
		var id uuid.UUID
		var table int
		var status string

		if err := rows.Scan(&id, &table, &status); err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}

		session, err := order.ParseSession(id, table, status)
		if err != nil {
			return nil, domain.NewInternalError("error parsing session", err)
		}
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

func (r *SessionRepository) Update(ctx context.Context, request *order.UpdateSessionRequest) error {
	var table sql.Null[int]
	if request.Table != nil {
		table.Valid = true
		table.V = request.Table.Number()
	}

	var status sql.NullString
	if request.Status != nil {
		status.Valid = true
		status.String = request.Status.String()
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE order_sessions SET table_number = $1, status = $2 WHERE id = $3`,
		table, status, request.Id,
	)
	if err != nil {
		return domain.NewInternalError("error updating session", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}

	if rows == 0 {
		return domain.NewNotFoundError(
			"order session not found",
			domain.ErrorCodeSessionNotFound,
			domain.ErrorDetail{
				Code:     domain.ErrorCodeSessionNotFoundByID,
				Message:  "order session not found by id",
				Metadata: map[string]any{"id": request.Id},
			})
	}

	return nil
}

func (r *SessionRepository) GetById(ctx context.Context, id uuid.UUID) (*order.Session, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, table_number, status FROM order_sessions WHERE id = $1`,
		id,
	)

	var sessionId uuid.UUID
	var table int
	var status string
	if err := row.Scan(&sessionId, &table, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewNotFoundError("session not found", domain.ErrorCodeSessionNotFound, domain.ErrorDetail{
				Code:     domain.ErrorCodeSessionNotFoundByID,
				Message:  "order session not found by id",
				Metadata: map[string]any{"id": id},
			})
		}

		return nil, domain.NewInternalError("error getting session by id", err)
	}

	session, err := order.ParseSession(sessionId, table, status)
	if err != nil {
		return nil, domain.NewInternalError("error parsing session", err)
	}
	return session, nil
}
