package app

import (
	"strings"
	"testing"

	"devenv/internal/domain/software"
)

func TestRenderComposerCommandUsesDevEnvPHP(t *testing.T) {
	content := renderGlobalCommand(globalCommandBinding{
		Target:     `C:\DevEnv\software\composer\2.8.0\composer.phar`,
		PHPBinary:  `C:\DevEnv\software\php\8.3.35\php.exe`,
		Definition: software.CommandDefinition{RequiresPHP: true},
	})
	for _, expected := range []string{
		`"C:\DevEnv\software\php\8.3.35\php.exe" "C:\DevEnv\software\composer\2.8.0\composer.phar" %*`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected Composer wrapper to contain %q\n%s", expected, content)
		}
	}
}

func TestRenderComposerCommandFallsBackToPHPFromPath(t *testing.T) {
	content := renderGlobalCommand(globalCommandBinding{
		Target:     `C:\DevEnv\software\composer\2.8.0\composer.phar`,
		Definition: software.CommandDefinition{RequiresPHP: true},
	})
	if strings.Contains(content, "if exist") {
		t.Fatalf("did not expect a fixed PHP path when no PHP runtime is available\n%s", content)
	}
	if !strings.Contains(content, `php "C:\DevEnv\software\composer\2.8.0\composer.phar" %*`) {
		t.Fatalf("expected Composer wrapper to fall back to PATH PHP\n%s", content)
	}
}

func TestRenderGlobalCommandUsesShellAndJavaHome(t *testing.T) {
	content := renderGlobalCommand(globalCommandBinding{
		Name:       "mvn",
		Target:     `C:\DevEnv\software\maven\3.9.9\bin\mvn.cmd`,
		JavaHome:   `C:\DevEnv\software\java\21`,
		Definition: software.CommandDefinition{Name: "mvn", RequiresJava: true, UseCommandShell: true},
	})
	for _, expected := range []string{
		`set "JAVA_HOME=C:\DevEnv\software\java\21"`,
		`call "C:\DevEnv\software\maven\3.9.9\bin\mvn.cmd" %*`,
		`endlocal & exit /b %exitCode%`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected global command wrapper to contain %q\n%s", expected, content)
		}
	}
}

func TestGlobalCommandDefinitionsCoverCommandLineSoftware(t *testing.T) {
	for _, kind := range []string{
		software.KIND_PHP,
		software.KIND_NODE,
		software.KIND_GO,
		software.KIND_PYTHON,
		software.KIND_COMPOSER,
		software.KIND_JAVA,
		software.KIND_DOTNET,
		software.KIND_DENO,
		software.KIND_BUN,
		software.KIND_GRADLE,
		software.KIND_MAVEN,
	} {
		if !software.HasGlobalCommands(kind) {
			t.Fatalf("expected %s to expose global commands", kind)
		}
	}
}
