package model

import "fmt"

// PolicyContext provides context for Cedar policy evaluation.
// This struct is serialized to JSON for Cedar entity evaluation.
type PolicyContext struct {
	Repo       RepoContext       `json:"repo"`
	PR         PRContext         `json:"pr"`
	Dependency DependencyContext `json:"dependency"`
	CI         CIContext         `json:"ci"`
	GoMod      GoModContext      `json:"goMod"`
	Profile    ProfileContext    `json:"profile"`
}

// ProfileContext contains merge profile settings for policy evaluation.
// This allows Cedar policies to reference profile-configured thresholds.
type ProfileContext struct {
	Name         string `json:"name"`
	MinAgeHours  int    `json:"minAgeHours"`
	MinAgeDays   int    `json:"minAgeDays"`
	MaxPRsPerRun int    `json:"maxPRsPerRun"`
}

// RepoContext contains repository information for policy evaluation.
type RepoContext struct {
	Owner      string   `json:"owner"`
	Name       string   `json:"name"`
	FullName   string   `json:"fullName"`
	Private    bool     `json:"private"`
	Archived   bool     `json:"archived"`
	Language   string   `json:"language"`
	Topics     []string `json:"topics"`
	IsMonorepo bool     `json:"isMonorepo"`
}

// PRContext contains pull request information for policy evaluation.
type PRContext struct {
	Number       int      `json:"number"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	IsDependency bool     `json:"isDependency"`
	DependBot    string   `json:"dependBot"`
	AgeHours     int      `json:"ageHours"`
	AgeDays      int      `json:"ageDays"`
	Mergeable    bool     `json:"mergeable"`
	Draft        bool     `json:"draft"`
	Labels       []string `json:"labels"`
	HasConflicts bool     `json:"hasConflicts"`

	// File gate fields
	ChangedFiles    []string `json:"changedFiles"`
	OnlyGoModFiles  bool     `json:"onlyGoModFiles"`
	ChangedFileExts []string `json:"changedFileExts"`
}

// DependencyContext contains dependency update information for policy evaluation.
type DependencyContext struct {
	Name        string `json:"name"`
	Ecosystem   string `json:"ecosystem"`
	FromVersion string `json:"fromVersion"`
	ToVersion   string `json:"toVersion"`
	UpdateType  string `json:"updateType"`
	IsMajor     bool   `json:"isMajor"`
	IsMinor     bool   `json:"isMinor"`
	IsPatch     bool   `json:"isPatch"`
}

// GoModContext contains go.mod-specific analysis for policy evaluation.
// This is used to detect potentially dangerous changes beyond version bumps.
type GoModContext struct {
	// Directive changes
	HasReplaceChange    bool `json:"hasReplaceChange"`
	HasExcludeChange    bool `json:"hasExcludeChange"`
	HasRetractChange    bool `json:"hasRetractChange"`
	HasToolchainChange  bool `json:"hasToolchainChange"`
	HasGoVersionChange  bool `json:"hasGoVersionChange"`
	HasDirectiveChanges bool `json:"hasDirectiveChanges"` // any of the above

	// Dependency changes
	HasNewDirectDependency bool     `json:"hasNewDirectDependency"`
	NewDirectDependencies  []string `json:"newDirectDependencies,omitempty"`
	RemovedDependencies    []string `json:"removedDependencies,omitempty"`

	// Version info
	OldGoVersion string `json:"oldGoVersion,omitempty"`
	NewGoVersion string `json:"newGoVersion,omitempty"`
}

// CIContext contains CI/test status for policy evaluation.
type CIContext struct {
	AllPassed      bool     `json:"allPassed"`
	AnyFailed      bool     `json:"anyFailed"`
	AnyPending     bool     `json:"anyPending"`
	PassedChecks   []string `json:"passedChecks"`
	FailedChecks   []string `json:"failedChecks"`
	PendingChecks  []string `json:"pendingChecks"`
	RequiredPassed bool     `json:"requiredPassed"`
}

// PolicyAction represents an action that can be evaluated against policies.
type PolicyAction string

const (
	PolicyActionReview  PolicyAction = "review"
	PolicyActionMerge   PolicyAction = "merge"
	PolicyActionRelease PolicyAction = "release"
)

// DecisionOutcome represents the type of action to take based on policy evaluation.
type DecisionOutcome string

const (
	// DecisionAutoApprove indicates the PR should be approved but not merged.
	DecisionAutoApprove DecisionOutcome = "AUTO_APPROVE"

	// DecisionAutoMerge indicates the PR should be approved and auto-merge enabled.
	DecisionAutoMerge DecisionOutcome = "AUTO_MERGE"

	// DecisionQueueForMerge indicates the PR should be added to merge queue.
	DecisionQueueForMerge DecisionOutcome = "QUEUE_FOR_MERGE"

	// DecisionManualReview indicates the PR needs human review.
	DecisionManualReview DecisionOutcome = "MANUAL_REVIEW"

	// DecisionSecurityReview indicates the PR needs security team review.
	DecisionSecurityReview DecisionOutcome = "SECURITY_TEAM_REVIEW"

	// DecisionReject indicates the PR should not be merged.
	DecisionReject DecisionOutcome = "REJECT"
)

// RequiredAction represents an action that should be taken on the PR.
type RequiredAction string

const (
	ActionApprove         RequiredAction = "approve"
	ActionEnableAutoMerge RequiredAction = "enable_auto_merge"
	ActionAddToQueue      RequiredAction = "add_to_queue"
	ActionAddLabel        RequiredAction = "add_label"
	ActionComment         RequiredAction = "comment"
	ActionRequestReview   RequiredAction = "request_review"
)

// PolicyDecision represents the result of policy evaluation.
type PolicyDecision struct {
	// Allowed indicates whether the action is permitted.
	Allowed bool `json:"allowed"`

	// Action is the evaluated action (merge, review, release).
	Action string `json:"action"`

	// Outcome specifies what action to take (AUTO_MERGE, MANUAL_REVIEW, etc.).
	Outcome DecisionOutcome `json:"outcome"`

	// RequiredActions lists the specific actions to perform.
	RequiredActions []RequiredAction `json:"requiredActions,omitempty"`

	// Reasons explains why the decision was made.
	Reasons []string `json:"reasons,omitempty"`

	// Policies lists the policy IDs that contributed to the decision.
	Policies []string `json:"policies,omitempty"`

	// Evidence summarizes the context used for the decision.
	Evidence *DecisionEvidence `json:"evidence,omitempty"`
}

// DecisionEvidence captures the key facts used in the decision.
type DecisionEvidence struct {
	Author       string `json:"author"`
	IsBot        bool   `json:"isBot"`
	PRAge        string `json:"prAge"`
	UpdateType   string `json:"updateType"`
	Ecosystem    string `json:"ecosystem"`
	FilesChanged string `json:"filesChanged"`
	CIStatus     string `json:"ciStatus"`
	Directives   string `json:"directives"`
}

// Summary returns a human-readable summary of the decision.
func (d *PolicyDecision) Summary() string {
	if d.Allowed {
		return string(d.Outcome) + ": " + d.joinReasons()
	}
	return "DENIED: " + d.joinReasons()
}

func (d *PolicyDecision) joinReasons() string {
	if len(d.Reasons) == 0 {
		return "no specific reason"
	}
	if len(d.Reasons) == 1 {
		return d.Reasons[0]
	}
	return fmt.Sprintf("%s (+%d more)", d.Reasons[0], len(d.Reasons)-1)
}

// MergeProfile defines a set of merge policies and behaviors.
type MergeProfile struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`

	// Timing controls
	MinAgeHours int `json:"minAgeHours" yaml:"minAgeHours"`
	MinAgeDays  int `json:"minAgeDays" yaml:"minAgeDays"` // N+5 quarantine period
	MaxAgeHours int `json:"maxAgeHours" yaml:"maxAgeHours"`

	// Update type controls
	AutoMergePatch bool `json:"autoMergePatch" yaml:"autoMergePatch"`
	AutoMergeMinor bool `json:"autoMergeMinor" yaml:"autoMergeMinor"`
	AutoMergeMajor bool `json:"autoMergeMajor" yaml:"autoMergeMajor"`

	// CI requirements
	RequireAllChecks   bool     `json:"requireAllChecks" yaml:"requireAllChecks"`
	RequiredChecks     []string `json:"requiredChecks,omitempty" yaml:"requiredChecks,omitempty"`
	AllowPendingChecks bool     `json:"allowPendingChecks" yaml:"allowPendingChecks"`

	// Merge settings
	MergeStrategy string `json:"mergeStrategy" yaml:"mergeStrategy"` // merge, squash, rebase
	DeleteBranch  bool   `json:"deleteBranch" yaml:"deleteBranch"`

	// Safety
	RequireApproval bool `json:"requireApproval" yaml:"requireApproval"`
	MaxPRsPerRun    int  `json:"maxPRsPerRun" yaml:"maxPRsPerRun"`
}
