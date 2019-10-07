package api

import (
	"fmt"
	"math"
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
		ID          string        `json:"id"`
		ProjectName string        `json:"project_name"`
		Type        string        `json:"type"`
		Start       *time.Time    `json:"start"`
		End         *time.Time    `json:"end"`
		Breaks      time.Duration `json:"breaks"`
		Comment     string        `json:"comment"`
	}
)

func (p *Project) String() string {
	return fmt.Sprintf("Project: %s (%s) (billable: %t)", p.Name, p.Description, p.Billable)
}

func formatEntryTime(entryType string, entryTime *time.Time) string {
	if entryTime == nil {
		return "                "
	}
	if entryType == "work" {
		return entryTime.In(time.Local).Format("2006-01-02 15:04")
	} else {
		return entryTime.In(time.Local).Format("2006-01-02      ")
	}
}

func formatEntryBreaks(entryType string, breaks time.Duration) string {
	if entryType == "work" && breaks > 0 {
		seconds := int64(breaks.Seconds())
		return fmt.Sprintf("(-%2dh%2dm)", seconds/3600, int(math.Round(float64(seconds%3600)/60.0)))
	}
	return fmt.Sprintf("         ")
}

func formatEntryDuration(e *TimeEntry) string {
	start := time.Now()
	if e.Start != nil {
		start = *e.Start
	}
	end := time.Now()
	if e.End != nil {
		end = *e.End
	}
	duration := end.Sub(start)
	duration -= e.Breaks

	if e.Type == "work" {
		seconds := int64(duration.Seconds())
		if seconds > 86400 {
			return fmt.Sprintf("%2dd%2dh  ", seconds/86400, int(math.Round(float64(seconds%86400)/60.0)))
		} else {
			return fmt.Sprintf("  %2dh%2dm", seconds/3600, int(math.Round(float64(seconds%3600)/60.0)))
		}
	} else {
		days := int(math.Round(duration.Round(24*time.Hour).Hours()/24 + 1))
		return fmt.Sprintf("%2dd     ", days)
	}
}

func (e *TimeEntry) String() string {
	return fmt.Sprintf("%s %s - %s %s (%s): %s %s %s",
		e.ID,
		formatEntryTime(e.Type, e.Start),
		formatEntryTime(e.Type, e.End),
		formatEntryBreaks(e.Type, e.Breaks),
		formatEntryDuration(e),
		e.ProjectName,
		e.Type,
		e.Comment)
}

func (e *TimeEntry) StartsBefore(other *TimeEntry) bool {
	if e.Start == nil {
		return false
	}
	if other.Start == nil {
		return true
	}
	return e.Start.Before(*(other.Start))
}
