package redis

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	domainredis "devenv/internal/domain/redis"
	redisclient "github.com/redis/go-redis/v9"
)

const QUERY_TIMEOUT = 5 * time.Second

type Browser struct{}

func NewBrowser() *Browser {
	return &Browser{}
}

func (browser *Browser) Databases(ctx context.Context, connection domainredis.Connection) ([]domainredis.Database, error) {
	operation, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	client := browser.open(connection, 0)
	defer client.Close()
	if err := client.Ping(operation).Err(); err != nil {
		return nil, err
	}
	info, err := client.Info(operation, "keyspace").Result()
	if err != nil {
		return nil, err
	}
	databases := parseDatabases(info)
	if len(databases) == 0 {
		databases = append(databases, domainredis.Database{Index: 0, Keys: 0, Expires: 0})
	}
	return databases, nil
}

func (browser *Browser) Keys(ctx context.Context, connection domainredis.Connection, query domainredis.KeyQuery) (domainredis.KeyPage, error) {
	operation, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	client := browser.open(connection, query.Database)
	defer client.Close()
	keys, cursor, err := client.Scan(operation, query.Cursor, query.Pattern, query.Limit).Result()
	if err != nil {
		return domainredis.KeyPage{Keys: make([]domainredis.Key, 0)}, err
	}
	pipeline := client.Pipeline()
	types := make([]*redisclient.StatusCmd, 0, len(keys))
	ttls := make([]*redisclient.DurationCmd, 0, len(keys))
	for _, key := range keys {
		types = append(types, pipeline.Type(operation, key))
		ttls = append(ttls, pipeline.TTL(operation, key))
	}
	if _, err := pipeline.Exec(operation); err != nil && err != redisclient.Nil {
		return domainredis.KeyPage{Keys: make([]domainredis.Key, 0)}, err
	}
	result := make([]domainredis.Key, 0, len(keys))
	for index, key := range keys {
		typeName, err := types[index].Result()
		if err != nil || typeName == "none" {
			continue
		}
		ttl, err := ttls[index].Result()
		if err != nil {
			continue
		}
		result = append(result, domainredis.Key{Name: key, Type: typeName, TTL: ttlSeconds(ttl)})
	}
	return domainredis.KeyPage{Cursor: cursor, Keys: result}, nil
}

func (browser *Browser) Value(ctx context.Context, connection domainredis.Connection, query domainredis.ValueQuery) (domainredis.Value, error) {
	operation, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	client := browser.open(connection, query.Database)
	defer client.Close()
	typeName, err := client.Type(operation, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	if typeName == "none" {
		return emptyValue(), fmt.Errorf("键已不存在")
	}
	ttl, err := client.TTL(operation, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	result := domainredis.Value{Type: typeName, TTL: ttlSeconds(ttl), Entries: make([]domainredis.Entry, 0)}
	switch typeName {
	case "string":
		return browser.stringValue(operation, client, query.Key, result)
	case "hash":
		return browser.hashValue(operation, client, query, result)
	case "list":
		return browser.listValue(operation, client, query, result)
	case "set":
		return browser.setValue(operation, client, query, result)
	case "zset":
		return browser.sortedSetValue(operation, client, query, result)
	default:
		return result, nil
	}
}

func (browser *Browser) open(connection domainredis.Connection, database int) *redisclient.Client {
	return redisclient.NewClient(&redisclient.Options{
		Addr:         net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port)),
		Password:     connection.Password,
		DB:           database,
		DialTimeout:  QUERY_TIMEOUT,
		ReadTimeout:  QUERY_TIMEOUT,
		WriteTimeout: QUERY_TIMEOUT,
	})
}

func (browser *Browser) stringValue(ctx context.Context, client *redisclient.Client, key string, result domainredis.Value) (domainredis.Value, error) {
	length, err := client.StrLen(ctx, key).Result()
	if err != nil {
		return emptyValue(), err
	}
	text, err := client.GetRange(ctx, key, 0, domainredis.MAX_STRING_BYTES-1).Result()
	if err != nil {
		return emptyValue(), err
	}
	value, encoding := displayValue([]byte(text))
	result.Length = length
	result.Text = value
	result.Encoding = encoding
	result.Truncated = length > domainredis.MAX_STRING_BYTES
	return result, nil
}

func (browser *Browser) hashValue(ctx context.Context, client *redisclient.Client, query domainredis.ValueQuery, result domainredis.Value) (domainredis.Value, error) {
	length, err := client.HLen(ctx, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	values, cursor, err := client.HScan(ctx, query.Key, query.Cursor, "*", query.Limit).Result()
	if err != nil {
		return emptyValue(), err
	}
	result.Length = length
	result.Cursor = cursor
	for index := 0; index+1 < len(values); index += 2 {
		field, _ := displayValue([]byte(values[index]))
		value, _ := displayValue([]byte(values[index+1]))
		result.Entries = append(result.Entries, domainredis.Entry{Field: field, Value: value})
	}
	return result, nil
}

func (browser *Browser) listValue(ctx context.Context, client *redisclient.Client, query domainredis.ValueQuery, result domainredis.Value) (domainredis.Value, error) {
	length, err := client.LLen(ctx, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	if query.Cursor > uint64(1<<63-1) {
		return emptyValue(), fmt.Errorf("列表游标超出范围")
	}
	start := int64(query.Cursor)
	result.Length = length
	if start >= length {
		return result, nil
	}
	values, err := client.LRange(ctx, query.Key, start, start+query.Limit-1).Result()
	if err != nil {
		return emptyValue(), err
	}
	for index, item := range values {
		value, _ := displayValue([]byte(item))
		result.Entries = append(result.Entries, domainredis.Entry{Field: strconv.FormatInt(start+int64(index), 10), Value: value})
	}
	next := start + int64(len(values))
	if next < length {
		result.Cursor = uint64(next)
	}
	return result, nil
}

func (browser *Browser) setValue(ctx context.Context, client *redisclient.Client, query domainredis.ValueQuery, result domainredis.Value) (domainredis.Value, error) {
	length, err := client.SCard(ctx, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	values, cursor, err := client.SScan(ctx, query.Key, query.Cursor, "*", query.Limit).Result()
	if err != nil {
		return emptyValue(), err
	}
	result.Length = length
	result.Cursor = cursor
	for _, item := range values {
		value, _ := displayValue([]byte(item))
		result.Entries = append(result.Entries, domainredis.Entry{Value: value})
	}
	return result, nil
}

func (browser *Browser) sortedSetValue(ctx context.Context, client *redisclient.Client, query domainredis.ValueQuery, result domainredis.Value) (domainredis.Value, error) {
	length, err := client.ZCard(ctx, query.Key).Result()
	if err != nil {
		return emptyValue(), err
	}
	values, cursor, err := client.ZScan(ctx, query.Key, query.Cursor, "*", query.Limit).Result()
	if err != nil {
		return emptyValue(), err
	}
	result.Length = length
	result.Cursor = cursor
	for index := 0; index+1 < len(values); index += 2 {
		value, _ := displayValue([]byte(values[index]))
		result.Entries = append(result.Entries, domainredis.Entry{Value: value, Score: values[index+1]})
	}
	return result, nil
}

func parseDatabases(info string) []domainredis.Database {
	result := make([]domainredis.Database, 0)
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "db") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		index, err := strconv.Atoi(strings.TrimPrefix(parts[0], "db"))
		if err != nil || index < 0 {
			continue
		}
		database := domainredis.Database{Index: index}
		for _, item := range strings.Split(parts[1], ",") {
			pair := strings.SplitN(item, "=", 2)
			if len(pair) != 2 {
				continue
			}
			value, err := strconv.ParseInt(pair[1], 10, 64)
			if err != nil {
				continue
			}
			if pair[0] == "keys" {
				database.Keys = value
			}
			if pair[0] == "expires" {
				database.Expires = value
			}
		}
		result = append(result, database)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].Index < result[right].Index })
	return result
}

func displayValue(value []byte) (string, string) {
	if utf8.Valid(value) && !strings.ContainsRune(string(value), '\x00') {
		return string(value), "text"
	}
	return base64.StdEncoding.EncodeToString(value), "base64"
}

func ttlSeconds(value time.Duration) int64 {
	if value == -1*time.Nanosecond {
		return -1
	}
	if value == -2*time.Nanosecond {
		return -2
	}
	return int64(value.Seconds())
}

func emptyValue() domainredis.Value {
	return domainredis.Value{Entries: make([]domainredis.Entry, 0)}
}
