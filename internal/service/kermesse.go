package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/request"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository"
)

var (
	ErrKermesseNotFound         = repository.ErrKermessNotFound
	ErrUserNotParticipant       = repository.ErrUserNotParticipant
	ErrTransactionNotFound      = repository.ErrTransactionNotFound
	ErrUnauthorizedOrganizer    = repository.ErrUnauthorizedOrganizer
	ErrInvalidTransactionStatus = repository.ErrInvalidTransactionStatus
	ErrNotParentOfStudent       = repository.ErrNotParentOfStudent
	ErrInsufficientTokens       = repository.ErrInsufficientTokens
	ErrStandNotInKermesse       = repository.ErrStandNotInKermesse
	ErrInsufficientStock        = repository.ErrInsufficientStock
	ErrInvalidUserRole          = repository.ErrInvalidUserRole
	ErrInvalidTransaction       = repository.ErrInvalidTransaction
)

type KermesseRepository interface {
	FindByUserID(user domain.User) ([]domain.Kermesse, error)
	CreateKermess(ctx context.Context, kermesse domain.Kermesse, organizerID uint) (domain.Kermesse, error)
	AddParticipant(ctx context.Context, kermesseID, userID uint) error
	GetByID(id uint) (domain.Kermesse, error)
	CreateStand(ctx context.Context, stand domain.Stand, stock []domain.Stock, standHolderID uint) (domain.Stand, error)
	CreateTokenTransaction(transaction domain.TokenTransaction) (domain.TokenTransaction, error)
	GetTokenTransactionByID(transactionID uint) (domain.TokenTransaction, error)
	IsUserKermesseOrganizer(kermesseID, userID uint) (bool, error)
	IncrementParentTokens(transactionFromID uint, transactionAmount int) error
	IncrementKermesseTokensSold(transactionFromID uint, transactionAmount int) error
	UpdateTokenTransaction(transaction domain.TokenTransaction) (domain.TokenTransaction, error)
	UpdateTokenBalances(parentUserID, studentUserID uint, amount int) error
	GetStandByID(standID uint) (domain.Stand, error)
	UpdateTransactionStatus(transactionID uint, status string) error
	UpdateStand(ctx context.Context, stand domain.Stand) (domain.Stand, error)
	UpdateStockQuantity(ctx context.Context, standID uint, itemName string, quantityChange int) error
	GetChildrenTransactions(parentID uint, kermesseID uint) ([]domain.TokenTransaction, error)
	GetStockByID(ctx context.Context, stockID uint) (domain.Stock, error)
	UpdateStock(ctx context.Context, updatedStock domain.Stock) (domain.Stock, error)
	GetStandsByKermesseID(kermesseID uint) ([]domain.Stand, error)
}

type KermesseService struct {
	repo     KermesseRepository
	userRepo UserRepository
}

func NewKermesseService(repo KermesseRepository, userRepo UserRepository) *KermesseService {
	return &KermesseService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *KermesseService) IsParticipating(kermessID uint, userID uint) (bool, error) {
	kermesses, err := s.repo.FindByUserID(domain.User{ID: userID})
	if err != nil {
		return false, fmt.Errorf("s.repo.FindKermessesByUserID -> %w", err)
	}

	for _, k := range kermesses {
		if k.ID == kermessID {
			return true, nil
		}
	}

	return false, nil
}

func (s *KermesseService) IsStandHolder(userID, standID uint) (bool, error) {
	stand, err := s.repo.GetStandByID(standID)
	if err != nil {
		return false, fmt.Errorf("s.repo.GetStandByID -> %w", err)
	}

	standHolder, err := s.userRepo.FindStandHolderByUserID(context.Background(), userID)
	if err != nil {
		return false, fmt.Errorf("s.userRepo.FindStandHolderByUserID -> %w", err)
	}

	if standHolder.StandID != stand.ID {
		return false, nil
	}

	return true, nil
}

func (s *KermesseService) GetStandsByKermesseID(kermesseID uint) ([]domain.Stand, error) {
	kermesse, err := s.repo.GetByID(kermesseID)
	if err != nil {
		return []domain.Stand{}, fmt.Errorf("s.repo.GetByID -> %w", err)
	}

	stands, err := s.repo.GetStandsByKermesseID(kermesse.ID)
	if err != nil {
		return []domain.Stand{}, fmt.Errorf("s.repo.GetStandsByKermesseID -> %w", err)
	}

	return stands, nil
}

func (s *KermesseService) IsKermesseOrganizer(kermesseID, userID uint) (bool, error) {
	return s.repo.IsUserKermesseOrganizer(kermesseID, userID)
}

func (s *KermesseService) GetKermesses(user domain.User) ([]domain.Kermesse, error) {
	kermesses, err := s.repo.FindByUserID(user)
	if err != nil {
		return []domain.Kermesse{}, fmt.Errorf("s.repo.FindKermessesByUserID -> %w", err)
	}

	return kermesses, nil
}

func (s *KermesseService) GetTokenTransactionByID(transactionID uint) (domain.TokenTransaction, error) {
	transaction, err := s.repo.GetTokenTransactionByID(transactionID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetTokenTransactionByID -> %w", err)
	}

	return transaction, nil
}

func (s *KermesseService) CreateKermesse(ctx context.Context, kermesse domain.Kermesse, organizerID uint) (domain.Kermesse, error) {
	createdKermesse, err := s.repo.CreateKermess(ctx, kermesse, organizerID)
	if err != nil {
		return domain.Kermesse{}, fmt.Errorf("s.repo.Create -> %w", err)
	}

	return createdKermesse, nil
}

func (s *KermesseService) AddParticipantToKermesse(ctx context.Context, kermesseID, userID uint) error {
	return s.repo.AddParticipant(ctx, kermesseID, userID)
}

func (s *KermesseService) CreateStand(ctx context.Context, stand domain.Stand, stock []domain.Stock, standHolderID uint) (domain.Stand, error) {
	// Check if the kermesse exists
	if _, err := s.repo.GetByID(stand.KermesseID); err != nil {
		return domain.Stand{}, fmt.Errorf("kermesse not found: %w", err)
	}

	// Check if the stand holder exists and is a valid user
	if _, err := s.userRepo.FindByID(ctx, standHolderID); err != nil {
		return domain.Stand{}, fmt.Errorf("invalid stand holder: %w", err)
	}

	createdStand, err := s.repo.CreateStand(ctx, stand, stock, standHolderID)
	if err != nil {
		return domain.Stand{}, fmt.Errorf("s.repo.CreateStand -> %w", err)
	}

	return createdStand, nil
}

func (s *KermesseService) CreateTokenTransaction(transaction domain.TokenTransaction, user domain.User) (domain.TokenTransaction, error) {
	// Check if the kermesse exists and if the parent is participating
	isParticipating, err := s.IsParticipating(transaction.KermesseID, user.ID)
	if err != nil {
		if errors.Is(err, ErrKermesseNotFound) {
			return domain.TokenTransaction{}, ErrKermesseNotFound
		}
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.IsParentParticipating -> %w", err)
	}
	if !isParticipating {
		return domain.TokenTransaction{}, ErrUserNotParticipant
	}

	// Create the transaction
	createdTransaction, err := s.repo.CreateTokenTransaction(transaction)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.CreateTokenTransaction -> %w", err)
	}

	return createdTransaction, nil
}

func (s *KermesseService) ValidateTokenTransaction(transactionID uint, user domain.User) (domain.TokenTransaction, error) {
	transaction, err := s.repo.GetTokenTransactionByID(transactionID)
	if err != nil {
		if errors.Is(err, repository.ErrTransactionNotFound) {
			return domain.TokenTransaction{}, ErrTransactionNotFound
		}
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetTokenTransactionByID -> %w", err)
	}

	// Check if the user is an organizer of this kermesse
	isOrganizer, err := s.repo.IsUserKermesseOrganizer(transaction.KermesseID, user.ID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.IsUserKermesseOrganizer -> %w", err)
	}
	if !isOrganizer {
		return domain.TokenTransaction{}, ErrUnauthorizedOrganizer
	}

	// Check if the transaction is in a valid state for validation
	if transaction.Status != "Pending" || transaction.Type != domain.TokenPurchase {
		return domain.TokenTransaction{}, ErrInvalidTransactionStatus
	}

	// Update transaction status
	transaction.Status = "Validated"

	// Update parent's token balance
	err = s.repo.IncrementParentTokens(transaction.FromID, transaction.Amount)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.IncrementParentTokens -> %w", err)
	}

	// Update kermesse's tokens sold
	err = s.repo.IncrementKermesseTokensSold(transaction.KermesseID, transaction.Amount)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.IncrementKermesseTokensSold -> %w", err)
	}

	// Save updated transaction
	updatedTransaction, err := s.repo.UpdateTokenTransaction(transaction)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.UpdateTokenTransaction -> %w", err)
	}

	return updatedTransaction, nil
}

func (s *KermesseService) CreateParentToChildTokenTransaction(ctx context.Context, transaction domain.TokenTransaction, user domain.User) (domain.TokenTransaction, error) {
	// Check if the parent has enough tokens
	parent, err := s.userRepo.FindParentByUserID(ctx, user.ID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetParentByUserID -> %w", err)
	}
	if parent.Tokens < transaction.Amount {
		return domain.TokenTransaction{}, ErrInsufficientTokens
	}

	// Check if the parent is the parent of the student
	student, err := s.userRepo.FindStudentByUserID(ctx, transaction.ToID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetStudentByUserID -> %w", err)
	}
	if student.ParentID != user.ID {
		return domain.TokenTransaction{}, ErrNotParentOfStudent
	}

	// Check if both parent and student are participating in the kermesse
	isParentParticipating, err := s.IsParticipating(transaction.KermesseID, user.ID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.IsParentParticipating -> %w", err)
	}
	isStudentParticipating, err := s.IsParticipating(transaction.KermesseID, student.UserID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.IsStudentParticipating -> %w", err)
	}
	if !isParentParticipating || !isStudentParticipating {
		return domain.TokenTransaction{}, ErrUserNotParticipant
	}

	// Create the transaction
	createdTransaction, err := s.repo.CreateTokenTransaction(transaction)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.CreateTokenTransaction -> %w", err)
	}

	// Update parent's and student's token balances
	err = s.repo.UpdateTokenBalances(parent.UserID, student.UserID, transaction.Amount)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.UpdateTokenBalances -> %w", err)
	}

	return createdTransaction, nil
}

func (s *KermesseService) GetStandByID(standID uint) (domain.Stand, error) {
	stand, err := s.repo.GetStandByID(standID)
	if err != nil {
		return domain.Stand{}, fmt.Errorf("s.repo.GetStandByID -> %w", err)
	}

	return stand, nil
}

func (s *KermesseService) GetStockItem(standID uint, itemName string) (domain.Stock, error) {
	stand, err := s.GetStandByID(standID)
	if err != nil {
		return domain.Stock{}, fmt.Errorf("s.GetStandByID -> %w", err)
	}

	for _, stock := range stand.Stock {
		if stock.ItemName == itemName {
			return stock, nil
		}
	}

	return domain.Stock{}, fmt.Errorf("stock item not found")
}

func (s *KermesseService) PerformPurchase(ctx context.Context, userID, kermesseID, standID uint, itemName string, quantity int, totalCost int) (domain.TokenTransaction, error) {
	// Check if the stand exists and belongs to the kermesse
	stand, err := s.repo.GetStandByID(standID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetStandByID -> %w", err)
	}
	if stand.KermesseID != kermesseID {
		return domain.TokenTransaction{}, ErrStandNotInKermesse
	}

	// Get the stock item
	stockItem, err := s.GetStockItem(standID, itemName)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.GetStockItem -> %w", err)
	}

	// Check if there's enough stock (skip for activity stands)
	if stand.Type != "activity" {
		if stockItem.Quantity < quantity {
			return domain.TokenTransaction{}, ErrInsufficientStock
		}
	}

	// Check if user has enough tokens
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.userRepo.FindUserByID -> %w", err)
	}

	var userTokens int
	var fromType string
	switch user.Role {
	case "student":
		student, err := s.userRepo.FindStudentByUserID(ctx, userID)
		if err != nil {
			return domain.TokenTransaction{}, fmt.Errorf("s.userRepo.FindStudentByUserID -> %w", err)
		}
		userTokens = student.Tokens
		fromType = "Student"
	case "parent":
		parent, err := s.userRepo.FindParentByUserID(ctx, userID)
		if err != nil {
			return domain.TokenTransaction{}, fmt.Errorf("s.userRepo.FindParentByUserID -> %w", err)
		}
		userTokens = parent.Tokens
		fromType = "Parent"
	default:
		return domain.TokenTransaction{}, ErrInvalidUserRole
	}

	if userTokens < totalCost {
		return domain.TokenTransaction{}, ErrInsufficientTokens
	}

	// Create pending TokenTransaction
	transaction := domain.TokenTransaction{
		KermesseID: kermesseID,
		FromID:     userID,
		FromType:   fromType,
		ToID:       standID,
		ToType:     "Stand",
		Amount:     totalCost,
		Type:       domain.TokenSpend,
		StandID:    &standID,
		Status:     "Pending",
	}

	if !transaction.IsValid() {
		return domain.TokenTransaction{}, ErrInvalidTransaction
	}

	// Create the pending transaction
	createdTransaction, err := s.repo.CreateTokenTransaction(transaction)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("s.repo.CreateTokenTransaction -> %w", err)
	}

	return createdTransaction, nil
}

func (s *KermesseService) IsStandHolderAssociatedWithStand(ctx context.Context, standHolderID, standID uint) (bool, error) {
	stand, err := s.GetStandByID(standID)
	if err != nil {
		return false, fmt.Errorf("s.GetStandByID -> %w", err)
	}

	standHolder, err := s.userRepo.FindStandHolderByUserID(ctx, standHolderID)
	if err != nil {
		return false, fmt.Errorf("s.userRepo.FindStandHolderByUserID -> %w", err)
	}

	if standHolder.StandID != stand.ID {
		return false, nil
	}

	return true, nil
}

func (s *KermesseService) ApproveTransaction(ctx context.Context, transactionID uint, standholderID uint, itemName string, quantity int) error {
	// Fetch the transaction
	transaction, err := s.repo.GetTokenTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("s.repo.GetTokenTransactionByID -> %w", err)
	}

	// Verify that the standholder is associated with the stand
	isAssociated, err := s.IsStandHolderAssociatedWithStand(ctx, standholderID, *transaction.StandID)
	if err != nil {
		return fmt.Errorf("s.IsStandholderAssociatedWithStand -> %w", err)
	}
	if !isAssociated {
		return ErrUnauthorizedOrganizer
	}

	// Update the transaction status
	if err := s.repo.UpdateTransactionStatus(transactionID, "Approved"); err != nil {
		return fmt.Errorf("s.repo.UpdateTransactionStatus -> %w", err)
	}

	// Deduct tokens from the user
	if err := s.userRepo.UpdateUserTokens(ctx, transaction.FromID, -transaction.Amount); err != nil {
		return fmt.Errorf("s.userRepo.UpdateUserTokens -> %w", err)
	}

	stand, err := s.repo.GetStandByID(*transaction.StandID)
	if err != nil {
		return fmt.Errorf("s.repo.GetStand -> %w", err)
	}

	stand.TokensSpent += transaction.Amount

	if _, err := s.repo.UpdateStand(ctx, stand); err != nil {
		return fmt.Errorf("s.repo.UpdateStandTokensSpent -> %w", err)
	}

	if stand.Type != "activity" {
		if err := s.repo.UpdateStockQuantity(ctx, *transaction.StandID, itemName, -quantity); err != nil {
			return fmt.Errorf("s.repo.UpdateStockQuantity -> %w", err)
		}
	}

	return nil
}

func (s *KermesseService) RejectTransaction(ctx context.Context, transactionID uint, standholderID uint) error {
	transaction, err := s.repo.GetTokenTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("s.repo.GetTokenTransactionByID -> %w", err)
	}

	isAssociated, err := s.IsStandHolderAssociatedWithStand(ctx, standholderID, *transaction.StandID)
	if err != nil {
		return fmt.Errorf("s.IsStandholderAssociatedWithStand -> %w", err)
	}
	if !isAssociated {
		return ErrUnauthorizedOrganizer
	}

	if err := s.repo.UpdateTransactionStatus(transactionID, "Rejected"); err != nil {
		return fmt.Errorf("s.repo.UpdateTransactionStatus -> %w", err)
	}

	return nil
}

func (s *KermesseService) GetChildrenTransactions(ctx context.Context, userID uint, kermesseID uint) ([]domain.TokenTransaction, error) {
	// Check if the user is a parent
	parent, err := s.userRepo.FindParentByUserID(ctx, userID)
	if err != nil {
		return []domain.TokenTransaction{}, fmt.Errorf("s.userRepo.FindParentByUserID -> %w", err)
	}

	// Check if the parent is participating in the kermesse
	isParticipating, err := s.IsParticipating(kermesseID, parent.UserID)
	if err != nil {
		return []domain.TokenTransaction{}, fmt.Errorf("s.IsParentParticipating -> %w", err)
	}
	if !isParticipating {
		return []domain.TokenTransaction{}, ErrUserNotParticipant
	}

	// Fetch the transactions
	transactions, err := s.repo.GetChildrenTransactions(parent.UserID, kermesseID)
	if err != nil {
		return []domain.TokenTransaction{}, fmt.Errorf("s.repo.GetChildrenTransactions -> %w", err)
	}

	return transactions, nil
}

func (s *KermesseService) UpdateStock(ctx context.Context, req request.StockUpdateRequest, userID uint, standID uint) error {
	stand, err := s.GetStandByID(standID)
	if err != nil {
		return fmt.Errorf("s.GetStandByID -> %w", err)
	}

	standHolder, err := s.userRepo.FindStandHolderByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("s.userRepo.FindStandHolderByUserID -> %w", err)
	}

	if standHolder.StandID != stand.ID {
		return ErrUnauthorizedOrganizer
	}

	// Fetch the existing stock item
	existingStock, err := s.repo.GetStockByID(ctx, req.StockID)
	if err != nil {
		return fmt.Errorf("s.repo.GetStockByID -> %w", err)
	}

	// Check if the stock item belongs to the given stand
	if existingStock.StandID != standID {
		return ErrUnauthorizedOrganizer
	}

	// Update the stock item
	updatedStock := domain.Stock{
		ID:        req.StockID,
		StandID:   standID,
		ItemName:  req.ItemName,
		Quantity:  req.Quantity,
		TokenCost: req.TokenCost,
	}

	if _, err := s.repo.UpdateStock(ctx, updatedStock); err != nil {
		return fmt.Errorf("s.repo.UpdateStock -> %w", err)
	}

	return nil
}
