package redis

const (
	MAX_KEY_PAGE_SIZE   = 100
	MAX_VALUE_PAGE_SIZE = 100
	MAX_STRING_BYTES    = 65536
)

type Connection struct {
	Host     string
	Port     int
	Password string
}

type Database struct {
	Index   int
	Keys    int64
	Expires int64
}

type KeyQuery struct {
	Database int
	Pattern  string
	Cursor   uint64
	Limit    int64
}

type Key struct {
	Name string
	Type string
	TTL  int64
}

type KeyPage struct {
	Cursor uint64
	Keys   []Key
}

type ValueQuery struct {
	Database int
	Key      string
	Cursor   uint64
	Limit    int64
}

type Entry struct {
	Field string
	Value string
	Score string
}

type Value struct {
	Type      string
	TTL       int64
	Length    int64
	Text      string
	Encoding  string
	Truncated bool
	Cursor    uint64
	Entries   []Entry
}
