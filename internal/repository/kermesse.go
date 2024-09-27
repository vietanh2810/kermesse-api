package repository

type KermesseDAO interface {
}

type KermesseRepository struct {
	dao KermesseDAO
}

func NewKermesseRepository(dao KermesseDAO) *KermesseRepository {
	return &KermesseRepository{
		dao: dao,
	}
}
