package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/config"
)

func TestReloadConfigPicksUpResponsiveChanges(t *testing.T) {
	tmpdir := t.TempDir()

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`name: Test
responsive:
  mobile: "768px"
  tablet: "1024px"
  desktop: "1200px"
`), 0644)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, err := config.Load("site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s := New(cfg, 8080, nil)

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`name: Test
responsive:
  mobile: "600px"
  tablet: "900px"
  desktop: "1100px"
`), 0644)

	s.reloadConfig()

	if s.Config.Responsive.Mobile != "600px" {
		t.Errorf("expected mobile 600px after reload, got %q", s.Config.Responsive.Mobile)
	}
	if s.Config.Responsive.Tablet != "900px" {
		t.Errorf("expected tablet 900px after reload, got %q", s.Config.Responsive.Tablet)
	}
	if s.Config.Responsive.Desktop != "1100px" {
		t.Errorf("expected desktop 1100px after reload, got %q", s.Config.Responsive.Desktop)
	}
}

func TestReloadConfigKeepsPreviousOnError(t *testing.T) {
	tmpdir := t.TempDir()

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`name: Test
responsive:
  mobile: "768px"
`), 0644)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(workingDir) })

	cfg, err := config.Load("site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s := New(cfg, 8080, nil)

	os.WriteFile(filepath.Join(tmpdir, "site.yaml"), []byte(`: broken yaml [`), 0644)

	s.reloadConfig()

	if s.Config.Responsive.Mobile != "768px" {
		t.Errorf("expected previous config kept after failed reload, got %q", s.Config.Responsive.Mobile)
	}
}
