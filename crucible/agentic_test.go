package crucible

import (
	"slices"
	"sort"
	"testing"
)

// coreSlugs are roles that must always be present in the catalog.
// This list is intentionally conservative — new roles may be added upstream
// without requiring test updates.
var coreSlugs = []string{
	"cicd", "dataeng", "devlead", "devrev", "entarch",
	"infoarch", "prodmktg", "qa", "releng", "secrev", "uxdev",
}

func TestLoadRole(t *testing.T) {
	role, err := LoadRole("devlead")
	if err != nil {
		t.Fatalf("LoadRole(\"devlead\") failed: %v", err)
	}
	if role.Slug != "devlead" {
		t.Errorf("Slug = %q, want %q", role.Slug, "devlead")
	}
	if role.Name == "" {
		t.Error("expected non-empty Name")
	}
	if role.Description == "" {
		t.Error("expected non-empty Description")
	}
	if role.Status == "" {
		t.Error("expected non-empty Status")
	}
	if len(role.Scope) == 0 {
		t.Error("expected at least one Scope item")
	}
	if len(role.Responsibilities) == 0 {
		t.Error("expected at least one Responsibility")
	}
	if len(role.DoesNot) == 0 {
		t.Error("expected at least one DoesNot item")
	}
	if len(role.EscalatesTo) == 0 {
		t.Error("expected at least one EscalatesTo entry")
	}
}

func TestLoadRole_Mindset(t *testing.T) {
	role, err := LoadRole("devlead")
	if err != nil {
		t.Fatalf("LoadRole(\"devlead\") failed: %v", err)
	}
	if role.Mindset == nil {
		t.Fatal("expected Mindset to be populated for devlead")
	}
	if len(role.Mindset.Focus) == 0 {
		t.Error("expected at least one Mindset.Focus item")
	}
	if len(role.Mindset.Principles) == 0 {
		t.Error("expected at least one Mindset.Principles item")
	}
}

func TestLoadRole_Escalations(t *testing.T) {
	role, err := LoadRole("devlead")
	if err != nil {
		t.Fatalf("LoadRole(\"devlead\") failed: %v", err)
	}
	for i, e := range role.EscalatesTo {
		if e.Target == "" {
			t.Errorf("EscalatesTo[%d].Target is empty", i)
		}
		if e.When == "" {
			t.Errorf("EscalatesTo[%d].When is empty", i)
		}
	}
}

func TestLoadRole_PrePushChecklist(t *testing.T) {
	role, err := LoadRole("releng")
	if err != nil {
		t.Fatalf("LoadRole(\"releng\") failed: %v", err)
	}
	if len(role.PrePushChecklist) == 0 {
		t.Error("expected non-empty PrePushChecklist for releng")
	}
}

func TestLoadRole_RequiredReading(t *testing.T) {
	role, err := LoadRole("releng")
	if err != nil {
		t.Fatalf("LoadRole(\"releng\") failed: %v", err)
	}
	if role.RequiredReading == nil {
		t.Fatal("expected RequiredReading to be populated for releng")
	}
	if role.RequiredReading.Description == "" {
		t.Error("expected non-empty RequiredReading.Description")
	}
	// Verify Files slice is accessible (full RoleRequiredReading shape).
	if len(role.RequiredReading.Files) == 0 {
		t.Error("expected non-empty RequiredReading.Files for releng")
	}
	for i, f := range role.RequiredReading.Files {
		if f.Path == "" {
			t.Errorf("RequiredReading.Files[%d].Path is empty", i)
		}
		if f.Reason == "" {
			t.Errorf("RequiredReading.Files[%d].Reason is empty", i)
		}
	}
}

func TestLoadRole_CrossRoleNote(t *testing.T) {
	role, err := LoadRole("releng")
	if err != nil {
		t.Fatalf("LoadRole(\"releng\") failed: %v", err)
	}
	if role.CrossRoleNote == "" {
		t.Error("expected non-empty CrossRoleNote for releng")
	}
}

func TestLoadRole_NotFound(t *testing.T) {
	_, err := LoadRole("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent role slug")
	}
}

func TestLoadRole_InvalidSlug(t *testing.T) {
	invalidSlugs := []string{
		"dev-lead", // hyphens not allowed
		"dev_lead", // underscores not allowed
		"1devlead", // must start with letter
		"",         // empty
		"DevLead",  // uppercase not allowed
	}
	for _, slug := range invalidSlugs {
		_, err := LoadRole(slug)
		if err == nil {
			t.Errorf("expected error for invalid slug %q", slug)
		}
	}
}

func TestListRoleSlugs(t *testing.T) {
	slugs, err := ListRoleSlugs()
	if err != nil {
		t.Fatalf("ListRoleSlugs() failed: %v", err)
	}

	// Verify core slugs are present.
	for _, expected := range coreSlugs {
		if !slices.Contains(slugs, expected) {
			t.Errorf("expected core slug %q not found in catalog", expected)
		}
	}

	// README must be excluded.
	if slices.Contains(slugs, "README") {
		t.Error("README should not appear as a slug")
	}

	// Result must be sorted.
	if !sort.StringsAreSorted(slugs) {
		t.Error("slugs should be sorted")
	}
}

func TestListRoleSlugs_MinCount(t *testing.T) {
	slugs, err := ListRoleSlugs()
	if err != nil {
		t.Fatalf("ListRoleSlugs() failed: %v", err)
	}
	// At least 11 roles (the known floor from v0.4.10). The catalog grows
	// as roles are added upstream — do not hard-code an exact count.
	if len(slugs) < len(coreSlugs) {
		t.Errorf("expected at least %d roles, got %d", len(coreSlugs), len(slugs))
	}
	t.Logf("catalog contains %d roles", len(slugs))
}

func TestLoadRoleCatalog(t *testing.T) {
	catalog, err := LoadRoleCatalog()
	if err != nil {
		t.Fatalf("LoadRoleCatalog() failed: %v", err)
	}
	if len(catalog) < len(coreSlugs) {
		t.Errorf("expected at least %d roles, got %d", len(coreSlugs), len(catalog))
	}

	// Verify well-known roles are keyed by slug.
	for _, slug := range []string{"devlead", "devrev", "secrev"} {
		role, ok := catalog[slug]
		if !ok {
			t.Errorf("catalog missing expected role %q", slug)
			continue
		}
		if role.Slug != slug {
			t.Errorf("catalog[%q].Slug = %q, want %q", slug, role.Slug, slug)
		}
	}
}

func TestConfigRegistryAgentic(t *testing.T) {
	t.Run("Role returns raw YAML", func(t *testing.T) {
		data, err := ConfigRegistry.Agentic().Role("devlead")
		if err != nil {
			t.Fatalf("ConfigRegistry.Agentic().Role(\"devlead\") failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty raw YAML")
		}
	})

	t.Run("Role nonexistent returns error", func(t *testing.T) {
		_, err := ConfigRegistry.Agentic().Role("doesnotexist")
		if err == nil {
			t.Error("expected error for nonexistent role")
		}
	})
}
