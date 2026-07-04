package merger

import "testing"

func TestMergeStrategy_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
		want     string
	}{
		{"merge", MergeStrategyMerge, "merge"},
		{"squash", MergeStrategySquash, "squash"},
		{"rebase", MergeStrategyRebase, "rebase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.strategy)
			if got != tt.want {
				t.Errorf("MergeStrategy = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Strategy != MergeStrategySquash {
		t.Errorf("Strategy = %q, want %q", opts.Strategy, MergeStrategySquash)
	}
	if !opts.DeleteBranch {
		t.Error("DeleteBranch = false, want true")
	}
	if opts.WaitForChecks {
		t.Error("WaitForChecks = true, want false")
	}
	if opts.ChecksTimeout != 300 {
		t.Errorf("ChecksTimeout = %d, want 300", opts.ChecksTimeout)
	}
}

func TestMergeInfo(t *testing.T) {
	info := MergeInfo{
		SHA:     "abc123",
		Message: "Merged PR #123",
		Merged:  true,
	}

	if info.SHA != "abc123" {
		t.Errorf("SHA = %q, want %q", info.SHA, "abc123")
	}
	if info.Message != "Merged PR #123" {
		t.Errorf("Message = %q, want %q", info.Message, "Merged PR #123")
	}
	if !info.Merged {
		t.Error("Merged = false, want true")
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
				Strategy:      MergeStrategyRebase,
				DeleteBranch:  false,
				CommitMessage: "Custom commit message",
				WaitForChecks: true,
				ChecksTimeout: 600,
			},
		},
		{
			name: "minimal options",
			options: Options{
				Strategy: MergeStrategyMerge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify options can be created and accessed
			if tt.options.Strategy == "" && tt.name == "custom options" {
				t.Error("Strategy should not be empty for custom options")
			}
		})
	}
}
