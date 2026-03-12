package permission

type Service interface{}
type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo} }
