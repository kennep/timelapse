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
	_, err := p.repo.UpdateProject(*p)
	return err
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

func (p *Project) AddEntry(entry *TimeEntry) (*TimeEntry, error) {
	return p.repo.AddTimeEntry(p, *entry)
}

func (p *Project) UpdateEntry(entry *TimeEntry) (*TimeEntry, error) {
	entry.Project = p
	return p.repo.UpdateTimeEntry(p, *entry)
}

func (p *Project) GetEntry(entryID string) (*TimeEntry, error) {
	return p.repo.GetProjectTimeEntry(p, entryID)
}
