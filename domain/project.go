package domain

type (
	Project struct {
		aggregateRoot
		ID          string
		User        *User
		Name        string
		Description string
		Billable    bool
	}
)

func (p *Project) Save() error {
	return p.repo.UpdateProject(p)
}

func (p *Project) GetEntries() ([]*TimeEntry, error) {
	timeentries, err := p.repo.GetProjectTimeEntries(p)
	if err != nil {
		return nil, err
	}
	for _, entry := range timeentries {
		p.copyDeps(&entry.aggregateRoot)
		entry.Project = p
	}
	return timeentries, nil
}
