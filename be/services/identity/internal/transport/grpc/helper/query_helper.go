package helper

import (
	"github.com/kainguyen/badminton-hub/be/services/identity/internal/model"
)

// ListQueryRequest is the common filter projection of user and staff list requests.
type ListQueryRequest interface {
	GetLimit() int32
	GetOffset() int32
	GetSearch() string
	GetCreatedFrom() string
	GetCreatedTo() string
}

func ParseListQuery(request ListQueryRequest) (model.ListQuery, error) {
	return model.NewListQuery(
		int(request.GetLimit()),
		int(request.GetOffset()),
		request.GetSearch(),
		request.GetCreatedFrom(),
		request.GetCreatedTo(),
	)
}
