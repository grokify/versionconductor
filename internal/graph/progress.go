package graph

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Progress tracks and reports build progress.
type Progress struct {
	mu           sync.Mutex
	writer       io.Writer
	enabled      bool
	startTime    time.Time
	totalOrgs    int
	currentOrg   int
	totalRepos   int
	currentRepo  int
	reposFound   int
	modulesFound int
	errors       int
	lastUpdate   time.Time
	minInterval  time.Duration
}

// ProgressConfig configures progress reporting.
type ProgressConfig struct {
	// Writer is where progress is written. Default is os.Stderr.
	Writer io.Writer

	// Enabled controls whether progress is reported.
	Enabled bool

	// MinInterval is the minimum time between progress updates.
	// Default is 100ms.
	MinInterval time.Duration
}

// NewProgress creates a new progress reporter.
func NewProgress(cfg ProgressConfig) *Progress {
	if cfg.Writer == nil {
		cfg.Writer = os.Stderr
	}
	if cfg.MinInterval == 0 {
		cfg.MinInterval = 100 * time.Millisecond
	}
	return &Progress{
		writer:      cfg.Writer,
		enabled:     cfg.Enabled,
		minInterval: cfg.MinInterval,
	}
}

// Start begins tracking progress for a build operation.
func (p *Progress) Start(totalOrgs int) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.startTime = time.Now()
	p.totalOrgs = totalOrgs
	p.currentOrg = 0
	p.totalRepos = 0
	p.currentRepo = 0
	p.reposFound = 0
	p.modulesFound = 0
	p.errors = 0
	p.lastUpdate = time.Time{}

	fmt.Fprintf(p.writer, "Building dependency graph for %d organization(s)...\n", totalOrgs)
}

// StartOrg begins processing an organization.
func (p *Progress) StartOrg(org string, repoCount int) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentOrg++
	p.totalRepos = repoCount
	p.currentRepo = 0

	fmt.Fprintf(p.writer, "[%d/%d] Scanning %s (%d repos)...\n",
		p.currentOrg, p.totalOrgs, org, repoCount)
}

// ProcessRepo reports progress on processing a repository.
func (p *Progress) ProcessRepo(repo string) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentRepo++
	p.reposFound++

	// Rate limit updates
	if time.Since(p.lastUpdate) < p.minInterval {
		return
	}
	p.lastUpdate = time.Now()

	fmt.Fprintf(p.writer, "  [%d/%d] %s\r",
		p.currentRepo, p.totalRepos, truncateString(repo, 50))
}

// FoundModule reports that a Go module was found.
func (p *Progress) FoundModule(module string) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.modulesFound++
}

// Error reports an error during processing.
func (p *Progress) Error(repo string, err error) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.errors++
	// Clear line and print error
	fmt.Fprintf(p.writer, "\r  Error: %s: %v\n", repo, err)
}

// Complete finishes progress tracking and prints summary.
func (p *Progress) Complete() {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime).Round(time.Millisecond)

	// Clear line
	fmt.Fprintf(p.writer, "\r%s\r", "                                                            ")

	fmt.Fprintf(p.writer, "\nGraph build complete:\n")
	fmt.Fprintf(p.writer, "  Organizations: %d\n", p.totalOrgs)
	fmt.Fprintf(p.writer, "  Repositories:  %d\n", p.reposFound)
	fmt.Fprintf(p.writer, "  Modules:       %d\n", p.modulesFound)
	if p.errors > 0 {
		fmt.Fprintf(p.writer, "  Errors:        %d\n", p.errors)
	}
	fmt.Fprintf(p.writer, "  Duration:      %s\n", elapsed)
}

// truncateString shortens a string to maxLen.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ProgressCallback is a function called during graph building for progress updates.
type ProgressCallback func(event ProgressEvent)

// ProgressEvent represents a progress update during graph building.
type ProgressEvent struct {
	Type    ProgressEventType `json:"type"`
	Org     string            `json:"org,omitempty"`
	Repo    string            `json:"repo,omitempty"`
	Module  string            `json:"module,omitempty"`
	Current int               `json:"current,omitempty"`
	Total   int               `json:"total,omitempty"`
	Error   error             `json:"error,omitempty"`
}

// ProgressEventType indicates the type of progress event.
type ProgressEventType string

const (
	ProgressEventStart    ProgressEventType = "start"
	ProgressEventOrg      ProgressEventType = "org"
	ProgressEventRepo     ProgressEventType = "repo"
	ProgressEventModule   ProgressEventType = "module"
	ProgressEventError    ProgressEventType = "error"
	ProgressEventComplete ProgressEventType = "complete"
)

// CallbackProgress adapts a callback function to the Progress interface.
type CallbackProgress struct {
	callback    ProgressCallback
	totalOrgs   int
	currentOrg  int
	totalRepos  int
	currentRepo int
}

// NewCallbackProgress creates a progress reporter that calls a callback.
func NewCallbackProgress(callback ProgressCallback) *CallbackProgress {
	return &CallbackProgress{callback: callback}
}

// Start begins tracking progress.
func (cp *CallbackProgress) Start(totalOrgs int) {
	cp.totalOrgs = totalOrgs
	cp.callback(ProgressEvent{
		Type:  ProgressEventStart,
		Total: totalOrgs,
	})
}

// StartOrg begins processing an organization.
func (cp *CallbackProgress) StartOrg(org string, repoCount int) {
	cp.currentOrg++
	cp.totalRepos = repoCount
	cp.currentRepo = 0
	cp.callback(ProgressEvent{
		Type:    ProgressEventOrg,
		Org:     org,
		Current: cp.currentOrg,
		Total:   cp.totalOrgs,
	})
}

// ProcessRepo reports progress on a repository.
func (cp *CallbackProgress) ProcessRepo(repo string) {
	cp.currentRepo++
	cp.callback(ProgressEvent{
		Type:    ProgressEventRepo,
		Repo:    repo,
		Current: cp.currentRepo,
		Total:   cp.totalRepos,
	})
}

// FoundModule reports a found module.
func (cp *CallbackProgress) FoundModule(module string) {
	cp.callback(ProgressEvent{
		Type:   ProgressEventModule,
		Module: module,
	})
}

// Error reports an error.
func (cp *CallbackProgress) Error(repo string, err error) {
	cp.callback(ProgressEvent{
		Type:  ProgressEventError,
		Repo:  repo,
		Error: err,
	})
}

// Complete finishes progress tracking.
func (cp *CallbackProgress) Complete() {
	cp.callback(ProgressEvent{
		Type: ProgressEventComplete,
	})
}
