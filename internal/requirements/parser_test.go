package requirements_test

import (
	"testing"

	"github.com/mchau/aiguard/internal/requirements"
)

const fixtureACSection = `# Ticket: Add user reason field

## Description

Users should be able to enter a reason.

## Acceptance Criteria

- AC must be visible in the UI
- The reason must persist to the database
- Audit log must record the reason

## Non-Goals

- Not needed for admin users
`

const fixtureNumbered = `# Ticket

## Acceptance Criteria

1. User can submit form
2. Form validates on submit
3. Success message shown
`

const fixtureCheckboxes = `# Ticket

## Definition of Done

- [ ] Feature is implemented
- [x] Tests are written
- [ ] Documentation updated
`

const fixtureNoACSection = `# Ticket

## Description

This ticket has no acceptance criteria section.

## Notes

Just some notes.
`

func TestExtractBulletAC(t *testing.T) {
	acs := requirements.ExtractSourceAC(fixtureACSection)
	if len(acs) != 3 {
		t.Fatalf("expected 3 ACs, got %d", len(acs))
	}
	for i, ac := range acs {
		if ac.Source != "ticket_acceptance_criteria" {
			t.Errorf("AC%d source: got %q", i+1, ac.Source)
		}
		if ac.ID != "AC"+itoa(i+1) {
			t.Errorf("AC%d ID: got %q", i+1, ac.ID)
		}
	}
}

func TestExtractNumberedAC(t *testing.T) {
	acs := requirements.ExtractSourceAC(fixtureNumbered)
	if len(acs) != 3 {
		t.Fatalf("expected 3 ACs, got %d", len(acs))
	}
	if acs[0].Text != "User can submit form" {
		t.Errorf("unexpected text: %q", acs[0].Text)
	}
}

func TestExtractCheckboxAC(t *testing.T) {
	acs := requirements.ExtractSourceAC(fixtureCheckboxes)
	if len(acs) != 3 {
		t.Fatalf("expected 3 ACs (including checked), got %d", len(acs))
	}
}

func TestExtractNoACSection(t *testing.T) {
	acs := requirements.ExtractSourceAC(fixtureNoACSection)
	if len(acs) != 0 {
		t.Errorf("expected 0 ACs, got %d", len(acs))
	}
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
