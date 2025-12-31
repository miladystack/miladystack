package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/miladystack/miladystack/pkg/store/logger/empty"
	"github.com/miladystack/miladystack/pkg/store/where"
)

// DBProvider defines an interface for providing a database connection.
type DBProvider interface {
	// DB returns the database instance for the given context.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
}

// Option defines a function type for configuring the Store.
type Option[T any] func(*Store[T])

// Store represents a generic data store with logging capabilities.
type Store[T any] struct {
	logger  Logger
	storage DBProvider
}

// WithLogger returns an Option function that sets the provided Logger to the Store for logging purposes.
func WithLogger[T any](logger Logger) Option[T] {
	return func(s *Store[T]) {
		s.logger = logger
	}
}

// NewStore creates a new instance of Store with the provided DBProvider.
func NewStore[T any](storage DBProvider, logger Logger) *Store[T] {
	if logger == nil {
		logger = empty.NewLogger()
	}

	return &Store[T]{
		logger:  logger,
		storage: storage,
	}
}

// db retrieves the database instance and applies the provided where conditions.
func (s *Store[T]) db(ctx context.Context, wheres ...where.Where) *gorm.DB {
	dbInstance := s.storage.DB(ctx)
	for _, whr := range wheres {
		if whr != nil {
			dbInstance = whr.Where(dbInstance)
		}
	}
	return dbInstance
}

// Create inserts a new object into the database.
func (s *Store[T]) Create(ctx context.Context, obj *T) error {
	if err := s.db(ctx).Create(obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to insert object into database", "object", obj)
		return err
	}
	return nil
}

// Update modifies an existing object in the database.
func (s *Store[T]) Update(ctx context.Context, obj *T) error {
	if err := s.db(ctx).Save(obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to update object in database", "object", obj)
		return err
	}
	return nil
}

// Delete removes an object from the database based on the provided where options.
func (s *Store[T]) Delete(ctx context.Context, opts *where.Options) error {
	err := s.db(ctx, opts).Delete(new(T)).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error(ctx, err, "Failed to delete object from database", "conditions", opts)
		return err
	}
	return nil
}

// Get retrieves a single object from the database based on the provided where options.
func (s *Store[T]) Get(ctx context.Context, opts *where.Options) (*T, error) {
	var obj T
	if err := s.db(ctx, opts).First(&obj).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to retrieve object from database", "conditions", opts)
		return nil, err
	}
	return &obj, nil
}

// List retrieves a list of objects from the database based on the provided where options.
func (s *Store[T]) List(ctx context.Context, orderStr string, isAsc bool, page, pageSize int, opts *where.Options) (count int64, ret []*T, err error) {
	// 根据 isAsc 参数确定排序方式
	sortDirection := "ASC"
	if !isAsc {
		sortDirection = "DESC"
	}

	// 如果用户未指定排序字段，则默认使用 id
	if orderStr == "" {
		orderStr = fmt.Sprintf("id %s", sortDirection)
	} else {
		orderStr = strings.TrimSpace(orderStr)
		orderStr = fmt.Sprintf("%s %s", orderStr, sortDirection)
	}

	// 计算分页偏移量
	offset := (page - 1) * pageSize

	// 构建查询：先统计总数，再查询分页数据
	db := s.db(ctx, opts)

	// 第一步：统计符合条件的总条数（不受分页影响）
	if err = db.Model(new(T)).Count(&count).Error; err != nil {
		s.logger.Error(ctx, err, "Failed to count objects from database", "conditions", opts)
		return
	}

	// 第二步：查询分页数据（使用 pageSize 限制条数）
	// 处理边界情况：pageSize <= 0 时返回空列表（避免查询全部数据）
	if pageSize > 0 {
		err = db.Order(orderStr).Offset(offset).Limit(pageSize).Find(&ret).Error
	} else {
		ret = []*T{}
		err = nil
	}

	if err != nil {
		s.logger.Error(ctx, err, "Failed to list objects from database", "conditions", opts)
	}
	return
}
