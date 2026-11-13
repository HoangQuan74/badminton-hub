package grpcError

import (
	"errors"
	"google.golang.org/protobuf/protoadapt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorRule struct {
	match   func(error) bool
	code    codes.Code
	message string
}

func MatchError(target error, code codes.Code, message string) ErrorRule {
	return ErrorRule{
		match: func(err error) bool {
			return errors.Is(err, target)
		},
		code:    code,
		message: message,
	}
}

type ErrorMapper struct {
	rules           []ErrorRule
	fallbackCode    codes.Code
	fallbackMessage string
}

type detailedError struct {
	error
	details []protoadapt.MessageV1
}

func (e *detailedError) Unwrap() error { return e.error }

// WithDetails attaches transport-owned status details without changing domain errors.
func WithDetails(err error, details ...protoadapt.MessageV1) error {
	if err == nil {
		return nil
	}
	return &detailedError{error: err, details: details}
}

func NewErrorMapper(fallbackCode codes.Code, fallbackMessage string, rules ...ErrorRule) ErrorMapper {
	return ErrorMapper{rules: rules, fallbackCode: fallbackCode, fallbackMessage: fallbackMessage}
}

func (m ErrorMapper) Map(err error) error {
	if err == nil {
		return nil
	}
	for _, rule := range m.rules {
		if rule.match(err) {
			return mappedStatus(err, rule.code, rule.message)
		}
	}
	return mappedStatus(err, m.fallbackCode, m.fallbackMessage)
}

func mappedStatus(err error, code codes.Code, message string) error {
	value := status.New(code, message)
	var detailed *detailedError
	if errors.As(err, &detailed) {
		withDetails, detailError := value.WithDetails(detailed.details...)
		if detailError == nil {
			return withDetails.Err()
		}
	}
	return value.Err()
}
