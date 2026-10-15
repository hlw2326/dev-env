package deployment

import (
	"strings"
	"testing"

	"devenv/internal/domain/project"
)

func TestDeploymentCapabilitiesAndValidation(t *testing.T) {
	if ProviderFor("php") != PROVIDER_SFTP || ProviderFor("node") != PROVIDER_SFTP || ProviderFor("miniprogram") != PROVIDER_WECHAT || ProviderFor("electron") != "" {
		t.Fatal("unexpected provider capabilities")
	}
	config := Default(1, "php")
	config.Host = "server.invalid"
	config.Username = "deploy"
	config.Fingerprint = "SHA256:" + strings.Repeat("A", 43)
	config.Remote = "/var/www/project"
	if err := Validate(config, "php"); err != nil {
		t.Fatal(err)
	}
	for _, remote := range []string{"/", "/var", "/var/www/../other", "relative", "/var/www/", "/var/www\nother"} {
		invalid := config
		invalid.Remote = remote
		if err := Validate(invalid, "php"); err == nil {
			t.Errorf("accepted unsafe target %q", remote)
		}
	}
	if err := Validate(config, "miniprogram"); err == nil {
		t.Fatal("wrong provider accepted")
	}
	if got := Default(2, project.KIND_MINIPROGRAM).Source; got != DEFAULT_WECHAT_SOURCE {
		t.Fatalf("unexpected mini program source: %s", got)
	}
	config = Default(3, project.KIND_MINIPROGRAM)
	config.AppID = "wx1234567890abcdef"
	config.AppIDs = []string{"wx1234567890abcdef", "wxabcdef1234567890", "wxabcdef1234567890"}
	config.DeveloperToolsCLI = "C:\\wechat\\cli.bat"
	config.Version = "1.0.0"
	NormalizeAppIDs(&config)
	if len(config.AppIDs) != 2 || config.AppID != config.AppIDs[0] {
		t.Fatalf("unexpected normalized AppIDs: %+v", config.AppIDs)
	}
	config.AppIDs = []string{"wxabcdef1234567890"}
	config.AppID = "wx1234567890abcdef"
	if values := config.AppIDValues(); len(values) != 1 || values[0] != "wxabcdef1234567890" {
		t.Fatalf("legacy AppID was not ignored after editing tags: %+v", values)
	}
	if err := Validate(config, project.KIND_MINIPROGRAM); err != nil {
		t.Fatal(err)
	}
}
