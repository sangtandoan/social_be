package repo

type Repo struct {
	User UserRepo
}

func NewRepo() *Repo {
	return &Repo{
		User: NewUserRepo(),
	}
}
