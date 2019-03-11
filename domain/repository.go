package domain

type TimelapseRepository interface {
	CreateUserFromContext(appCtx *ApplicationContext) (*User, error)
	AddProject(p *Project) (*Project, error)
	UpdateProject(p *Project) error
	GetProject(u *User, projectName string) (*Project, error)
	GetProjects(u *User) ([]*Project, error)
	AddTimeEntry(u *User, projectName string, e TimeEntry) (*TimeEntry, error)
	GetProjectTimeEntries(p *Project) ([]*TimeEntry, error)
	GetUserTimeEntries(u *User) ([]*TimeEntry, error)
}
