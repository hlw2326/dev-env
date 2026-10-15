package deployment

import (
	"strings"
	"testing"
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
}
