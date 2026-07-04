package releaser

import "testing"

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Prefix != "v" {
		t.Errorf("Prefix = %q, want %q", opts.Prefix, "v")
	}
	if !opts.GenerateNotes {
		t.Error("GenerateNotes = false, want true")
	}
	if opts.Draft {
		t.Error("Draft = true, want false")
	}
	if opts.Prerelease {
		t.Error("Prerelease = true, want false")
	}
	if !opts.IncludeBody {
		t.Error("IncludeBody = false, want true")
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name    string
		options Options
	}{
		{
			name: "custom options",
			options: Options{
				Prefix:        "",
				GenerateNotes: false,
				Draft:         true,
				Prerelease:    true,
				IncludeBody:   false,
			},
		},
		{
			name: "minimal options",
			options: Options{
				Prefix: "v",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify options can be created and accessed
			_ = tt.options
		})
	}
}
