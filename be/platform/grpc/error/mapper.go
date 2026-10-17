package grpcError

import (
	"errors"

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

func NewErrorMapper(fallbackCode codes.Code, fallbackMessage string, rules ...ErrorRule) ErrorMapper {
	return ErrorMapper{rules: rules, fallbackCode: fallbackCode, fallbackMessage: fallbackMessage}
}

func (m ErrorMapper) Map(err error) error {
	if err == nil {
		return nil
	}
	for _, rule := range m.rules {
		if rule.match(err) {
			return status.Error(rule.code, rule.message)
		}
	}
	return status.Error(m.fallbackCode, m.fallbackMessage)
}
