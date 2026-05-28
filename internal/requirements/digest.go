package requirements

// BuildDigest merges source and inferred ACs into a RequirementDigest.
// Source-provided ACs are always preserved — inferred ACs are appended,
// never allowed to replace source ACs.
func BuildDigest(raw string, inferred []AcceptanceCriterion) *RequirementDigest {
	source := ExtractSourceAC(raw)

	// Renumber inferred to avoid ID collisions with source IDs
	nextID := len(source) + 1
	for i := range inferred {
		inferred[i].ID = formatAcID(nextID)
		inferred[i].Source = "inferred"
		nextID++
	}

	normalized := make([]AcceptanceCriterion, 0, len(source)+len(inferred))
	normalized = append(normalized, source...)
	normalized = append(normalized, inferred...)

	return &RequirementDigest{
		NormalizedCriteria: normalized,
	}
}

func formatAcID(n int) string {
	return "AC" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
