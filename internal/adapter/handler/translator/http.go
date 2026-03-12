package translator

import (
	"net/http"
	"restaurant/internal/domain"
)

var mapErrorKindHTTPStatus = map[domain.ErrorKind]int{
	domain.ErrorKindInternal:   http.StatusInternalServerError,
	domain.ErrorKindValidation: http.StatusUnprocessableEntity,
	domain.ErrorKindConflict:   http.StatusConflict,
	domain.ErrorKindBadRequest: http.StatusBadRequest,
	domain.ErrorKindNotFound:   http.StatusNotFound,
}

// ErrorKindToHTTPStatus maps [domain.ErrorKind] to HTTP status code.
func ErrorKindToHTTPStatus(kind domain.ErrorKind) int {
	if val, ok := mapErrorKindHTTPStatus[kind]; ok {
		return val
	}
	return http.StatusInternalServerError
}
