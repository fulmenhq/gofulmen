package appidentity

import (
	"context"
	"testing"
)

func TestNewFixture(t *testing.T) {
	t.Run("default_values", func(t *testing.T) {
		identity := NewFixture()

		if identity.BinaryName != "testapp" {
			t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "testapp")
		}
		if identity.Vendor != "testvendor" {
			t.Errorf("Vendor = %q, want %q", identity.Vendor, "testvendor")
		}
		if identity.EnvPrefix != "TESTAPP_" {
			t.Errorf("EnvPrefix = %q, want %q", identity.EnvPrefix, "TESTAPP_")
		}
		if identity.ConfigName != "testapp" {
			t.Errorf("ConfigName = %q, want %q", identity.ConfigName, "testapp")
		}
		if identity.Description != "Test application fixture" {
			t.Errorf("Description = %q, want %q", identity.Description, "Test application fixture")
		}
	})

	t.Run("with_overrides", func(t *testing.T) {
		identity := NewFixture(func(id *Identity) {
			id.BinaryName = "myapp"
			id.EnvPrefix = "MYAPP_"
		})

		if identity.BinaryName != "myapp" {
			t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "myapp")
		}
		if identity.EnvPrefix != "MYAPP_" {
			t.Errorf("EnvPrefix = %q, want %q", identity.EnvPrefix, "MYAPP_")
		}
		// Other fields should retain defaults
		if identity.Vendor != "testvendor" {
			t.Errorf("Vendor = %q, want %q", identity.Vendor, "testvendor")
		}
	})

	t.Run("with_metadata_overrides", func(t *testing.T) {
		identity := NewFixture(func(id *Identity) {
			id.Metadata.License = "Apache-2.0"
			id.Metadata.ProjectURL = "https://github.com/example/app"
		})

		if identity.Metadata.License != "Apache-2.0" {
			t.Errorf("License = %q, want %q", identity.Metadata.License, "Apache-2.0")
		}
		if identity.Metadata.ProjectURL != "https://github.com/example/app" {
			t.Errorf("ProjectURL = %q, want %q", identity.Metadata.ProjectURL, "https://github.com/example/app")
		}
	})

	t.Run("multiple_overrides", func(t *testing.T) {
		identity := NewFixture(
			func(id *Identity) {
				id.BinaryName = "app1"
			},
			func(id *Identity) {
				id.Vendor = "vendor1"
			},
			func(id *Identity) {
				id.Metadata.License = "MIT"
			},
		)

		if identity.BinaryName != "app1" {
			t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "app1")
		}
		if identity.Vendor != "vendor1" {
			t.Errorf("Vendor = %q, want %q", identity.Vendor, "vendor1")
		}
		if identity.Metadata.License != "MIT" {
			t.Errorf("License = %q, want %q", identity.Metadata.License, "MIT")
		}
	})
}

func TestNewCompleteFixture(t *testing.T) {
	t.Run("all_fields_populated", func(t *testing.T) {
		identity := NewCompleteFixture()

		// Core fields
		if identity.BinaryName != "myapp" {
			t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "myapp")
		}
		if identity.Vendor != "myvendor" {
			t.Errorf("Vendor = %q, want %q", identity.Vendor, "myvendor")
		}

		// Metadata fields
		if identity.Metadata.ProjectURL != "https://github.com/example/myapp" {
			t.Errorf("ProjectURL = %q, want %q", identity.Metadata.ProjectURL, "https://github.com/example/myapp")
		}
		if identity.Metadata.License != "MIT" {
			t.Errorf("License = %q, want %q", identity.Metadata.License, "MIT")
		}
		if identity.Metadata.RepositoryCategory != "cli" {
			t.Errorf("RepositoryCategory = %q, want %q", identity.Metadata.RepositoryCategory, "cli")
		}

		// Python metadata
		if identity.Metadata.Python == nil {
			t.Fatal("Python metadata should not be nil")
		}
		if identity.Metadata.Python.DistributionName != "my-app" {
			t.Errorf("DistributionName = %q, want %q", identity.Metadata.Python.DistributionName, "my-app")
		}
		if len(identity.Metadata.Python.ConsoleScripts) != 2 {
			t.Errorf("len(ConsoleScripts) = %d, want 2", len(identity.Metadata.Python.ConsoleScripts))
		}

		// Extras
		if identity.Metadata.Extras == nil {
			t.Fatal("Extras should not be nil")
		}
		if identity.Metadata.Extras["custom_field"] != "custom_value" {
			t.Errorf("custom_field = %v, want %q", identity.Metadata.Extras["custom_field"], "custom_value")
		}
	})

	t.Run("with_overrides", func(t *testing.T) {
		identity := NewCompleteFixture(func(id *Identity) {
			id.Metadata.License = "GPL-3.0"
			id.BinaryName = "customapp"
		})

		if identity.Metadata.License != "GPL-3.0" {
			t.Errorf("License = %q, want %q", identity.Metadata.License, "GPL-3.0")
		}
		if identity.BinaryName != "customapp" {
			t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "customapp")
		}
		// Other fields should retain defaults
		if identity.Vendor != "myvendor" {
			t.Errorf("Vendor = %q, want %q", identity.Vendor, "myvendor")
		}
	})
}

func TestFixturesCompatibleWithValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("minimal_fixture_validates", func(t *testing.T) {
		identity := NewFixture()
		if err := ValidateIdentity(ctx, identity); err != nil {
			t.Errorf("NewFixture() should produce valid identity: %v", err)
		}
	})

	t.Run("complete_fixture_validates", func(t *testing.T) {
		identity := NewCompleteFixture()
		if err := ValidateIdentity(ctx, identity); err != nil {
			t.Errorf("NewCompleteFixture() should produce valid identity: %v", err)
		}
	})
}
