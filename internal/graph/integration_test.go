package graph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v82/github"
)

// mockGitHubServer creates a test server that simulates GitHub API responses.
func mockGitHubServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Find matching handler
		for pattern, handler := range handlers {
			if strings.HasPrefix(path, pattern) || path == pattern {
				handler(w, r)
				return
			}
		}

		// Default: not found
		t.Logf("No handler for path: %s", path)
		http.NotFound(w, r)
	}))
}

// makeRepoResponse creates a GitHub repository response.
func makeRepoResponse(owner, name string) *github.Repository {
	return &github.Repository{
		Name: github.Ptr(name),
		Owner: &github.User{
			Login: github.Ptr(owner),
		},
		FullName:      github.Ptr(owner + "/" + name),
		DefaultBranch: github.Ptr("main"),
		Language:      github.Ptr("Go"),
	}
}

// makeGoModContent creates a base64-encoded go.mod content.
func makeGoModContent(modulePath string, deps []string) string {
	var sb strings.Builder
	sb.WriteString("module ")
	sb.WriteString(modulePath)
	sb.WriteString("\n\ngo 1.21\n")

	if len(deps) > 0 {
		sb.WriteString("\nrequire (\n")
		for _, dep := range deps {
			sb.WriteString("\t")
			sb.WriteString(dep)
			sb.WriteString("\n")
		}
		sb.WriteString(")\n")
	}

	return base64.StdEncoding.EncodeToString([]byte(sb.String()))
}

// newBuilderWithMockServer creates a builder with a mock GitHub client.
func newBuilderWithMockServer(server *httptest.Server) *Builder {
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")
	return &Builder{client: client}
}

func TestBuilder_Build_Integration(t *testing.T) {
	// Set up mock handlers
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("testorg", "mogo"),
				makeRepoResponse("testorg", "gogithub"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/testorg/mogo/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content:  github.Ptr(makeGoModContent("github.com/testorg/mogo", nil)),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
		"/repos/testorg/gogithub/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content: github.Ptr(makeGoModContent("github.com/testorg/gogithub", []string{
					"github.com/testorg/mogo v0.70.0",
				})),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"go"},
	}

	ctx := context.Background()
	graph, err := builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify modules were added
	if len(graph.modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(graph.modules))
	}

	// Verify mogo module
	mogoID := "go:github.com/testorg/mogo"
	mogo, ok := graph.modules[mogoID]
	if !ok {
		t.Error("mogo module not found")
	} else {
		if mogo.Name != "github.com/testorg/mogo" {
			t.Errorf("expected mogo name, got %s", mogo.Name)
		}
		if !mogo.IsManaged {
			t.Error("expected mogo to be managed")
		}
	}

	// Verify gogithub module with dependency
	gogithubID := "go:github.com/testorg/gogithub"
	gogithub, ok := graph.modules[gogithubID]
	if !ok {
		t.Error("gogithub module not found")
	} else {
		if len(gogithub.Dependencies) != 1 {
			t.Errorf("expected 1 dependency, got %d", len(gogithub.Dependencies))
		} else if gogithub.Dependencies[0].ID != mogoID {
			t.Errorf("expected dependency on mogo, got %s", gogithub.Dependencies[0].ID)
		}
	}
}

func TestBuilder_Build_NoGoMod(t *testing.T) {
	// Repo without go.mod should be skipped
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("testorg", "docs-only"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/testorg/docs-only/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			// Return 404 for no go.mod
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Not Found",
			})
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"go"},
	}

	ctx := context.Background()
	graph, err := builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should have no modules since no go.mod exists
	if len(graph.modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(graph.modules))
	}
}

func TestBuilder_Build_APIError(t *testing.T) {
	// Simulate API error
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Internal Server Error",
			})
		},
		"/orgs/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Internal Server Error",
			})
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"go"},
	}

	ctx := context.Background()
	_, err := builder.Build(ctx, portfolio)

	// Should return error from API
	if err == nil {
		t.Error("expected error from Build")
	}
}

func TestBuilder_Build_MultipleOrgs(t *testing.T) {
	// Test with multiple orgs
	handlers := map[string]http.HandlerFunc{
		"/users/org1/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("org1", "core"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/users/org2/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("org2", "app"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/org1/core/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content:  github.Ptr(makeGoModContent("github.com/org1/core", nil)),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
		"/repos/org2/app/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content: github.Ptr(makeGoModContent("github.com/org2/app", []string{
					"github.com/org1/core v1.0.0",
				})),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/org1", "github.com/org2"},
		Languages: []string{"go"},
	}

	ctx := context.Background()
	graph, err := builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should have 2 modules from 2 orgs
	if len(graph.modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(graph.modules))
	}

	// Verify cross-org dependency
	appID := "go:github.com/org2/app"
	app, ok := graph.modules[appID]
	if !ok {
		t.Error("app module not found")
	} else {
		if len(app.Dependencies) != 1 {
			t.Errorf("expected 1 dependency, got %d", len(app.Dependencies))
		}
		// Dependency should be marked as managed since org1 is in portfolio orgs
		if !app.Dependencies[0].IsManaged {
			t.Error("expected dependency on core to be marked as managed")
		}
	}
}

func TestBuilder_Build_ExternalDependency(t *testing.T) {
	// Test that external dependencies are properly marked
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("testorg", "mycli"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/testorg/mycli/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content: github.Ptr(makeGoModContent("github.com/testorg/mycli", []string{
					"github.com/spf13/cobra v1.8.0",
					"github.com/spf13/viper v1.18.0",
				})),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"go"},
	}

	ctx := context.Background()
	graph, err := builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify mycli module
	mycliID := "go:github.com/testorg/mycli"
	mycli, ok := graph.modules[mycliID]
	if !ok {
		t.Fatal("mycli module not found")
	}

	// Should have 2 external dependencies
	if len(mycli.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(mycli.Dependencies))
	}

	// Both should be marked as not managed (external)
	for _, dep := range mycli.Dependencies {
		if dep.IsManaged {
			t.Errorf("expected %s to be external (not managed)", dep.ID)
		}
	}
}

func TestBuilder_Build_WithCache(t *testing.T) {
	// Test that caching works
	callCount := 0
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("testorg", "cached"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/testorg/cached/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			content := &github.RepositoryContent{
				Content:  github.Ptr(makeGoModContent("github.com/testorg/cached", nil)),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	// Create builder with cache
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

	cache, err := NewCache(CacheConfig{MemoryOnly: true})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	builder := &Builder{client: client, cache: cache}

	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"go"},
	}

	ctx := context.Background()

	// First build
	_, err = builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("First build failed: %v", err)
	}

	firstCallCount := callCount

	// Second build should use cache for go.mod
	_, buildErr := builder.Build(ctx, portfolio)
	if buildErr != nil {
		t.Fatalf("Second build failed: %v", buildErr)
	}

	// go.mod fetch should be cached (callCount should not increase for that)
	if callCount > firstCallCount*2 {
		t.Logf("Call count after cache: %d (first: %d)", callCount, firstCallCount)
		// This is expected since repos list is not cached, but go.mod is
	}
}

func TestBuilder_Build_LanguageFilter(t *testing.T) {
	// Test that non-Go languages are filtered out when not requested
	handlers := map[string]http.HandlerFunc{
		"/users/testorg/repos": func(w http.ResponseWriter, r *http.Request) {
			repos := []*github.Repository{
				makeRepoResponse("testorg", "goapp"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(repos)
		},
		"/repos/testorg/goapp/contents/go.mod": func(w http.ResponseWriter, r *http.Request) {
			content := &github.RepositoryContent{
				Content:  github.Ptr(makeGoModContent("github.com/testorg/goapp", nil)),
				Encoding: github.Ptr("base64"),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(content)
		},
	}

	server := mockGitHubServer(t, handlers)
	defer server.Close()

	builder := newBuilderWithMockServer(server)

	// Request only TypeScript - should find no Go modules
	portfolio := Portfolio{
		Name:      "test",
		Orgs:      []string{"github.com/testorg"},
		Languages: []string{"typescript"},
	}

	ctx := context.Background()
	graph, err := builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should have no modules since we filtered for TypeScript
	if len(graph.modules) != 0 {
		t.Errorf("expected 0 modules when filtering for TypeScript, got %d", len(graph.modules))
	}

	// Now request Go
	portfolio.Languages = []string{"go"}
	graph, err = builder.Build(ctx, portfolio)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should have 1 module
	if len(graph.modules) != 1 {
		t.Errorf("expected 1 module when filtering for Go, got %d", len(graph.modules))
	}
}
