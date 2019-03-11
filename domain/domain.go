package domain

import (
	"fmt"
)

type (
	aggregateRoot struct {
		repo TimelapseRepository
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

var EntryTypes = []string{"normal", "sick", "sick-child", "vacation"}

func (p *Project) String() string {
	return fmt.Sprintf("Project: %s (%s) (billable: %t)", p.Name, p.Description, p.Billable)
}

func (src *aggregateRoot) copyDeps(dest *aggregateRoot) {
	dest.repo = src.repo
}
