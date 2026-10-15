package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	domainredis "devenv/internal/domain/redis"
	"devenv/internal/domain/software"
	"devenv/internal/infra/installer"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

func (s *SoftwareService) RedisDatabases(ctx context.Context, id int64) ([]domainredis.Database, error) {
	connection, err := s.redisConnection(ctx, id)
	if err != nil {
		return nil, err
	}
	databases, err := s.redis.Databases(ctx, connection)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取 Redis 数据库失败")
	}
	return databases, nil
}

func (s *SoftwareService) RedisKeys(ctx context.Context, id int64, query domainredis.KeyQuery) (domainredis.KeyPage, error) {
	if err := validateRedisKeyQuery(query); err != nil {
		return emptyRedisKeyPage(), err
	}
	connection, err := s.redisConnection(ctx, id)
	if err != nil {
		return emptyRedisKeyPage(), err
	}
	result, err := s.redis.Keys(ctx, connection, query)
	if err != nil {
		return emptyRedisKeyPage(), appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取 Redis 键列表失败")
	}
	return result, nil
}

func (s *SoftwareService) RedisValue(ctx context.Context, id int64, query domainredis.ValueQuery) (domainredis.Value, error) {
	if err := validateRedisValueQuery(query); err != nil {
		return emptyRedisValue(), err
	}
	connection, err := s.redisConnection(ctx, id)
	if err != nil {
		return emptyRedisValue(), err
	}
	result, err := s.redis.Value(ctx, connection, query)
	if err != nil {
		return emptyRedisValue(), appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取 Redis 键内容失败")
	}
	return result, nil
}

func (s *SoftwareService) redisConnection(ctx context.Context, id int64) (domainredis.Connection, error) {
	if id <= 0 {
		return domainredis.Connection{}, appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 软件标识无效")
	}
	if s.redis == nil {
		return domainredis.Connection{}, appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "Redis 数据浏览不可用")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return domainredis.Connection{}, err
	}
	if item == nil {
		return domainredis.Connection{}, appErrors.New(appErrors.NOT_FOUND, "Redis 软件不存在")
	}
	if item.Kind != software.KIND_REDIS {
		return domainredis.Connection{}, appErrors.New(appErrors.INVALID_ARGUMENT, "仅 Redis 服务支持数据浏览")
	}
	if item.State != software.STATE_INSTALLED {
		return domainredis.Connection{}, appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "Redis 尚未安装完成")
	}
	if item.Port <= 0 {
		return domainredis.Connection{}, appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 端口无效")
	}
	password, err := redisPassword(*item)
	if err != nil {
		return domainredis.Connection{}, err
	}
	return domainredis.Connection{Host: paths.LOOPBACK, Port: item.Port, Password: password}, nil
}

func validateRedisKeyQuery(query domainredis.KeyQuery) error {
	if query.Database < 0 {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 数据库序号无效")
	}
	if strings.TrimSpace(query.Pattern) == "" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 键匹配规则不能为空")
	}
	if query.Limit <= 0 || query.Limit > domainredis.MAX_KEY_PAGE_SIZE {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 键分页数量无效")
	}
	return nil
}

func validateRedisValueQuery(query domainredis.ValueQuery) error {
	if query.Database < 0 {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 数据库序号无效")
	}
	if query.Key == "" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 键名不能为空")
	}
	if query.Limit <= 0 || query.Limit > domainredis.MAX_VALUE_PAGE_SIZE {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 内容分页数量无效")
	}
	return nil
}

func redisPassword(item software.Software) (string, error) {
	folder := item.Path
	binary := installer.FindRedis(item.Path)
	if binary != "" {
		folder = filepath.Dir(binary)
	}
	content, err := os.ReadFile(filepath.Join(folder, "redis.conf"))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 Redis 配置失败")
	}
	return readValue(software.KIND_REDIS, string(content), "requirepass"), nil
}

func emptyRedisKeyPage() domainredis.KeyPage {
	return domainredis.KeyPage{Keys: make([]domainredis.Key, 0)}
}

func emptyRedisValue() domainredis.Value {
	return domainredis.Value{Entries: make([]domainredis.Entry, 0)}
}
