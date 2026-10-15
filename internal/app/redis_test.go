package app

import (
	"context"
	"testing"

	domainredis "devenv/internal/domain/redis"
	"devenv/internal/domain/software"
	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

type redisBrowserFixture struct {
	databases  []domainredis.Database
	page       domainredis.KeyPage
	value      domainredis.Value
	connection domainredis.Connection
	keyQuery   domainredis.KeyQuery
	valueQuery domainredis.ValueQuery
}

func (fixture *redisBrowserFixture) Databases(_ context.Context, connection domainredis.Connection) ([]domainredis.Database, error) {
	fixture.connection = connection
	return fixture.databases, nil
}

func (fixture *redisBrowserFixture) Keys(_ context.Context, connection domainredis.Connection, query domainredis.KeyQuery) (domainredis.KeyPage, error) {
	fixture.connection = connection
	fixture.keyQuery = query
	return fixture.page, nil
}

func (fixture *redisBrowserFixture) Value(_ context.Context, connection domainredis.Connection, query domainredis.ValueQuery) (domainredis.Value, error) {
	fixture.connection = connection
	fixture.valueQuery = query
	return fixture.value, nil
}

func TestRedisDataReadsManagedRedisOnly(t *testing.T) {
	base := t.TempDir()
	database, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	ctx := context.Background()
	repo := sqlite.NewSoftwareRepository(database.DB)
	redisItem := software.Software{
		Kind:    software.KIND_REDIS,
		Version: "8.10.2",
		Path:    t.TempDir(),
		State:   software.STATE_INSTALLED,
		Port:    6379,
	}
	if err := repo.Create(ctx, &redisItem); err != nil {
		t.Fatal(err)
	}
	phpItem := software.Software{
		Kind:    software.KIND_PHP,
		Version: "8.3.35",
		Path:    t.TempDir(),
		State:   software.STATE_INSTALLED,
	}
	if err := repo.Create(ctx, &phpItem); err != nil {
		t.Fatal(err)
	}
	browser := &redisBrowserFixture{
		databases: []domainredis.Database{{Index: 0, Keys: 2, Expires: 1}},
		page:      domainredis.KeyPage{Keys: []domainredis.Key{{Name: "cache:home", Type: "string", TTL: 60}}},
		value:     domainredis.Value{Type: "string", Text: "ok", Entries: make([]domainredis.Entry, 0)},
	}
	service := &SoftwareService{repo: repo, redis: browser}

	databases, err := service.RedisDatabases(ctx, redisItem.ID)
	if err != nil || len(databases) != 1 || databases[0].Keys != 2 {
		t.Fatalf("unexpected Redis databases: %+v, %v", databases, err)
	}
	page, err := service.RedisKeys(ctx, redisItem.ID, domainredis.KeyQuery{Database: 0, Pattern: "cache:*", Limit: 100})
	if err != nil || len(page.Keys) != 1 || browser.keyQuery.Pattern != "cache:*" {
		t.Fatalf("unexpected Redis keys: %+v, %+v, %v", page, browser.keyQuery, err)
	}
	value, err := service.RedisValue(ctx, redisItem.ID, domainredis.ValueQuery{Database: 0, Key: "cache:home", Limit: 100})
	if err != nil || value.Text != "ok" || browser.valueQuery.Key != "cache:home" {
		t.Fatalf("unexpected Redis value: %+v, %+v, %v", value, browser.valueQuery, err)
	}
	if browser.connection.Host != "127.0.0.1" || browser.connection.Port != 6379 {
		t.Fatalf("unexpected Redis connection: %+v", browser.connection)
	}
	if _, err := service.RedisKeys(ctx, redisItem.ID, domainredis.KeyQuery{Database: 0, Pattern: "", Limit: 100}); err == nil {
		t.Fatal("accepted empty key pattern")
	}
	if _, err := service.RedisDatabases(ctx, phpItem.ID); err == nil {
		t.Fatal("accepted non-Redis software")
	}
}
