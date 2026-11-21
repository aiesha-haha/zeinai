package domain

import "context"

type SchoolRepository interface {
	List(ctx context.Context, regionID int, limit, offset int) ([]School, error)
	Search(ctx context.Context, query string, limit int) ([]School, error)
}
