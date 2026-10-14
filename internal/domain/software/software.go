package software

const (
	KindPHP   = "php"
	KindNode  = "nodejs"
	KindMySQL = "mysql"
	KindRedis = "redis"
	KindNginx = "nginx"
)

const (
	StateInstalling = "installing"
	StateInstalled  = "installed"
	StateFailed     = "failed"
)

type Software struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	State     string `json:"state"`
	Progress  int    `json:"progress"`
	Error     string `json:"error"`
	Main      bool   `json:"main"`
	Port      int    `json:"port"`
	Autostart bool   `json:"autostart"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type Item struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Count       int    `json:"count"`
	Recommended string `json:"recommended"`
}

type Version struct {
	Version     string `json:"version"`
	Recommended bool   `json:"recommended"`
	Installed   bool   `json:"installed"`
	Note        string `json:"note"`
	URL         string `json:"-"`
	Checksum    string `json:"-"`
}

type ConfigItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type Config struct {
	Items []ConfigItem `json:"items"`
	File  string       `json:"file"`
}

type Extension struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func IsDaemon(kind string) bool {
	return kind == KindMySQL || kind == KindRedis || kind == KindNginx
}

func IsValidKind(kind string) bool {
	return kind == KindPHP || kind == KindNode || kind == KindMySQL || kind == KindRedis || kind == KindNginx
}
