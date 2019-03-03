package domain

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
		ID          string `json:"id"`
		UserID      string `json:"userid"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Billable    bool   `json:"billable"`
	}

	TimeEntry struct {
		ID        string        `json:"id"`
		ProjectID string        `json:"projectid"`
		UserID    string        `json:"userid"`
		Start     time.Time     `json:"start"`
		End       time.Time     `json:"end"`
		Breaks    time.Duration `json:"breaks"`
		Comment   string        `json:"comment"`
	}

	DayEntry struct {
		ID     string    `json:"id"`
		Type   string    `json:"type"`
		UserID string    `json:"userid"`
		Date   time.Time `json:"date"`
	}

	// UserContext represent what information we have about the loggged-in user
	UserContext struct {
		SubjectID string
		Issuer    string
		Email     string
	}

	// ApplicationContext is our application-specific request context
	ApplicationContext struct {
		RequestID string
		User      UserContext
	}
)

var DayTypes = []string{"sick", "sick-child", "vacation"}

func (p *Project) String() string {
	return fmt.Sprintf("Project: %s (%s) (billable: %t)", p.Name, p.Description, p.Billable)
}
