package incremental

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type FileHash struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Mod  int64  `json:"mod"`
}

type PageState struct {
	Route      string     `json:"route"`
	SourcePath string     `json:"source_path"`
	OutputPath string     `json:"output_path"`
	Hashes     []FileHash `json:"hashes"`
	BuiltAt    int64      `json:"built_at"`
}

type BuildCache struct {
	Version   int                   `json:"version"`
	Pages     map[string]*PageState `json:"pages"`
	Global    []FileHash            `json:"global"`
	CreatedAt int64                 `json:"created_at"`
}

type Tracker struct {
	cache    *BuildCache
	cacheDir string
	mu       sync.RWMutex
}

const CacheVersion = 1

func NewTracker(cacheDir string) *Tracker {
	return &Tracker{
		cacheDir: cacheDir,
		cache: &BuildCache{
			Version:   CacheVersion,
			Pages:     make(map[string]*PageState),
			CreatedAt: time.Now().Unix(),
		},
	}
}

func (t *Tracker) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cachePath := t.cachePath()
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var cache BuildCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}

	if cache.Version != CacheVersion {
		t.cache = &BuildCache{
			Version:   CacheVersion,
			Pages:     make(map[string]*PageState),
			CreatedAt: time.Now().Unix(),
		}
		return nil
	}

	if cache.Pages == nil {
		cache.Pages = make(map[string]*PageState)
	}

	t.cache = &cache
	return nil
}

func (t *Tracker) Save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := os.MkdirAll(t.cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t.cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.cachePath(), data, 0644)
}

func (t *Tracker) cachePath() string {
	return filepath.Join(t.cacheDir, "build-cache.json")
}

func (t *Tracker) HashFile(path string) (FileHash, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileHash{}, err
	}

	sum := sha256.Sum256(data)
	return FileHash{
		Path: path,
		Hash: hex.EncodeToString(sum[:]),
		Size: int64(len(data)),
		Mod:  time.Now().Unix(),
	}, nil
}

func (t *Tracker) HashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func (t *Tracker) GetPageState(route string) (*PageState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	state, ok := t.cache.Pages[route]
	return state, ok
}

func (t *Tracker) SetPageState(route string, state *PageState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cache.Pages[route] = state
}

func (t *Tracker) NeedsRebuild(route string, currentHashes []FileHash) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, ok := t.cache.Pages[route]
	if !ok {
		return true
	}

	if len(state.Hashes) != len(currentHashes) {
		return true
	}

	for i, h := range currentHashes {
		if i >= len(state.Hashes) || state.Hashes[i].Hash != h.Hash {
			return true
		}
	}

	return false
}

func (t *Tracker) SetGlobalHashes(hashes []FileHash) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cache.Global = hashes
}

func (t *Tracker) GetGlobalHashes() []FileHash {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cache.Global
}

func (t *Tracker) GlobalChanged(currentHashes []FileHash) bool {
	stored := t.GetGlobalHashes()
	if len(stored) != len(currentHashes) {
		return true
	}

	for i, h := range currentHashes {
		if i >= len(stored) || stored[i].Hash != h.Hash {
			return true
		}
	}

	return false
}

func (t *Tracker) InvalidatePage(route string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.cache.Pages, route)
}

func (t *Tracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cache = &BuildCache{
		Version:   CacheVersion,
		Pages:     make(map[string]*PageState),
		CreatedAt: time.Now().Unix(),
	}
}

func (t *Tracker) Stats() (pages int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.cache.Pages)
}

type DependencyGraph struct {
	pages       map[string][]string
	pageToFiles map[string]map[string]bool
	files       map[string][]string
	mu          sync.RWMutex
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		pages:       make(map[string][]string),
		pageToFiles: make(map[string]map[string]bool),
		files:       make(map[string][]string),
	}
}

func (g *DependencyGraph) AddDependency(page, file string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.pages[page] == nil {
		g.pages[page] = []string{}
	}
	if g.files[file] == nil {
		g.files[file] = []string{}
	}

	g.pages[page] = append(g.pages[page], file)
	g.files[file] = append(g.files[file], page)

	if g.pageToFiles[page] == nil {
		g.pageToFiles[page] = make(map[string]bool)
	}
	g.pageToFiles[page][file] = true
}

func (g *DependencyGraph) PagesAffectedBy(changedFile string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pages, ok := g.files[changedFile]
	if !ok {
		return nil
	}

	result := make([]string, len(pages))
	copy(result, pages)
	sort.Strings(result)
	return result
}

func (g *DependencyGraph) AllPages() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pages := make([]string, 0, len(g.pages))
	for p := range g.pages {
		pages = append(pages, p)
	}
	sort.Strings(pages)
	return pages
}

func (g *DependencyGraph) FilesForPage(page string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	files, ok := g.pageToFiles[page]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(files))
	for f := range files {
		result = append(result, f)
	}
	sort.Strings(result)
	return result
}

func (g *DependencyGraph) BuildFromLayout(components map[string]string, layouts map[string][]string) {
	for page, layout := range layouts {
		g.AddDependency(page, layout)
	}
	for page, comp := range components {
		g.AddDependency(page, comp)
	}
}

func PrintSummary(built, skipped int, duration time.Duration) {
	fmt.Printf("Built: %d pages\n", built)
	if skipped > 0 {
		fmt.Printf("Skipped (cached): %d pages\n", skipped)
	}
	fmt.Printf("Total time: %s\n", duration.Round(time.Millisecond))
	if built+skipped > 0 && skipped > 0 {
		saved := time.Duration(skipped) * duration / time.Duration(built+skipped)
		fmt.Printf("Approx. time saved: %s\n", saved)
	}
}
