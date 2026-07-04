package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/plexusone/versionconductor/internal/collector"
	"github.com/plexusone/versionconductor/internal/policy"
	"github.com/plexusone/versionconductor/pkg/model"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage and evaluate Cedar policies",
	Long: `Policy commands for managing and evaluating Cedar policies.

Cedar policies are used to make decisions about whether dependency PRs
should be auto-merged, queued, or require manual review.`,
}

var policyEvaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate Cedar policies against a PR",
	Long: `Evaluate Cedar policies against a pull request context.

The context can be provided in several ways:
  1. From stdin as JSON (--stdin)
  2. By fetching PR data from GitHub (--repo and --pr)
  3. From a JSON file (--context-file)

Examples:
  # Evaluate against a GitHub PR
  versionconductor policy evaluate --repo owner/repo --pr 123

  # Evaluate with a specific profile
  versionconductor policy evaluate --repo owner/repo --pr 123 --profile quarantine

  # Evaluate with custom policies
  versionconductor policy evaluate --repo owner/repo --pr 123 --policies ./policies

  # Evaluate context from stdin
  echo '{"pr":{"number":123},...}' | versionconductor policy evaluate --stdin

  # Output as JSON for scripts
  versionconductor policy evaluate --repo owner/repo --pr 123 --format json`,
	RunE: runPolicyEvaluate,
}

func init() {
	rootCmd.AddCommand(policyCmd)
	policyCmd.AddCommand(policyEvaluateCmd)

	// Evaluate command flags
	policyEvaluateCmd.Flags().String("repo", "", "Repository (owner/repo format)")
	policyEvaluateCmd.Flags().Int("pr", 0, "Pull request number")
	policyEvaluateCmd.Flags().String("policies", "", "Path to Cedar policies directory or file")
	policyEvaluateCmd.Flags().String("profile", "balanced", "Merge profile: aggressive, balanced, conservative, quarantine")
	policyEvaluateCmd.Flags().Bool("stdin", false, "Read policy context from stdin as JSON")
	policyEvaluateCmd.Flags().String("context-file", "", "Read policy context from JSON file")
	policyEvaluateCmd.Flags().String("action", "merge", "Action to evaluate: merge, review, release")
	policyEvaluateCmd.Flags().Bool("comment", false, "Post decision as a comment on the PR")
	policyEvaluateCmd.Flags().Bool("update-comment", true, "Update existing comment instead of creating new (default true)")

	_ = viper.BindPFlag("policy.repo", policyEvaluateCmd.Flags().Lookup("repo"))
	_ = viper.BindPFlag("policy.pr", policyEvaluateCmd.Flags().Lookup("pr"))
	_ = viper.BindPFlag("policy.policies", policyEvaluateCmd.Flags().Lookup("policies"))
	_ = viper.BindPFlag("policy.profile", policyEvaluateCmd.Flags().Lookup("profile"))
	_ = viper.BindPFlag("policy.stdin", policyEvaluateCmd.Flags().Lookup("stdin"))
	_ = viper.BindPFlag("policy.context-file", policyEvaluateCmd.Flags().Lookup("context-file"))
	_ = viper.BindPFlag("policy.action", policyEvaluateCmd.Flags().Lookup("action"))
	_ = viper.BindPFlag("policy.comment", policyEvaluateCmd.Flags().Lookup("comment"))
	_ = viper.BindPFlag("policy.update-comment", policyEvaluateCmd.Flags().Lookup("update-comment"))
}

func runPolicyEvaluate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	verbose := viper.GetBool("verbose")

	// Get policy context
	pctx, err := getPolicyContext(ctx, verbose)
	if err != nil {
		return err
	}

	// Create policy engine
	policiesPath := viper.GetString("policy.policies")
	profileName := viper.GetString("policy.profile")

	engine, err := policy.NewCedarEngine(policiesPath, profileName)
	if err != nil {
		return fmt.Errorf("creating policy engine: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Loaded %d policies using profile '%s'\n", engine.PolicyCount(), profileName)
	}

	// Parse action
	actionStr := viper.GetString("policy.action")
	var action model.PolicyAction
	switch actionStr {
	case "merge":
		action = model.PolicyActionMerge
	case "review":
		action = model.PolicyActionReview
	case "release":
		action = model.PolicyActionRelease
	default:
		return fmt.Errorf("invalid action '%s': must be merge, review, or release", actionStr)
	}

	// Evaluate policy
	decision, err := engine.Evaluate(ctx, action, pctx)
	if err != nil {
		return fmt.Errorf("evaluating policy: %w", err)
	}

	// Post comment if requested
	if viper.GetBool("policy.comment") {
		repoStr := viper.GetString("policy.repo")
		prNumber := viper.GetInt("policy.pr")

		if repoStr == "" || prNumber == 0 {
			return fmt.Errorf("--comment requires --repo and --pr to be specified")
		}

		if err := postDecisionComment(ctx, repoStr, prNumber, decision, profileName, verbose); err != nil {
			return fmt.Errorf("posting comment: %w", err)
		}
	}

	// Output result
	format := viper.GetString("format")
	return outputDecision(decision, format, verbose)
}

// getPolicyContext builds the policy context from various sources.
func getPolicyContext(ctx context.Context, verbose bool) (*model.PolicyContext, error) {
	// Check for stdin input
	if viper.GetBool("policy.stdin") {
		return readContextFromStdin()
	}

	// Check for context file
	if contextFile := viper.GetString("policy.context-file"); contextFile != "" {
		return readContextFromFile(contextFile)
	}

	// Fetch from GitHub
	repoStr := viper.GetString("policy.repo")
	prNumber := viper.GetInt("policy.pr")

	if repoStr == "" || prNumber == 0 {
		return nil, fmt.Errorf("either --repo and --pr, --stdin, or --context-file is required")
	}

	return fetchContextFromGitHub(ctx, repoStr, prNumber, verbose)
}

// readContextFromStdin reads policy context from stdin as JSON.
func readContextFromStdin() (*model.PolicyContext, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}

	var pctx model.PolicyContext
	if err := json.Unmarshal(data, &pctx); err != nil {
		return nil, fmt.Errorf("parsing policy context JSON: %w", err)
	}

	return &pctx, nil
}

// readContextFromFile reads policy context from a JSON file.
func readContextFromFile(path string) (*model.PolicyContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading context file: %w", err)
	}

	var pctx model.PolicyContext
	if err := json.Unmarshal(data, &pctx); err != nil {
		return nil, fmt.Errorf("parsing policy context JSON: %w", err)
	}

	return &pctx, nil
}

// fetchContextFromGitHub fetches PR data from GitHub and builds policy context.
func fetchContextFromGitHub(ctx context.Context, repoStr string, prNumber int, verbose bool) (*model.PolicyContext, error) {
	token := viper.GetString("token")
	if token == "" {
		return nil, fmt.Errorf("GitHub token required. Set GITHUB_TOKEN or use --token flag")
	}

	// Use the concrete GitHubCollector to access AnalyzePRForGoMod
	coll, err := collector.NewGitHubCollector(token)
	if err != nil {
		return nil, fmt.Errorf("creating collector: %w", err)
	}

	ref := model.ParseRepoRef(repoStr)
	if verbose {
		fmt.Fprintf(os.Stderr, "Fetching PR #%d from %s...\n", prNumber, ref.FullName())
	}

	// Get PR details
	pr, err := coll.GetPRDetails(ctx, ref, prNumber)
	if err != nil {
		return nil, fmt.Errorf("fetching PR: %w", err)
	}

	// Get CI checks
	checks, err := coll.GetPRChecks(ctx, ref, prNumber)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to get checks: %v\n", err)
		}
		checks = nil
	}

	// Analyze go.mod changes
	goModAnalysis, err := coll.AnalyzePRForGoMod(ctx, ref, prNumber)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze go.mod: %v\n", err)
		}
		goModAnalysis = &collector.GoModAnalysis{}
	}

	// Get profile
	profileName := viper.GetString("policy.profile")
	profile := policy.GetProfile(profileName)
	if profile == nil {
		profile = &policy.ProfileBalanced
	}

	// Build policy context
	return buildPolicyContext(ref, pr, checks, goModAnalysis, profile), nil
}

// buildPolicyContext constructs PolicyContext from GitHub data.
func buildPolicyContext(ref model.RepoRef, pr *model.PullRequest, checks []model.CheckRun, goModAnalysis *collector.GoModAnalysis, profile *model.MergeProfile) *model.PolicyContext {
	// Build CI context
	ciCtx := model.CIContext{}
	if len(checks) > 0 {
		for _, check := range checks {
			switch check.Conclusion {
			case "success":
				ciCtx.PassedChecks = append(ciCtx.PassedChecks, check.Name)
			case "failure":
				ciCtx.FailedChecks = append(ciCtx.FailedChecks, check.Name)
			default:
				if check.Status == "in_progress" || check.Status == "queued" {
					ciCtx.PendingChecks = append(ciCtx.PendingChecks, check.Name)
				}
			}
		}

		ciCtx.AllPassed = len(ciCtx.FailedChecks) == 0 && len(ciCtx.PendingChecks) == 0 && len(ciCtx.PassedChecks) > 0
		ciCtx.AnyFailed = len(ciCtx.FailedChecks) > 0
		ciCtx.AnyPending = len(ciCtx.PendingChecks) > 0
		ciCtx.RequiredPassed = ciCtx.AllPassed // TODO: check required checks specifically
	}

	// Get file info from go.mod analysis
	changedFiles, onlyGoModFiles, changedFileExts := goModAnalysis.ToPRContextFields()

	// Calculate PR age in days
	ageDays := pr.AgeHours() / 24

	// Detect conflicts from mergeable state
	hasConflicts := pr.MergeableStr == "dirty" || pr.MergeableStr == "conflicting"

	return &model.PolicyContext{
		Repo: model.RepoContext{
			Owner:    ref.Owner,
			Name:     ref.Name,
			FullName: ref.FullName(),
			// Note: Private, Archived, Language, Topics would need additional API call
		},
		PR: model.PRContext{
			Number:          pr.Number,
			Title:           pr.Title,
			Author:          pr.Author,
			IsDependency:    pr.IsDependency,
			DependBot:       string(pr.DependBot),
			AgeHours:        pr.AgeHours(),
			AgeDays:         ageDays,
			Mergeable:       pr.Mergeable,
			Draft:           pr.Draft,
			Labels:          pr.Labels,
			HasConflicts:    hasConflicts,
			ChangedFiles:    changedFiles,
			OnlyGoModFiles:  onlyGoModFiles,
			ChangedFileExts: changedFileExts,
		},
		Dependency: model.DependencyContext{
			Name:        pr.Dependency.Name,
			Ecosystem:   pr.Dependency.Ecosystem,
			FromVersion: pr.Dependency.FromVersion,
			ToVersion:   pr.Dependency.ToVersion,
			UpdateType:  string(pr.Dependency.UpdateType),
			IsMajor:     pr.Dependency.UpdateType == model.UpdateTypeMajor,
			IsMinor:     pr.Dependency.UpdateType == model.UpdateTypeMinor,
			IsPatch:     pr.Dependency.UpdateType == model.UpdateTypePatch,
		},
		CI:    ciCtx,
		GoMod: goModAnalysis.ToGoModContext(),
		Profile: model.ProfileContext{
			Name:         profile.Name,
			MinAgeHours:  profile.MinAgeHours,
			MinAgeDays:   profile.MinAgeDays,
			MaxPRsPerRun: profile.MaxPRsPerRun,
		},
	}
}

// postDecisionComment posts or updates a decision comment on the PR.
func postDecisionComment(ctx context.Context, repoStr string, prNumber int, decision *model.PolicyDecision, profileName string, verbose bool) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("GitHub token required for posting comments")
	}

	coll, err := collector.NewGitHubCollector(token)
	if err != nil {
		return fmt.Errorf("creating collector: %w", err)
	}

	ref := model.ParseRepoRef(repoStr)
	dryRun := viper.GetBool("dry-run")

	// Format the comment
	commentBody := policy.FormatDecisionComment(decision, profileName)

	if dryRun {
		fmt.Fprintf(os.Stderr, "Would post comment to %s#%d:\n%s\n", ref.FullName(), prNumber, commentBody)
		return nil
	}

	// Check if we should update an existing comment
	if viper.GetBool("policy.update-comment") {
		existing, err := coll.FindBotCommentByMarker(ctx, ref, prNumber, policy.CommentMarker)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to find existing comment: %v\n", err)
			}
		} else if existing != nil {
			// Update existing comment
			_, err := coll.UpdatePRComment(ctx, ref, existing.ID, commentBody)
			if err != nil {
				return fmt.Errorf("updating comment: %w", err)
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "Updated existing comment on %s#%d\n", ref.FullName(), prNumber)
			}
			return nil
		}
	}

	// Create new comment
	_, err = coll.CreatePRComment(ctx, ref, prNumber, commentBody)
	if err != nil {
		return fmt.Errorf("creating comment: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Posted comment to %s#%d\n", ref.FullName(), prNumber)
	}

	return nil
}

// outputDecision formats and outputs the policy decision.
func outputDecision(decision *model.PolicyDecision, format string, verbose bool) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(decision, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling decision: %w", err)
		}
		fmt.Println(string(data))

	default:
		// Human-readable format
		if decision.Allowed {
			fmt.Printf("✅ ALLOWED: %s\n", decision.Outcome)
		} else {
			fmt.Printf("❌ DENIED: %s\n", decision.Outcome)
		}

		if len(decision.RequiredActions) > 0 {
			fmt.Println("\nRequired Actions:")
			for _, action := range decision.RequiredActions {
				fmt.Printf("  • %s\n", action)
			}
		}

		if len(decision.Reasons) > 0 {
			fmt.Println("\nReasons:")
			for _, reason := range decision.Reasons {
				fmt.Printf("  • %s\n", reason)
			}
		}

		if verbose && len(decision.Policies) > 0 {
			fmt.Println("\nMatching Policies:")
			for _, p := range decision.Policies {
				fmt.Printf("  • %s\n", p)
			}
		}

		if verbose && decision.Evidence != nil {
			fmt.Println("\nEvidence:")
			fmt.Printf("  Author:     %s\n", decision.Evidence.Author)
			fmt.Printf("  PR Age:     %s\n", decision.Evidence.PRAge)
			fmt.Printf("  Update:     %s\n", decision.Evidence.UpdateType)
			fmt.Printf("  Files:      %s\n", decision.Evidence.FilesChanged)
			fmt.Printf("  CI Status:  %s\n", decision.Evidence.CIStatus)
			fmt.Printf("  Directives: %s\n", decision.Evidence.Directives)
		}
	}

	return nil
}
