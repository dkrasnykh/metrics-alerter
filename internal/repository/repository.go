package repository

type Storage interface {
	Create(name string, value any) error
	Get(name string) (any, error)
	GetAll() (any, error)
	Update(name string, value any) error
	Delete(name string) error
}
