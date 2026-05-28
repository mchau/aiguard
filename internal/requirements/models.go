package requirements

// AcceptanceCriterion is a single AC item from a ticket or inferred by AI.
type AcceptanceCriterion struct {
	ID          string `json:"id"`           // AC1, AC2, ...
	Text        string `json:"text"`
	Source      string `json:"source"`       // "ticket_acceptance_criteria" or "inferred"
	Status      string `json:"status"`       // open, met, not_applicable
	PlanStep    string `json:"plan_step,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// RequirementDigest is the normalized output of requirement ingestion.
type RequirementDigest struct {
	RawSummary          string                `json:"raw_summary,omitempty"`
	NormalizedCriteria  []AcceptanceCriterion `json:"normalized_criteria"`
	OpenQuestions       []ClarificationQuestion `json:"open_questions,omitempty"`
	Assumptions         []Assumption          `json:"assumptions,omitempty"`
	NonGoals            []string              `json:"non_goals,omitempty"`
}

// ClarificationQuestion is a question raised during digestion.
type ClarificationQuestion struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Blocking bool   `json:"blocking"`
	Status   string `json:"status"` // open, answered, assumed
	Answer   string `json:"answer,omitempty"`
}

// Assumption is a documented assumption made in lieu of a clarification answer.
type Assumption struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Risk   string `json:"risk,omitempty"`
}
