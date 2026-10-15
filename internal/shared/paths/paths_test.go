package paths

import "testing"

func TestNginxSiteConfigName(t *testing.T) {
	tests := []struct {
		id   int64
		name string
		want string
	}{
		{id: 7, name: "hlw-admin", want: "website-7-hlw-admin.conf"},
		{id: 8, name: "中文网站", want: "website-8-中文网站.conf"},
		{id: 9, name: "", want: "website-9.conf"},
	}
	for _, item := range tests {
		got := NginxSiteConfigName(item.id, item.name)
		if got != item.want {
			t.Fatalf("NginxSiteConfigName(%q) = %q, want %q", item.name, got, item.want)
		}
	}
}
