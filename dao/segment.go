package dao

import (
	"context"

	"leaf/dal/model"
	"leaf/dal/query"
	"leaf/service"

	"gorm.io/gorm"
)

type SegmentRepoImpl struct {
	db *gorm.DB
}

func (s *SegmentRepoImpl) GetLeafByTag(ctx context.Context, tag string) (*model.LeafAlloc, error) {
	return query.LeafAlloc.WithContext(ctx).Where(query.LeafAlloc.BizTag.Eq(tag)).First()
}

func (s *SegmentRepoImpl) GetAllLeafs(ctx context.Context) ([]*model.LeafAlloc, error) {
	return query.LeafAlloc.WithContext(ctx).Find()
}

func (s *SegmentRepoImpl) UpdateLeaf(ctx context.Context, leaf *model.LeafAlloc) error {
	res, err := query.LeafAlloc.WithContext(ctx).Updates(leaf)
	if err != nil {
		return err
	} else if res.Error != nil {
		return res.Error
	}
	return nil
}

func (s *SegmentRepoImpl) UpdateAndGetLeaf(ctx context.Context, tag string) (leaf *model.LeafAlloc, err error) {
	q := query.Use(s.db)

	err = q.Transaction(func(tx *query.Query) error {
		res, err := query.LeafAlloc.WithContext(ctx).
			Where(query.LeafAlloc.BizTag.Eq(tag)).
			Update(query.LeafAlloc.MaxID, gorm.Expr("max_id + step"))
		if err != nil {
			return err
		} else if res.Error != nil {
			return res.Error
		}

		leaf, err = query.LeafAlloc.WithContext(ctx).
			Where(query.LeafAlloc.BizTag.Eq(tag)).
			First()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return
}

// UpdateAndGetLeafWithStep 按给定的step更新db对应leaf的MaxID
func (s *SegmentRepoImpl) UpdateAndGetLeafWithStep(ctx context.Context, tag string, step int64) (leaf *model.LeafAlloc, err error) {
	q := query.Use(s.db)
	err = q.Transaction(func(tx *query.Query) error {
		res, err := query.LeafAlloc.WithContext(ctx).
			Where(query.LeafAlloc.BizTag.Eq(tag)).
			Update(query.LeafAlloc.MaxID, query.LeafAlloc.MaxID.Add(step))
		if err != nil {
			return err
		} else if res.Error != nil {
			return res.Error
		}

		leaf, err = query.LeafAlloc.WithContext(ctx).Where(query.LeafAlloc.BizTag.Eq(tag)).First()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return leaf, nil
}

func (s *SegmentRepoImpl) ListAllTags(ctx context.Context) ([]string, error) {
	leafs, err := query.LeafAlloc.WithContext(ctx).Select(query.LeafAlloc.BizTag).Find()
	if err != nil {
		return nil, err
	}

	res := make([]string, len(leafs))
	for i, leaf := range leafs {
		res[i] = leaf.BizTag
	}
	return res, nil
}

func (s *SegmentRepoImpl) CreateAndGetLeaf(ctx context.Context, tag string, step int64) (*model.LeafAlloc, error) {
	leaf := &model.LeafAlloc{BizTag: tag, MaxID: step, Step: step}
	if err := query.LeafAlloc.WithContext(ctx).Create(leaf); err != nil {
		return nil, err
	}
	return leaf, nil
}

func (s *SegmentRepoImpl) CleanMaxID(ctx context.Context, tags []string) error {
	panic("implement me!")
}

func NewSegmentRepoImpl(db *gorm.DB) service.SegmentRepo {
	repo := &SegmentRepoImpl{
		db: db,
	}
	return repo
}
