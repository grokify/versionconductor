package model

// PolicyContext provides context for Cedar policy evaluation.
// This struct is serialized to JSON for Cedar entity evaluation.
type PolicyContext struct {
	Repo       RepoContext       `json:"repo"`
	PR         PRContext         `json:"pr"`
	Dependency DependencyContext `json:"dependency"`
	CI         CIContext         `json:"ci"`
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

// PolicyDecision represents the result of policy evaluation.
type PolicyDecision struct {
	Allowed  bool     `json:"allowed"`
	Action   string   `json:"action"`
	Reasons  []string `json:"reasons,omitempty"`
	Policies []string `json:"policies,omitempty"`
}

// MergeProfile defines a set of merge policies and behaviors.
type MergeProfile struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`

	// Timing controls
	MinAgeHours int `json:"minAgeHours" yaml:"minAgeHours"`
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
