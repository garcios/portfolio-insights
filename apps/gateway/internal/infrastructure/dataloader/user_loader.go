package dataloader

import (
	"context"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/entity"
	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
	"github.com/graph-gophers/dataloader/v7"
)

// UserLoader handles batching and caching of user requests
type UserLoader struct {
	loader *dataloader.Loader[string, *entity.User]
}

// NewUserLoader creates a new initialized UserLoader
func NewUserLoader(userGateway gateway.UserGateway) *UserLoader {
	batchFn := func(ctx context.Context, userIDs []string) []*dataloader.Result[*entity.User] {
		results := make([]*dataloader.Result[*entity.User], len(userIDs))

		// Since we often fetch the same user ID many times (e.g. for holdings),
		// the DataLoader will deduplicate these keys.
		// For the remaining unique keys, we fetch them here.
		// Note: A BatchGetUser(ids) API in UserGateway would be more efficient for multiple *different* users.
		for i, id := range userIDs {
			user, err := userGateway.GetUser(ctx, id)
			results[i] = &dataloader.Result[*entity.User]{
				Data:  user,
				Error: err,
			}
		}
		return results
	}

	return &UserLoader{
		loader: dataloader.NewBatchedLoader(batchFn),
	}
}

// Load fetches a user by ID, using cache and batching
func (l *UserLoader) Load(ctx context.Context, userID string) (*entity.User, error) {
	return l.loader.Load(ctx, userID)()
}
