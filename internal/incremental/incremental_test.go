package incremental

import (
	"testing"
)

func TestTracker_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewTracker(tmpDir)

	// Set some state
	route := "/about"
	state := &PageState{
		Route:      route,
		SourcePath: "about.md",
		OutputPath: "about/index.html",
		Hashes: []FileHash{
			{Path: "about.md", Hash: "hash1", Size: 100, Mod: 123},
		},
		BuiltAt: 456,
	}
	tracker.SetPageState(route, state)

	// Save and Load
	if err := tracker.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	tracker2 := NewTracker(tmpDir)
	if err := tracker2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	got, ok := tracker2.GetPageState(route)
	if !ok {
		t.Fatalf("PageState not found after load")
	}
	if got.SourcePath != state.SourcePath {
		t.Errorf("Expected SourcePath %s, got %s", state.SourcePath, got.SourcePath)
	}
}

func TestTracker_NeedsRebuild(t *testing.T) {
	tracker := NewTracker(t.TempDir())
	route := "/test"
	
	hashes1 := []FileHash{{Path: "f1", Hash: "h1"}}
	hashes2 := []FileHash{{Path: "f1", Hash: "h2"}}
	hashes3 := []FileHash{{Path: "f1", Hash: "h1"}, {Path: "f2", Hash: "h2"}}

	// Case 1: No state exists -> needs rebuild
	if !tracker.NeedsRebuild(route, hashes1) {
		t.Error("Expected rebuild when no state exists")
	}

	// Case 2: State exists and matches -> no rebuild
	tracker.SetPageState(route, &PageState{
		Route:  route,
		Hashes: hashes1,
	})
	if tracker.NeedsRebuild(route, hashes1) {
		t.Error("Expected no rebuild when hashes match")
	}

	// Case 3: Hash changed -> needs rebuild
	if !tracker.NeedsRebuild(route, hashes2) {
		t.Error("Expected rebuild expected when hash changes")
	}

	// Case 4: Number of files changed -> needs rebuild
	if !tracker.NeedsRebuild(route, hashes3) {
		t.Error("rebuild expected when number of files changes")
	}
}

func TestTracker_GlobalChanged(t *testing.T) {
	tracker := NewTracker(t.TempDir())
	
	hashes1 := []FileHash{{Path: "g1", Hash: "gh1"}}
	hashes2 := []FileHash{{Path: "g1", Hash: "gh2"}}

	tracker.SetGlobalHashes(hashes1)

	if tracker.GlobalChanged(hashes1) {
		t.Error("Expected no global change when hashes match")
	}

	if !tracker.GlobalChanged(hashes2) {
		t.Error("Expected global change when hash differs")
	}
}

func TestTracker_HashContent(t *testing.T) {
	tracker := NewTracker(t.TempDir())
	content := "hello world"
	h1 := tracker.HashContent(content)
	h2 := tracker.HashContent(content)
	h3 := tracker.HashContent("different")

	if h1 != h2 {
		t.Errorf("Hash should be deterministic: %s != %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("Different content should have different hashes")
	}
}
