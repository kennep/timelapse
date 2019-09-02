package domain

type TimelapseRepository interface {
	CreateUserFromContext(appCtx *ApplicationContext) (*User, error)
	AddProject(p Project) (*Project, error)
	UpdateProject(p Project) (*Project, error)
	GetProject(u *User, projectName string) (*Project, error)
	GetProjects(u *User) ([]*Project, error)
	AddTimeEntry(p *Project, e TimeEntry) (*TimeEntry, error)
	UpdateTimeEntry(p *Project, e TimeEntry) (*TimeEntry, error)
	GetProjectTimeEntries(p *Project) ([]*TimeEntry, error)
	GetProjectTimeEntry(p *Project, entryID string) (*TimeEntry, error)
	GetUserTimeEntries(u *User) ([]*TimeEntry, error)
}
