package repository

import (
	"context"
	"fmt"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
)

var (
	ErrUserEmailExists = dao.ErrUserEmailExists
	ErrUserNotFound    = dao.ErrUserNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user dao.User) (dao.User, error)
	FindByID(ctx context.Context, id uint) (dao.User, error)
	FindByEmail(ctx context.Context, email string) (dao.User, error)
	InsertStudent(ctx context.Context, user dao.User, student dao.Student) (dao.Student, error)
	UpdateStudent(ctx context.Context, user dao.User, student dao.Student) (dao.Student, error)
	InsertParent(ctx context.Context, user dao.User, parent dao.Parent) (dao.Parent, error)
	InsertStandHolder(ctx context.Context, user dao.User, stand dao.Stand, standHolder dao.StandHolder) (dao.StandHolder, error)
	InsertOrganizer(ctx context.Context, user dao.User) (dao.Organizer, error)
	FindStudentByEmail(ctx context.Context, email string) (dao.Student, error)
	FindStudentByUserID(ctx context.Context, id uint) (dao.Student, error)
	FindParentByUserID(ctx context.Context, id uint) (dao.Parent, error)
	FindStandHolderByUserID(ctx context.Context, id uint) (dao.StandHolder, error)
	UpdateParent(ctx context.Context, user dao.User, parent dao.Parent) (dao.Parent, error)
	FindStudentOnlyByUserID(ctx context.Context, userID uint) (dao.Student, error)
	FindParentOnlyByUserID(ctx context.Context, userID uint) (dao.Parent, error)
	FindStudentsByParentID(ctx context.Context, parentID uint) ([]dao.Student, error)
}

type UserRepository struct {
	dao UserDAO
}

func NewUserRepository(dao UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	created, err := r.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.Insert -> %w", err)
	}

	return r.daoToDomain(created), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (domain.User, error) {
	found, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.FindByID -> %w", err)
	}

	return r.daoToDomain(found), nil
}
func (r *UserRepository) FindByIDWithDetails(ctx context.Context, id uint) (domain.UserWithDetails, error) {
	user, err := r.dao.FindByID(ctx, id)
	if err != nil {
		return domain.UserWithDetails{}, fmt.Errorf("r.dao.FindByID -> %w", err)
	}

	userWithDetails := domain.UserWithDetails{
		User: r.daoToDomain(user),
	}

	switch user.Role {
	case "student":
		student, err := r.dao.FindStudentByUserID(ctx, id)
		if err != nil {
			return domain.UserWithDetails{}, fmt.Errorf("r.dao.FindStudentByUserID -> %w", err)
		}
		userWithDetails.Tokens = student.Tokens
	case "parent":
		parent, err := r.dao.FindParentByUserID(ctx, id)
		if err != nil {
			return domain.UserWithDetails{}, fmt.Errorf("r.dao.FindParentByUserID -> %w", err)
		}
		userWithDetails.Tokens = parent.Tokens

		students, err := r.dao.FindStudentsByParentID(ctx, id)
		if err != nil {
			return domain.UserWithDetails{}, fmt.Errorf("r.dao.FindStudentsByParentID -> %w", err)
		}
		userWithDetails.Students = r.studentsDaoToDomain(students)
	case "stand_holder":
		standHolder, err := r.dao.FindStandHolderByUserID(ctx, id)
		if err != nil {
			return domain.UserWithDetails{}, fmt.Errorf("r.dao.FindStandHolderByUserID -> %w", err)
		}
		userWithDetails.StandID = standHolder.StandID
	}

	return userWithDetails, nil
}

func (r *UserRepository) FindStudentByUserID(ctx context.Context, id uint) (domain.Student, error) {
	found, err := r.dao.FindStudentByUserID(ctx, id)
	if err != nil {
		return domain.Student{}, fmt.Errorf("r.dao.FindStudentByID -> %w", err)
	}

	return r.studentDaoToDomain(found), nil
}

func (r *UserRepository) FindParentByUserID(ctx context.Context, id uint) (domain.Parent, error) {
	found, err := r.dao.FindParentByUserID(ctx, id)
	if err != nil {
		return domain.Parent{}, fmt.Errorf("r.dao.FindParentByID -> %w", err)
	}

	return r.parentDaoToDomain(found), nil
}

func (r *UserRepository) UpdateParent(ctx context.Context, parent domain.Parent) (domain.Parent, error) {
	parentDAO := r.parentDomainToDao(parent)

	userDao := dao.User{
		ID:       parent.User.ID,
		Email:    parent.User.Email,
		Name:     parent.User.Name,
		Role:     parent.User.Role,
		Password: parent.User.Password,
	}

	updated, err := r.dao.UpdateParent(ctx, userDao, parentDAO)
	if err != nil {
		return domain.Parent{}, fmt.Errorf("r.dao.UpdateParent -> %w", err)
	}
	return r.parentDaoToDomain(updated), nil
}

func (r *UserRepository) FindStandHolderByUserID(ctx context.Context, id uint) (domain.StandHolder, error) {
	found, err := r.dao.FindStandHolderByUserID(ctx, id)
	if err != nil {
		return domain.StandHolder{}, fmt.Errorf("r.dao.FindStandHolderByID -> %w", err)
	}

	return r.standHolderDaoToDomain(found), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	found, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("r.dao.FindByEmail -> %w", err)
	}

	return r.daoToDomain(found), nil
}

func (r *UserRepository) FindStudentByEmail(ctx context.Context, email string) (domain.Student, error) {
	// Implementation
	found, err := r.dao.FindStudentByEmail(ctx, email)
	if err != nil {
		return domain.Student{}, fmt.Errorf("r.dao.FindByEmail -> %w", err)
	}

	return r.studentDaoToDomain(found), nil
}

func (r *UserRepository) daoToDomain(u dao.User) domain.User {
	return domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (r *UserRepository) daosToDomain(us []dao.User) []domain.User {
	users := make([]domain.User, len(us))
	for i, u := range us {
		users[i] = r.daoToDomain(u)
	}

	return users
}

func (r *UserRepository) parentDomainToDao(p domain.Parent) dao.Parent {
	return dao.Parent{
		UserID: p.UserID,
		Tokens: p.Tokens,
	}
}

func (r *UserRepository) studentDaoToDomain(s dao.Student) domain.Student {
	return domain.Student{
		User:     r.daoToDomain(s.User),
		UserID:   s.UserID,
		Points:   s.Points,
		Tokens:   s.Tokens,
		ParentID: s.ParentID,
		IsActive: s.IsActive,
	}
}

func (r *UserRepository) studentsDaoToDomain(ss []dao.Student) []domain.Student {
	students := make([]domain.Student, len(ss))
	for i, s := range ss {
		students[i] = r.studentDaoToDomain(s)
	}

	return students
}

func (r *UserRepository) parentDaoToDomain(p dao.Parent) domain.Parent {
	return domain.Parent{
		User:   r.daoToDomain(p.User),
		UserID: p.UserID,
		Tokens: p.Tokens,
	}
}

func (r *UserRepository) standHolderDaoToDomain(s dao.StandHolder) domain.StandHolder {
	return domain.StandHolder{
		User:    r.daoToDomain(s.User),
		UserID:  s.UserID,
		StandID: s.StandID,
		Stand: domain.Stand{
			ID:          s.Stand.ID,
			Name:        s.Stand.Name,
			Type:        s.Stand.Type,
			Description: s.Stand.Description,
		},
	}
}

func (r *UserRepository) organizerDaoToDomain(o dao.Organizer) domain.Organizer {
	return domain.Organizer{
		User:   r.daoToDomain(o.User),
		UserID: o.UserID,
	}
}

func (r *UserRepository) organizersDaoToDomain(os []dao.Organizer) []domain.Organizer {
	organizers := make([]domain.Organizer, len(os))
	for i, o := range os {
		organizers[i] = r.organizerDaoToDomain(o)
	}

	return organizers
}

func (r *UserRepository) CreateStudent(ctx context.Context, student domain.Student) (domain.Student, error) {
	daoUser := dao.User{
		Email:    student.User.Email,
		Password: student.User.Password,
		Name:     student.User.Name,
		Role:     "student",
	}

	daoStudent := dao.Student{
		Points:   student.Points,
		Tokens:   student.Tokens,
		ParentID: student.ParentID,
		IsActive: student.IsActive,
	}

	created, err := r.dao.InsertStudent(ctx, daoUser, daoStudent)
	if err != nil {
		return domain.Student{}, fmt.Errorf("r.dao.InsertStudent -> %w", err)
	}

	return r.studentDaoToDomain(created), nil
}

func (r *UserRepository) CreateParent(ctx context.Context, parent domain.Parent) (domain.Parent, error) {
	daoUser := dao.User{
		Email:    parent.User.Email,
		Password: parent.User.Password,
		Name:     parent.User.Name,
		Role:     "parent",
	}

	daoParent := dao.Parent{
		Tokens: 0,
	}

	created, err := r.dao.InsertParent(ctx, daoUser, daoParent)
	if err != nil {
		return domain.Parent{}, fmt.Errorf("r.dao.InsertParent -> %w", err)
	}

	return r.parentDaoToDomain(created), nil
}

func (r *UserRepository) CreateStandHolder(ctx context.Context, standHolder domain.StandHolder) (domain.StandHolder, error) {
	daoUser := dao.User{
		Email:    standHolder.User.Email,
		Password: standHolder.User.Password,
		Name:     standHolder.User.Name,
		Role:     "stand_holder",
	}

	daoStand := dao.Stand{
		Name:        standHolder.Stand.Name,
		Type:        standHolder.Stand.Type,
		Description: standHolder.Stand.Description,
	}

	daoStandHolder := dao.StandHolder{}

	created, err := r.dao.InsertStandHolder(ctx, daoUser, daoStand, daoStandHolder)
	if err != nil {
		return domain.StandHolder{}, fmt.Errorf("r.dao.InsertStandHolder -> %w", err)
	}

	return r.standHolderDaoToDomain(created), nil
}

func (r *UserRepository) CreateOrganizer(ctx context.Context, organizer domain.Organizer) (domain.Organizer, error) {
	daoUser := dao.User{
		Email:    organizer.User.Email,
		Password: organizer.User.Password,
		Name:     organizer.User.Name,
		Role:     "organizer",
	}

	created, err := r.dao.InsertOrganizer(ctx, daoUser)
	if err != nil {
		return domain.Organizer{}, fmt.Errorf("r.dao.InsertOrganizer -> %w", err)
	}

	return r.organizerDaoToDomain(created), nil
}

func (r *UserRepository) UpdateStudent(ctx context.Context, student domain.Student) (domain.Student, error) {
	daoUser := dao.User{
		Email:    student.User.Email,
		Password: student.User.Password,
		Name:     student.User.Name,
	}

	daoStudent := dao.Student{
		UserID:   student.UserID,
		Points:   student.Points,
		Tokens:   student.Tokens,
		ParentID: student.ParentID,
		IsActive: student.IsActive,
	}

	updated, err := r.dao.UpdateStudent(ctx, daoUser, daoStudent)
	if err != nil {
		return domain.Student{}, fmt.Errorf("r.dao.UpdateStudent -> %w", err)
	}

	return r.studentDaoToDomain(updated), nil
}

func (r *UserRepository) UpdateUserTokens(ctx context.Context, userID uint, amount int) error {
	user, err := r.dao.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("r.userDao.FindByID -> %w", err)
	}

	if user.Role == "parent" {
		parent, err := r.dao.FindParentByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("r.userDao.FindParentByUserID -> %w", err)
		}
		parent.Tokens += amount
		_, err = r.dao.UpdateParent(ctx, user, parent)
		if err != nil {
			return fmt.Errorf("r.userDao.UpdateParent -> %w", err)
		}
	} else if user.Role == "student" {
		student, err := r.dao.FindStudentByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("r.userDao.FindStudentByUserID -> %w", err)
		}
		student.Tokens += amount
		_, err = r.dao.UpdateStudent(ctx, user, student)
		if err != nil {
			return fmt.Errorf("r.userDao.UpdateStudent -> %w", err)
		}
	} else {
		return ErrInvalidUserRole
	}

	return nil
}
