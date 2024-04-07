package service

import (
	"context"

	"leaf/dal/model"
)

type SegmentRepo interface {
	GetLeafByTag(ctx context.Context, tag string) (*model.LeafAlloc, error)
	GetAllLeafs(ctx context.Context) ([]*model.LeafAlloc, error)

	UpdateLeaf(ctx context.Context, leaf *model.LeafAlloc) error
	UpdateAndGetLeaf(ctx context.Context, tag string) (*model.LeafAlloc, error)
	UpdateAndGetLeafWithStep(ctx context.Context, tag string, step int64) (*model.LeafAlloc, error)

	ListAllTags(ctx context.Context) ([]string, error)
	CreateAndGetLeaf(ctx context.Context, tag string, step int64) (*model.LeafAlloc, error)

	CleanMaxID(ctx context.Context, tags []string) error
}
