package api

import (
	"fmt"
	"time"
)

type (
	Identity struct {
		Issuer    string `json:"provider"`
		SubjectID string `json:"subjectid"`
		Email     string `json:"email"`
	}

	User struct {
		ID         string     `json:"id"`
		Identities []Identity `json:"identities"`
	}

	Project struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Billable    bool   `json:"billable"`
	}

	TimeEntry struct {
		Type    string        `json:"type"`
		Start   time.Time     `json:"start"`
		End     time.Time     `json:"end"`
		Breaks  time.Duration `json:"breaks"`
		Comment string        `json:"comment"`
	}
)

func (p *Project) String() string {
	return fmt.Sprintf("Project: %s (%s) (billable: %t)", p.Name, p.Description, p.Billable)
}
