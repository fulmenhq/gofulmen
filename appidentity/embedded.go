package appidentity

import (
	"context"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	embeddedMu       sync.Mutex
	embeddedIdentity *Identity
	embeddedSet      bool
)

// RegisterEmbeddedIdentityYAML registers an embedded (compiled-in) application
// identity YAML payload.
//
// This enables distributed artifacts (installed binaries, packaged CLIs, etc.)
// to resolve identity without depending on a repository checkout or external
// `.fulmen/app.yaml` at runtime.
//
// Registration semantics:
//   - First registration wins. Subsequent calls return an error.
//   - The payload is parsed and schema-validated at registration time.
//
// Discovery precedence:
//  1. Explicit path (Options.ExplicitPath)
//  2. Environment variable override (FULMEN_APP_IDENTITY_PATH)
//  3. Embedded identity (registered via this function)
//  4. Filesystem discovery (CWD ancestor search, optional exe-dir fallback)
func RegisterEmbeddedIdentityYAML(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("embedded identity payload is empty")
	}

	var file identityFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return &MalformedError{Path: "<embedded>", Err: err}
	}

	file.App.Metadata = file.Metadata

	identity := &file.App
	if err := ValidateIdentity(context.Background(), identity); err != nil {
		return err
	}

	embeddedMu.Lock()
	defer embeddedMu.Unlock()

	if embeddedSet {
		return fmt.Errorf("embedded identity already registered")
	}

	embeddedIdentity = identity
	embeddedSet = true
	return nil
}

func getEmbeddedIdentity() (*Identity, bool) {
	embeddedMu.Lock()
	defer embeddedMu.Unlock()

	if !embeddedSet || embeddedIdentity == nil {
		return nil, false
	}
	return embeddedIdentity, true
}

func resetEmbeddedIdentity() {
	embeddedMu.Lock()
	embeddedIdentity = nil
	embeddedSet = false
	embeddedMu.Unlock()
}
