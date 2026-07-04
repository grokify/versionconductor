package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"

	"github.com/plexusone/versionconductor/pkg/model"
)

// CedarEngine evaluates Cedar policies for PR merge decisions.
type CedarEngine struct {
	policySet         *cedar.PolicySet
	profile           *model.MergeProfile
	policyAnnotations map[cedar.PolicyID]policyAnnotation
}

// policyAnnotation stores extracted annotations from a Cedar policy.
type policyAnnotation struct {
	id          string                // @id annotation
	action      model.DecisionOutcome // @action annotation
	description string                // @description annotation
}

// NewCedarEngine creates a new Cedar policy engine.
// If policyPath is empty, uses the default profile-based evaluation.
// If policyPath is provided, loads Cedar policies from that file or directory.
func NewCedarEngine(policyPath string, profileName string) (*CedarEngine, error) {
	profile := GetProfile(profileName)
	if profile == nil {
		profile = &ProfileBalanced
	}

	engine := &CedarEngine{
		policySet:         cedar.NewPolicySet(),
		profile:           profile,
		policyAnnotations: make(map[cedar.PolicyID]policyAnnotation),
	}

	if policyPath != "" {
		if err := engine.LoadPolicies(policyPath); err != nil {
			return nil, fmt.Errorf("loading policies: %w", err)
		}
	}

	return engine, nil
}

// LoadPolicies loads Cedar policies from a file or directory.
func (e *CedarEngine) LoadPolicies(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if info.IsDir() {
		return e.loadPoliciesFromDir(path)
	}

	return e.loadPolicyFile(path)
}

// loadPoliciesFromDir loads all .cedar files from a directory recursively.
func (e *CedarEngine) loadPoliciesFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recurse into subdirectories
			if err := e.loadPoliciesFromDir(path); err != nil {
				return err
			}
			continue
		}

		if filepath.Ext(entry.Name()) != ".cedar" {
			continue
		}

		if err := e.loadPolicyFile(path); err != nil {
			return err
		}
	}

	return nil
}

// loadPolicyFile loads Cedar policies from a single file.
func (e *CedarEngine) loadPolicyFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading policy file %s: %w", path, err)
	}

	policies, err := cedar.NewPolicyListFromBytes(path, content)
	if err != nil {
		return fmt.Errorf("parsing policy file %s: %w", path, err)
	}

	// Add each policy to the set with a unique ID
	baseName := filepath.Base(path)
	for i, p := range policies {
		policyID := cedar.PolicyID(fmt.Sprintf("%s:%d", baseName, i))
		e.policySet.Add(policyID, p)

		// Extract annotations from the policy
		ann := e.extractAnnotations(p)
		e.policyAnnotations[policyID] = ann
	}

	return nil
}

// extractAnnotations extracts relevant annotations from a Cedar policy.
func (e *CedarEngine) extractAnnotations(p *cedar.Policy) policyAnnotation {
	ann := policyAnnotation{}
	annotations := p.Annotations()

	if id, ok := annotations[types.Ident("id")]; ok {
		ann.id = string(id)
	}
	if action, ok := annotations[types.Ident("action")]; ok {
		ann.action = model.DecisionOutcome(action)
	}
	if desc, ok := annotations[types.Ident("description")]; ok {
		ann.description = string(desc)
	}

	return ann
}

// Evaluate evaluates the policy for the given action and context.
func (e *CedarEngine) Evaluate(ctx context.Context, action model.PolicyAction, pctx *model.PolicyContext) (*model.PolicyDecision, error) {
	result := &model.PolicyDecision{
		Action: string(action),
	}

	// If no policies are loaded, fall back to profile-based evaluation
	if e.policySet == nil || !e.hasPolicies() {
		return e.evaluateWithProfile(ctx, action, pctx)
	}

	// Build Cedar request
	req := e.buildRequest(action, pctx)

	// Evaluate against Cedar policies
	decision, diag := cedar.Authorize(e.policySet, nil, req)

	result.Allowed = decision == cedar.Allow

	// Collect policy IDs and determine strongest outcome
	var outcomes []model.DecisionOutcome
	for _, reason := range diag.Reasons {
		result.Policies = append(result.Policies, string(reason.PolicyID))

		// Get the outcome from the policy's @action annotation
		if ann, ok := e.policyAnnotations[reason.PolicyID]; ok {
			if ann.action != "" {
				outcomes = append(outcomes, ann.action)
			}
			if ann.description != "" {
				result.Reasons = append(result.Reasons, ann.description)
			}
		}
	}

	// Collect any evaluation errors
	for _, diagErr := range diag.Errors {
		result.Reasons = append(result.Reasons, fmt.Sprintf("%s: %s", diagErr.PolicyID, diagErr.Message))
	}

	// Determine the outcome based on matched policies
	result.Outcome = e.determineOutcome(result.Allowed, outcomes)

	// Build required actions based on outcome
	result.RequiredActions = e.buildRequiredActions(result.Outcome)

	// Build evidence from context
	result.Evidence = e.buildEvidence(pctx)

	if !result.Allowed && len(result.Reasons) == 0 {
		result.Reasons = []string{"no policy permitted the action"}
	}

	return result, nil
}

// outcomePriority returns the priority of a decision outcome.
// Higher priority = more restrictive. Forbid outcomes override permit outcomes.
func outcomePriority(o model.DecisionOutcome) int {
	switch o {
	case model.DecisionReject:
		return 100
	case model.DecisionSecurityReview:
		return 80
	case model.DecisionManualReview:
		return 60
	case model.DecisionQueueForMerge:
		return 40
	case model.DecisionAutoMerge:
		return 20
	case model.DecisionAutoApprove:
		return 10
	default:
		return 0
	}
}

// determineOutcome selects the strongest (most restrictive) outcome from matched policies.
func (e *CedarEngine) determineOutcome(allowed bool, outcomes []model.DecisionOutcome) model.DecisionOutcome {
	if len(outcomes) == 0 {
		if allowed {
			return model.DecisionAutoMerge
		}
		return model.DecisionManualReview
	}

	// Find the highest priority outcome
	strongest := outcomes[0]
	for _, o := range outcomes[1:] {
		if outcomePriority(o) > outcomePriority(strongest) {
			strongest = o
		}
	}

	return strongest
}

// buildRequiredActions determines what actions should be taken based on the outcome.
func (e *CedarEngine) buildRequiredActions(outcome model.DecisionOutcome) []model.RequiredAction {
	switch outcome {
	case model.DecisionAutoMerge:
		return []model.RequiredAction{
			model.ActionApprove,
			model.ActionEnableAutoMerge,
		}
	case model.DecisionAutoApprove:
		return []model.RequiredAction{
			model.ActionApprove,
		}
	case model.DecisionQueueForMerge:
		return []model.RequiredAction{
			model.ActionApprove,
			model.ActionAddToQueue,
		}
	case model.DecisionManualReview:
		return []model.RequiredAction{
			model.ActionComment,
			model.ActionAddLabel,
		}
	case model.DecisionSecurityReview:
		return []model.RequiredAction{
			model.ActionComment,
			model.ActionRequestReview,
			model.ActionAddLabel,
		}
	case model.DecisionReject:
		return []model.RequiredAction{
			model.ActionComment,
		}
	default:
		return nil
	}
}

// buildEvidence creates a summary of the context used for the decision.
func (e *CedarEngine) buildEvidence(pctx *model.PolicyContext) *model.DecisionEvidence {
	// Determine CI status string
	ciStatus := "unknown"
	switch {
	case pctx.CI.AllPassed:
		ciStatus = "all passed"
	case pctx.CI.AnyFailed:
		ciStatus = "failed"
	case pctx.CI.AnyPending:
		ciStatus = "pending"
	}

	// Determine directive status
	directives := "none"
	if pctx.GoMod.HasDirectiveChanges {
		directives = "changed"
		if pctx.GoMod.HasReplaceChange {
			directives = "replace"
		} else if pctx.GoMod.HasExcludeChange {
			directives = "exclude"
		} else if pctx.GoMod.HasToolchainChange {
			directives = "toolchain"
		}
	}

	// Build files changed summary
	filesChanged := fmt.Sprintf("%d files", len(pctx.PR.ChangedFiles))
	if pctx.PR.OnlyGoModFiles {
		filesChanged = "go.mod only"
	}

	return &model.DecisionEvidence{
		Author:       pctx.PR.Author,
		IsBot:        pctx.PR.IsDependency,
		PRAge:        fmt.Sprintf("%d days", pctx.PR.AgeDays),
		UpdateType:   pctx.Dependency.UpdateType,
		Ecosystem:    pctx.Dependency.Ecosystem,
		FilesChanged: filesChanged,
		CIStatus:     ciStatus,
		Directives:   directives,
	}
}

// hasPolicies checks if the policy set has any policies.
func (e *CedarEngine) hasPolicies() bool {
	for range e.policySet.All() {
		return true
	}
	return false
}

// evaluateWithProfile falls back to profile-based evaluation when no Cedar policies are loaded.
func (e *CedarEngine) evaluateWithProfile(_ context.Context, action model.PolicyAction, pctx *model.PolicyContext) (*model.PolicyDecision, error) {
	result := &model.PolicyDecision{
		Action: string(action),
	}

	switch action {
	case model.PolicyActionMerge:
		// Build a minimal PullRequest from context for profile evaluation
		pr := &model.PullRequest{
			Number:     pctx.PR.Number,
			Title:      pctx.PR.Title,
			Author:     pctx.PR.Author,
			Mergeable:  pctx.PR.Mergeable,
			Draft:      pctx.PR.Draft,
			Dependency: buildDependencyFromContext(pctx),
		}

		checks := buildChecksFromContext(pctx)
		allowed, reason := EvaluateProfile(e.profile, pr, checks)

		result.Allowed = allowed
		if allowed {
			result.Outcome = model.DecisionAutoMerge
			result.RequiredActions = e.buildRequiredActions(model.DecisionAutoMerge)
		} else {
			result.Outcome = model.DecisionManualReview
			result.RequiredActions = e.buildRequiredActions(model.DecisionManualReview)
			result.Reasons = []string{reason}
		}

	case model.PolicyActionReview:
		if pctx.CI.AllPassed && !pctx.PR.Draft {
			result.Allowed = true
			result.Outcome = model.DecisionAutoApprove
			result.RequiredActions = e.buildRequiredActions(model.DecisionAutoApprove)
		} else {
			result.Allowed = false
			result.Outcome = model.DecisionManualReview
			result.RequiredActions = e.buildRequiredActions(model.DecisionManualReview)
			if !pctx.CI.AllPassed {
				result.Reasons = append(result.Reasons, "CI checks not passed")
			}
			if pctx.PR.Draft {
				result.Reasons = append(result.Reasons, "PR is a draft")
			}
		}

	case model.PolicyActionRelease:
		result.Allowed = true
		result.Outcome = model.DecisionAutoMerge

	default:
		result.Allowed = false
		result.Outcome = model.DecisionReject
		result.Reasons = []string{"unknown action"}
	}

	// Build evidence for all profile-based decisions
	result.Evidence = e.buildEvidence(pctx)

	return result, nil
}

// buildRequest constructs a Cedar Request from PolicyContext.
func (e *CedarEngine) buildRequest(action model.PolicyAction, pctx *model.PolicyContext) cedar.Request {
	return cedar.Request{
		Principal: cedar.NewEntityUID("Bot", cedar.String(pctx.PR.Author)),
		Action:    cedar.NewEntityUID("Action", cedar.String(string(action))),
		Resource:  cedar.NewEntityUID("PullRequest", cedar.String(fmt.Sprintf("%s/%s#%d", pctx.Repo.Owner, pctx.Repo.Name, pctx.PR.Number))),
		Context:   e.buildContext(pctx),
	}
}

// buildContext constructs a Cedar Record from PolicyContext.
func (e *CedarEngine) buildContext(pctx *model.PolicyContext) cedar.Record {
	return cedar.NewRecord(cedar.RecordMap{
		// Repo context
		"repo": cedar.NewRecord(cedar.RecordMap{
			"owner":      cedar.String(pctx.Repo.Owner),
			"name":       cedar.String(pctx.Repo.Name),
			"fullName":   cedar.String(pctx.Repo.FullName),
			"private":    cedar.Boolean(pctx.Repo.Private),
			"archived":   cedar.Boolean(pctx.Repo.Archived),
			"language":   cedar.String(pctx.Repo.Language),
			"isMonorepo": cedar.Boolean(pctx.Repo.IsMonorepo),
		}),

		// PR context
		"pr": cedar.NewRecord(cedar.RecordMap{
			"number":         cedar.Long(int64(pctx.PR.Number)),
			"title":          cedar.String(pctx.PR.Title),
			"author":         cedar.String(pctx.PR.Author),
			"isDependency":   cedar.Boolean(pctx.PR.IsDependency),
			"dependBot":      cedar.String(pctx.PR.DependBot),
			"ageHours":       cedar.Long(int64(pctx.PR.AgeHours)),
			"ageDays":        cedar.Long(int64(pctx.PR.AgeDays)),
			"mergeable":      cedar.Boolean(pctx.PR.Mergeable),
			"draft":          cedar.Boolean(pctx.PR.Draft),
			"hasConflicts":   cedar.Boolean(pctx.PR.HasConflicts),
			"onlyGoModFiles": cedar.Boolean(pctx.PR.OnlyGoModFiles),
		}),

		// Dependency context
		"dependency": cedar.NewRecord(cedar.RecordMap{
			"name":        cedar.String(pctx.Dependency.Name),
			"ecosystem":   cedar.String(pctx.Dependency.Ecosystem),
			"fromVersion": cedar.String(pctx.Dependency.FromVersion),
			"toVersion":   cedar.String(pctx.Dependency.ToVersion),
			"updateType":  cedar.String(pctx.Dependency.UpdateType),
			"isMajor":     cedar.Boolean(pctx.Dependency.IsMajor),
			"isMinor":     cedar.Boolean(pctx.Dependency.IsMinor),
			"isPatch":     cedar.Boolean(pctx.Dependency.IsPatch),
		}),

		// CI context
		"ci": cedar.NewRecord(cedar.RecordMap{
			"allPassed":      cedar.Boolean(pctx.CI.AllPassed),
			"anyFailed":      cedar.Boolean(pctx.CI.AnyFailed),
			"anyPending":     cedar.Boolean(pctx.CI.AnyPending),
			"requiredPassed": cedar.Boolean(pctx.CI.RequiredPassed),
		}),

		// GoMod context
		"goMod": cedar.NewRecord(cedar.RecordMap{
			"hasReplaceChange":       cedar.Boolean(pctx.GoMod.HasReplaceChange),
			"hasExcludeChange":       cedar.Boolean(pctx.GoMod.HasExcludeChange),
			"hasRetractChange":       cedar.Boolean(pctx.GoMod.HasRetractChange),
			"hasToolchainChange":     cedar.Boolean(pctx.GoMod.HasToolchainChange),
			"hasGoVersionChange":     cedar.Boolean(pctx.GoMod.HasGoVersionChange),
			"hasDirectiveChanges":    cedar.Boolean(pctx.GoMod.HasDirectiveChanges),
			"hasNewDirectDependency": cedar.Boolean(pctx.GoMod.HasNewDirectDependency),
		}),

		// Profile context
		"profile": cedar.NewRecord(cedar.RecordMap{
			"name":        cedar.String(pctx.Profile.Name),
			"minAgeHours": cedar.Long(int64(pctx.Profile.MinAgeHours)),
			"minAgeDays":  cedar.Long(int64(pctx.Profile.MinAgeDays)),
		}),
	})
}

// buildDependencyFromContext builds a model.Dependency from PolicyContext.
func buildDependencyFromContext(pctx *model.PolicyContext) model.Dependency {
	updateType := model.UpdateTypeUnknown
	switch {
	case pctx.Dependency.IsMajor:
		updateType = model.UpdateTypeMajor
	case pctx.Dependency.IsMinor:
		updateType = model.UpdateTypeMinor
	case pctx.Dependency.IsPatch:
		updateType = model.UpdateTypePatch
	}

	return model.Dependency{
		Name:        pctx.Dependency.Name,
		Ecosystem:   pctx.Dependency.Ecosystem,
		FromVersion: pctx.Dependency.FromVersion,
		ToVersion:   pctx.Dependency.ToVersion,
		UpdateType:  updateType,
	}
}

// buildChecksFromContext builds check runs from PolicyContext.
func buildChecksFromContext(pctx *model.PolicyContext) []model.CheckRun {
	var checks []model.CheckRun

	for _, name := range pctx.CI.PassedChecks {
		checks = append(checks, model.CheckRun{
			Name:       name,
			Status:     "completed",
			Conclusion: "success",
		})
	}

	for _, name := range pctx.CI.FailedChecks {
		checks = append(checks, model.CheckRun{
			Name:       name,
			Status:     "completed",
			Conclusion: "failure",
		})
	}

	for _, name := range pctx.CI.PendingChecks {
		checks = append(checks, model.CheckRun{
			Name:   name,
			Status: "in_progress",
		})
	}

	return checks
}

// CanMerge evaluates whether a PR can be auto-merged.
func (e *CedarEngine) CanMerge(ctx context.Context, pctx *model.PolicyContext) (*model.PolicyDecision, error) {
	return e.Evaluate(ctx, model.PolicyActionMerge, pctx)
}

// CanReview evaluates whether a PR can be auto-reviewed.
func (e *CedarEngine) CanReview(ctx context.Context, pctx *model.PolicyContext) (*model.PolicyDecision, error) {
	return e.Evaluate(ctx, model.PolicyActionReview, pctx)
}

// Profile returns the merge profile associated with this engine.
func (e *CedarEngine) Profile() *model.MergeProfile {
	return e.profile
}

// PolicyCount returns the number of loaded policies.
func (e *CedarEngine) PolicyCount() int {
	count := 0
	for range e.policySet.All() {
		count++
	}
	return count
}
