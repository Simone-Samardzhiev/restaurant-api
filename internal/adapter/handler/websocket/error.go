package websocket

import (
	"encoding/json"
	"errors"
	"restaurant/internal/adapter/handler/translator"
	"restaurant/internal/domain"

	"go.uber.org/zap"
)

// handleInvalidJSON translates the error into [Message] and encodes it into JSON.
func handleInvalidJSON(err error) []byte {
	var message Message
	message.Event = EventAddSession

	data, err := json.Marshal(translator.DomainError(domain.NewBadRequestError("invalid json", domain.ErrorCodeMalformedRequest, err)))
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}
	return body
}

// handleDomainError translates [domain.Error] into [Message] and encodes it into JSON.
//
// If the error is not of type [domain.Error], the error will be logged.
func handleDomainError(err error) []byte {
	var message Message
	message.Event = EventError

	domainErr, ok := errors.AsType[*domain.Error](err)
	if !ok {
		zap.L().Error("unknown error", zap.Error(err))
	}

	data, err := json.Marshal(translator.DomainError(domainErr))
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}
	return body
}

// handleInvalidEvent translated invalid event into [Message] and encodes it into JSON.
func handleInvalidEvent() []byte {
	var message Message
	message.Event = EventError

	data, err := json.Marshal(translator.ErrorResponse{
		Code:    "INVALID_EVENT",
		Message: "Event is not supported.",
		Details: nil,
	})
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}
	return body
}
