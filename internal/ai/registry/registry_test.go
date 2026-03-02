package registry

import (
	"testing"
)

func TestGlobalRegistry(t *testing.T) {
	// Get global registry - should be singleton
	reg1 := Global()
	reg2 := Global()

	if reg1 != reg2 {
		t.Error("Global() should return the same instance")
	}
}

func TestGetNonExistentTool(t *testing.T) {
	reg := Global()

	tool := reg.Get("non_existent_tool_xyz123")
	if tool != nil {
		t.Error("Get should return nil for non-existent tool")
	}
}

func TestListMetadata(t *testing.T) {
	reg := Global()

	// List should work without error
	list := reg.ListMetadata()
	if list == nil {
		t.Error("ListMetadata should never return nil")
	}
}

func TestEnableDisableNonExistent(t *testing.T) {
	reg := Global()

	// Try to disable non-existent tool
	err := reg.Disable("non_existent_tool_xyz")
	if err == nil {
		t.Error("Expected error when disabling non-existent tool")
	}

	// Try to enable non-existent tool
	err = reg.Enable("non_existent_tool_xyz")
	if err == nil {
		t.Error("Expected error when enabling non-existent tool")
	}
}

func TestGetMetadataNonExistent(t *testing.T) {
	reg := Global()

	_, ok := reg.GetMetadata("non_existent_tool_xyz")
	if ok {
		t.Error("GetMetadata should return false for non-existent tool")
	}
}

func TestCount(t *testing.T) {
	reg := Global()

	// Just test that Count doesn't panic
	count := reg.Count()
	if count < 0 {
		t.Errorf("Count should be non-negative, got %d", count)
	}

	enabledCount := reg.CountEnabled()
	if enabledCount < 0 || enabledCount > count {
		t.Errorf("CountEnabled should be between 0 and Count, got %d (total=%d)", enabledCount, count)
	}
}

func TestGetAll(t *testing.T) {
	reg := Global()

	// GetAll should work without error for various agent types
	allTools := reg.GetAll("")
	// Note: nil is acceptable - we just check it doesn't panic
	if allTools != nil && len(allTools) > 0 {
		t.Logf("Found %d tools", len(allTools))
	}

	chatTools := reg.GetAll("chat")
	if chatTools != nil && len(chatTools) > 0 {
		t.Logf("Found %d chat tools", len(chatTools))
	}

	planTools := reg.GetAll("plan_execute")
	if planTools != nil && len(planTools) > 0 {
		t.Logf("Found %d plan_execute tools", len(planTools))
	}
}

func TestFilterByCategory(t *testing.T) {
	reg := Global()

	// Filter by category - should work even if no tools match
	filtered := reg.FilterByCategory("non_existent_category_xyz")
	if filtered != nil && len(filtered) == 0 {
		// Expected - empty category returns empty or nil
	}
}

func TestGetByPrefix(t *testing.T) {
	reg := Global()

	// GetByPrefix should work even if no tools match
	filtered := reg.GetByPrefix("non_existent_prefix_xyz_")
	if filtered != nil && len(filtered) == 0 {
		// Expected - no tools with this prefix
	}
}

func TestUnregisterNonExistent(t *testing.T) {
	reg := Global()

	// Unregister should not panic for non-existent tool
	reg.Unregister("non_existent_tool_xyz")
	// If we get here, test passes
}
