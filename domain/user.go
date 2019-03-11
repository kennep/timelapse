package domain

type (
	Identity struct {
		Issuer    string
		SubjectID string
		Email     string
	}

	User struct {
		aggregateRoot
		ID         string
		Identities []Identity
	}
)

func (u *User) GetProject(projectName string) (*Project, error) {
	project, err := u.repo.GetProject(u, projectName)
	if err != nil {
		return nil, err
	}
	u.copyDeps(&project.aggregateRoot)
	return project, nil
}

func (u *User) AddProject(project *Project) (*Project, error) {
	project.User = u
	project, err := u.repo.AddProject(project)
	if err != nil {
		return nil, err
	}
	u.copyDeps(&project.aggregateRoot)
	return project, nil
}

func (u *User) GetProjects() ([]*Project, error) {
	projects, err := u.repo.GetProjects(u)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		u.copyDeps(&project.aggregateRoot)
	}
	return projects, nil
}

func (u *User) GetEntries() ([]*TimeEntry, error) {
	timeentries, err := u.repo.GetUserTimeEntries(u)
	if err != nil {
		return nil, err
	}
	for _, entry := range timeentries {
		u.copyDeps(&entry.aggregateRoot)
	}
	return timeentries, nil

}
