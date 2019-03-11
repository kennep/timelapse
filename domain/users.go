package domain

type Users struct {
	aggregateRoot
}

func InitUsersCollection(repo TimelapseRepository) *Users {
	var users Users
	users.repo = repo
	return &users
}

func (u *Users) GetOrCreateUserFromContext(ctx *ApplicationContext) (*User, error) {
	user, err := u.repo.CreateUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	u.copyDeps(&user.aggregateRoot)
	return user, nil
}
