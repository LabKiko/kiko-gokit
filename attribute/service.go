package attribute

import "context"

type Service interface {
	Save(ctx context.Context, kv []*AttributeKV) (string, error)
	UpdateOneByBizId(ctx context.Context, id string, m *AttributeDto) (int64, error)
	SaveBatch(ctx context.Context, batchKv *BatchAttribute) ([]string, error)
	UpdateBatch(ctx context.Context, id string, dto *AttributeDto) (int64, error)
	DeleteAllByBizId(ctx context.Context, id string) (int64, error)
	DeleteAttributeByBizId(ctx context.Context, id string, delKeys []string) (int64, error)
	GetAttributeByBizId(ctx context.Context, id string, path []string) (interface{}, error)
}
