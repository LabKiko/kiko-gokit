package attribute

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LabKiko/kiko-gokit/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const BizId = "biz_id"

type MgoAttribute struct {
	mgo  *mongo.Database
	coll *mongo.Collection
}

func NewMongoAttribute(m *mongo.Database, c string) *MgoAttribute {
	return &MgoAttribute{
		mgo:  m,
		coll: m.Collection(c),
	}
}
func (s *MgoAttribute) UseCollection(c string) *MgoAttribute {
	s.coll = s.mgo.Collection(c)
	return s
}

func (s *MgoAttribute) Save(ctx context.Context, kv []*AttributeKV) (string, error) {

	one, err := s.coll.InsertOne(ctx, convertBatch(kv))
	if err != nil {
		return "", err
	}
	objectID, ok := one.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("attributeService: id can not convert ObjectID")
	}
	return objectID.Hex(), nil
}

func (s *MgoAttribute) UpdateOneByBizId(ctx context.Context, id string, m *AttributeDto) (int64, error) {
	updateRet, err := s.coll.UpdateOne(ctx,
		bson.D{
			bson.E{
				Key:   BizId,
				Value: id,
			},
		}, bson.D{{"$set", m}})
	if err != nil {
		return 0, err
	}

	return updateRet.ModifiedCount, nil
}
func (s *MgoAttribute) SaveBatch(ctx context.Context, batchKv *BatchAttribute) ([]string, error) {
	var b []interface{}
	for _, batch := range batchKv.BatchKv {
		b = append(b, convertBatch(batch.KV))
	}
	many, err := s.coll.InsertMany(ctx, b)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, d := range many.InsertedIDs {
		if b, ok := d.(primitive.ObjectID); ok {
			ids = append(ids, b.Hex())
		}
	}
	return ids, nil
}
func (s *MgoAttribute) UpdateBatch(ctx context.Context, id string, dto *AttributeDto) (int64, error) {

	many, err := s.coll.UpdateMany(ctx, bson.D{
		{BizId, id},
	}, bson.D{
		{"$set", convertBatch(dto.KV)},
	})
	if err != nil {
		return 0, err
	}
	logger.WithContext(ctx).Infof("MgoAttributeBackup.SetProperty success ret:%+v", many)
	return many.ModifiedCount, nil
}
func (s *MgoAttribute) DeleteAllByBizId(ctx context.Context, id string) (int64, error) {

	one, err := s.coll.DeleteOne(ctx, bson.D{
		{BizId, id},
	})
	if err != nil {
		return 0, err
	}

	return one.DeletedCount, nil
}
func (s *MgoAttribute) DeleteAttributeByBizId(ctx context.Context, id string, delKeys []string) (int64, error) {
	var del bson.D
	for _, key := range delKeys {
		del = append(del, bson.E{Key: key, Value: 1})
	}

	result, err := s.coll.UpdateOne(ctx, bson.D{
		{BizId, id},
	}, bson.D{
		{"$unset", del},
	})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// GetProperty

func (s *MgoAttribute) GetAttributeByBizId(ctx context.Context, id string, path []string) (interface{}, error) {

	filter := bson.M{
		BizId: id,
	}
	_path := strings.Join(path, ".")
	b := bson.D{
		bson.E{
			Key:   _path,
			Value: 1,
		}}

	opt := options.FindOne()
	if len(b) > 0 {
		opt.SetProjection(b)
	}
	findRet := s.coll.FindOne(ctx, filter, opt)
	var result map[string]interface{}
	if err := findRet.Decode(&result); err != nil {
		return "", err
	}
	for i := 0; i < len(path); i++ {

		if r, ok := result[path[i]]; !ok {
			return nil, errors.New("can not find path " + path[i])
		} else {
			_, yes := r.(map[string]interface{})
			if !yes {
				return nil, errors.New(fmt.Sprintf("can not convert to map"))
			}
		}

	}
	return result, nil
}
func (s *MgoAttribute) ListByBizIds(ctx context.Context, ids []string) ([]map[string]interface{}, error) {

	filter := bson.D{{
		BizId, bson.D{
			{"$in", ids},
		},
	},
	}

	findRet, err := s.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if e := findRet.All(ctx, &result); e != nil {
		return nil, e
	}

	return result, nil
}
func (s *MgoAttribute) GetAssetAttribute(ctx context.Context, id string) (map[string]interface{}, error) {

	filter := bson.M{
		BizId: id,
	}

	findRet := s.coll.FindOne(ctx, filter)
	var result = make(map[string]interface{})
	if err := findRet.Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.Infof("mongodb id %s no documents")
			return result, nil
		}
		return nil, err
	}

	return result, nil
}

// GetProperty

func (s *MgoAttribute) Page(ctx context.Context, filter bson.D, p *Page) (PageResult, error) {
	optFind := &options.FindOptions{}
	var projection []bson.E
	// 设置返回字段 默认全部返回
	if len(p.Field) > 0 {
		for _, v := range p.Field {
			projection = append(projection, primitive.E{
				Key:   v,
				Value: 1,
			})

		}
		optFind.SetProjection(projection)
	}
	// 设置排序字段
	var sort []bson.E
	if len(p.S) > 0 {
		for _, _s := range p.S {
			if _s.Key == "" {
				continue
			}
			_value := 1
			if _s.Desc {
				_value = -1
			}
			sort = append(sort, primitive.E{
				Key:   _s.Key,
				Value: _value,
			})

		}
		optFind.SetSort(sort)
	}
	// 设置分页参数
	if p.PageNo != 0 || p.PageSize != 0 {
		skip := (p.PageNo - 1) * p.PageSize
		optFind.SetLimit(p.PageSize)
		optFind.SetSkip(skip)
	}
	result := PageResult{}
	documents, err := s.coll.CountDocuments(ctx, filter)
	if err != nil {
		return result, err
	}

	cursor, err := s.coll.Find(ctx, filter, optFind)

	if err != nil {
		return result, err
	}
	// 将结果转换成 map
	var _m []map[string]interface{}
	if err := cursor.All(ctx, &_m); err != nil {
		return result, err
	}

	result.Total = documents
	result.Data = _m

	return result, nil
}

func convertBatch(batch []*AttributeKV) bson.D {
	var b bson.D
	for _, kv := range batch {
		b = append(b,
			bson.E{
				Key:   kv.Key,
				Value: kv.Value,
			})
	}
	return b
}