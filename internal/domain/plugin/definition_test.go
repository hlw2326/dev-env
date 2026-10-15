package plugin

import "testing"

func TestDefinitionsUseSharedNodePlugin(t *testing.T) {
	definition, ok := DefinitionFor(ID_NODE_PACKAGE)
	if !ok {
		t.Fatal("Node package plugin should be registered")
	}
	found := false
	for _, kind := range definition.ProjectKinds {
		if kind == "node" {
			found = true
		}
	}
	if !found || definition.Handler != HANDLER_NODE {
		t.Fatalf("unexpected Node plugin definition: %#v", definition)
	}
}
