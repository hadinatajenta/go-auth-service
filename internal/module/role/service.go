package role

type Service interface {
	// List methods as needed
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}
