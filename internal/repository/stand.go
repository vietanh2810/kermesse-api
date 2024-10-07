package repository

type StandDAO interface {
}

type StandRepository struct {
	dao StandDAO
}

func NewStandRepository(dao StandDAO) *StandRepository {
	return &StandRepository{
		dao: dao,
	}
}
