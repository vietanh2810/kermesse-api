package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
	"gorm.io/gorm"
)

var (
	ErrKermessNotFound          = dao.ErrKermessNotFound
	ErrUserNotParticipant       = dao.ErrUserNotParticipant
	ErrTransactionNotFound      = dao.ErrTransactionNotFound
	ErrUnauthorizedOrganizer    = dao.ErrUnauthorizedOrganizer
	ErrInvalidTransactionStatus = dao.ErrInvalidTransactionStatus
	ErrNotParentOfStudent       = dao.ErrNotParentOfStudent
	ErrInsufficientTokens       = dao.ErrInsufficientTokens
	ErrStandNotInKermesse       = dao.ErrStandNotInKermesse
	ErrInvalidUserRole          = dao.ErrInvalidUserRole
	ErrInsufficientStock        = dao.ErrInsufficientStock
	ErrInvalidTransaction       = dao.ErrInvalidTransaction
)

type KermesseDAO interface {
	FindByUserID(user dao.User) ([]dao.Kermesse, error)
	GetAllKermesses() ([]dao.Kermesse, error)
	CreateKermess(ctx context.Context, kermesse dao.Kermesse, organizerID uint) (dao.Kermesse, error)
	AddParticipant(ctx context.Context, kermesseID, userID uint) error
	GetByID(id uint) (dao.Kermesse, error)
	CreateStand(ctx context.Context, stand dao.Stand, stock []dao.Stock, standHolderID uint) (dao.Stand, error)
	CreateTokenTransaction(transaction dao.TokenTransaction) (dao.TokenTransaction, error)
	GetTokenTransactionByID(transactionID uint) (dao.TokenTransaction, error)
	IsUserKermesseOrganizer(kermesseID uint, userID uint) (bool, error)
	IncrementParentTokens(transactionFromID uint, transactionAmount int) error
	IncrementKermesseTokensSold(kermesseID uint, transactionAmount int) error
	UpdateTokenTransaction(transactionDAO dao.TokenTransaction) (dao.TokenTransaction, error)
	UpdateTokenBalances(parentUserID, studentUserID uint, amount int) error
	GetStandByID(standID uint) (dao.Stand, error)
	GetStockItem(standID uint, stockID uint) (dao.Stock, error)
	UpdateStand(ctx context.Context, stand dao.Stand) (dao.Stand, error)
	UpdateStock(ctx context.Context, stock dao.Stock) (dao.Stock, error)
	CreateStock(ctx context.Context, stockDAO dao.Stock) (dao.Stock, error)
	GetChildrenByParentID(parentID uint) ([]dao.Student, error)
	GetChildrenTransactions(childrenIDs []uint) ([]dao.TokenTransaction, error)
	GetStockByID(ctx context.Context, stockID uint) (dao.Stock, error)
	GetStandsByKermesseID(kermesseID uint) ([]dao.Stand, error)
	SaveChatMessage(message dao.ChatMessage) (dao.ChatMessage, error)
	IsUserStandHolder(standID, userID uint) (bool, error)
	GetChatMessages(kermesseID, standID uint, limit, offset int) ([]dao.ChatMessage, error)
	AttributePointsToStudent(ctx context.Context, studentID uint, points int) (dao.PointAttributionResult, error)
	IncrementStandPointsGiven(ctx context.Context, standID uint, points int) error
}

type KermesseRepository struct {
	dao   KermesseDAO
	uRepo *UserRepository
}

func NewKermesseRepository(dao KermesseDAO, uRepo *UserRepository) *KermesseRepository {
	return &KermesseRepository{
		dao:   dao,
		uRepo: uRepo,
	}
}

func (r *KermesseRepository) domainToDao(k domain.Kermesse) dao.Kermesse {
	return dao.Kermesse{
		ID:          k.ID,
		Name:        k.Name,
		Date:        k.Date,
		Location:    k.Location,
		Description: k.Description,
		TokensSold:  k.TokensSold,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}

func (r *KermesseRepository) daoToDomain(k dao.Kermesse) domain.Kermesse {
	return domain.Kermesse{
		ID:          k.ID,
		Name:        k.Name,
		Date:        k.Date,
		Location:    k.Location,
		Description: k.Description,
		TokensSold:  k.TokensSold,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}

func (r *KermesseRepository) daoToDomainKermesse(daoKermesse []dao.Kermesse) []domain.Kermesse {
	var kermesses []domain.Kermesse
	for _, k := range daoKermesse {
		kermesses = append(kermesses, domain.Kermesse{
			ID:           k.ID,
			Name:         k.Name,
			Date:         k.Date,
			Location:     k.Location,
			Description:  k.Description,
			TokensSold:   k.TokensSold,
			CreatedAt:    k.CreatedAt,
			UpdatedAt:    k.UpdatedAt,
			Stands:       r.standsDaoToDomain(k.Stands),
			Organizers:   r.uRepo.organizersDaoToDomain(k.Organizers),
			Participants: r.uRepo.daosToDomain(k.Participants),
		})
	}
	return kermesses
}

func (r *KermesseRepository) standsDomainToDao(stands []domain.Stand) []dao.Stand {
	daoStands := make([]dao.Stand, len(stands))
	for i, stand := range stands {
		daoStands[i] = r.standDomainToDao(stand)
	}
	return daoStands
}

func (r *KermesseRepository) standsDaoToDomain(stands []dao.Stand) []domain.Stand {
	domainStands := make([]domain.Stand, len(stands))
	for i, stand := range stands {
		domainStands[i] = r.standDaoToDomain(stand)
	}
	return domainStands
}

func (r *KermesseRepository) standDomainToDao(stand domain.Stand) dao.Stand {
	return dao.Stand{
		ID:          stand.ID,
		Name:        stand.Name,
		Type:        stand.Type,
		Description: stand.Description,
		KermesseID:  &stand.KermesseID,
		TokensSpent: stand.TokensSpent,
		PointsGiven: stand.PointsGiven,
		CreatedAt:   stand.CreatedAt,
		UpdatedAt:   stand.UpdatedAt,
	}
}

func (r *KermesseRepository) stockDomainToDao(stocks []domain.Stock) []dao.Stock {
	daoStocks := make([]dao.Stock, len(stocks))
	for i, stock := range stocks {
		daoStocks[i] = dao.Stock{
			ID:        stock.ID,
			StandID:   stock.StandID,
			ItemName:  stock.ItemName,
			Quantity:  stock.Quantity,
			TokenCost: stock.TokenCost,
		}
	}
	return daoStocks
}

func (r *KermesseRepository) domainToDaoStock(stock domain.Stock) dao.Stock {
	return dao.Stock{
		ID:        stock.ID,
		StandID:   stock.StandID,
		ItemName:  stock.ItemName,
		Quantity:  stock.Quantity,
		TokenCost: stock.TokenCost,
	}
}

func (r *KermesseRepository) standDaoToDomain(stand dao.Stand) domain.Stand {
	var domainStand domain.Stand

	if stand.ID != 0 {
		domainStand.ID = stand.ID
	}

	if stand.Name != "" {
		domainStand.Name = stand.Name
	}

	if stand.Type != "" {
		domainStand.Type = stand.Type
	}

	if stand.Description != "" {
		domainStand.Description = stand.Description
	}

	if stand.KermesseID != nil && *stand.KermesseID != 0 {
		domainStand.KermesseID = *stand.KermesseID
	}

	if stand.TokensSpent != 0 {
		domainStand.TokensSpent = stand.TokensSpent
	}

	if stand.PointsGiven != 0 {
		domainStand.PointsGiven = stand.PointsGiven
	}

	if !stand.CreatedAt.IsZero() {
		domainStand.CreatedAt = stand.CreatedAt
	}

	if !stand.UpdatedAt.IsZero() {
		domainStand.UpdatedAt = stand.UpdatedAt
	}

	if len(stand.Stock) > 0 {
		domainStand.Stock = make([]domain.Stock, 0, len(stand.Stock))
		for _, stock := range stand.Stock {
			domainStock := r.daoToDomainStock(stock)
			if domainStock != (domain.Stock{}) {
				domainStand.Stock = append(domainStand.Stock, domainStock)
			}
		}
		if len(domainStand.Stock) == 0 {
			domainStand.Stock = nil
		}
	}

	return domainStand
}

func (r *KermesseRepository) daoToDomainStock(stock dao.Stock) domain.Stock {
	var domainStock domain.Stock

	if stock.ID != 0 {
		domainStock.ID = stock.ID
	}

	if stock.StandID != 0 {
		domainStock.StandID = stock.StandID
	}

	if stock.ItemName != "" {
		domainStock.ItemName = stock.ItemName
	}

	if stock.Quantity != 0 {
		domainStock.Quantity = stock.Quantity
	}

	if stock.TokenCost != 0 {
		domainStock.TokenCost = stock.TokenCost
	}

	return domainStock
}

func (r *KermesseRepository) domainToDAOTokenTransaction(dt domain.TokenTransaction) dao.TokenTransaction {
	return dao.TokenTransaction{
		ID:         dt.ID,
		KermesseID: dt.KermesseID,
		FromID:     dt.FromID,
		FromType:   dt.FromType,
		ToID:       dt.ToID,
		ToType:     dt.ToType,
		Amount:     dt.Amount,
		Type:       dao.TokenTransactionType(dt.Type),
		StandID:    dt.StandID,
		Status:     dt.Status,
		CreatedAt:  dt.CreatedAt,
		UpdatedAt:  dt.UpdatedAt,
	}
}

func (r *KermesseRepository) daoToDomainTokenTransaction(dt dao.TokenTransaction) domain.TokenTransaction {
	return domain.TokenTransaction{
		ID:         dt.ID,
		KermesseID: dt.KermesseID,
		FromID:     dt.FromID,
		FromType:   dt.FromType,
		ToID:       dt.ToID,
		ToType:     dt.ToType,
		Amount:     dt.Amount,
		Type:       domain.TokenTransactionType(dt.Type),
		StandID:    dt.StandID,
		Status:     dt.Status,
		CreatedAt:  dt.CreatedAt,
		UpdatedAt:  dt.UpdatedAt,
	}
}

func (r *KermesseRepository) FindByUserID(user domain.User) ([]domain.Kermesse, error) {
	userDao := dao.User{
		ID:   user.ID,
		Role: user.Role,
	}
	kermesses, err := r.dao.FindByUserID(userDao)
	if err != nil {
		return []domain.Kermesse{}, fmt.Errorf("r.dao.FindKermessesByUserID -> %w", err)
	}

	return r.daoToDomainKermesse(kermesses), nil
}

func (r *KermesseRepository) GetAllKermesses() ([]domain.Kermesse, error) {
	kermesses, err := r.dao.GetAllKermesses()
	if err != nil {
		return nil, fmt.Errorf("r.dao.GetAllKermesses -> %w", err)
	}

	return r.daoToDomainKermesse(kermesses), nil
}

func (r *KermesseRepository) GetByID(id uint) (domain.Kermesse, error) {
	kermesse, err := r.dao.GetByID(id)
	if err != nil {
		return domain.Kermesse{}, fmt.Errorf("r.dao.GetByID -> %w", err)
	}

	return r.daoToDomain(kermesse), nil
}

func (r *KermesseRepository) CreateKermess(ctx context.Context, kermesse domain.Kermesse, organizerID uint) (domain.Kermesse, error) {
	daoKermesse := r.domainToDao(kermesse)
	created, err := r.dao.CreateKermess(ctx, daoKermesse, organizerID)
	if err != nil {
		return domain.Kermesse{}, fmt.Errorf("r.dao.Create -> %w", err)
	}

	return r.daoToDomain(created), nil
}

func (r *KermesseRepository) AddParticipant(ctx context.Context, kermesseID, userID uint) error {
	return r.dao.AddParticipant(ctx, kermesseID, userID)
}

func (r *KermesseRepository) CreateStand(ctx context.Context, stand domain.Stand, stock []domain.Stock, standHolderID uint) (domain.Stand, error) {
	daoStand := r.standDomainToDao(stand)
	daoStock := r.stockDomainToDao(stock)

	createdStand, err := r.dao.CreateStand(ctx, daoStand, daoStock, standHolderID)
	if err != nil {
		return domain.Stand{}, fmt.Errorf("r.dao.CreateStand -> %w", err)
	}

	return r.standDaoToDomain(createdStand), nil
}

func (r *KermesseRepository) UpdateStand(ctx context.Context, stand domain.Stand) (domain.Stand, error) {
	standDAO := r.standDomainToDao(stand)
	updatedStand, err := r.dao.UpdateStand(ctx, standDAO)
	if err != nil {
		return domain.Stand{}, fmt.Errorf("r.dao.UpdateStand -> %w", err)
	}
	return r.standDaoToDomain(updatedStand), nil
}

func (r *KermesseRepository) CreateTokenTransaction(transaction domain.TokenTransaction) (domain.TokenTransaction, error) {
	transactionDAO := r.domainToDAOTokenTransaction(transaction)
	createdTransactionDAO, err := r.dao.CreateTokenTransaction(transactionDAO)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("r.dao.CreateTokenTransaction -> %w", err)
	}
	return r.daoToDomainTokenTransaction(createdTransactionDAO), nil
}

func (r *KermesseRepository) GetTokenTransactionByID(id uint) (domain.TokenTransaction, error) {
	transaction, err := r.dao.GetTokenTransactionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TokenTransaction{}, ErrTransactionNotFound
		}
		return domain.TokenTransaction{}, fmt.Errorf("r.dao.GetTokenTransactionByID -> %w", err)
	}
	return r.daoToDomainTokenTransaction(transaction), nil
}

func (r *KermesseRepository) IsUserKermesseOrganizer(kermesseID, userID uint) (bool, error) {
	isOrganizer, err := r.dao.IsUserKermesseOrganizer(kermesseID, userID)
	if err != nil {
		return false, fmt.Errorf("r.dao.IsUserKermesseOrganizer -> %w", err)
	}
	return isOrganizer, nil
}

func (r *KermesseRepository) SaveChatMessage(message domain.ChatMessage) (domain.ChatMessage, error) {
	messageDAO := r.chatMessageDomainToDAO(message)
	savedMessage, err := r.dao.SaveChatMessage(messageDAO)
	if err != nil {
		return domain.ChatMessage{}, fmt.Errorf("r.dao.SaveChatMessage -> %w", err)
	}
	return r.chatMessageDAOToDomain(savedMessage), nil
}

func (r *KermesseRepository) GetChatMessages(kermesseID, standID uint, limit, offset int) ([]domain.ChatMessage, error) {
	messagesDAO, err := r.dao.GetChatMessages(kermesseID, standID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("r.dao.GetChatMessages -> %w", err)
	}

	messages := make([]domain.ChatMessage, len(messagesDAO))
	for i, msgDAO := range messagesDAO {
		messages[i] = r.chatMessageDAOToDomain(msgDAO)
	}

	return messages, nil
}

func (r *KermesseRepository) IsUserStandHolder(standID, userID uint) (bool, error) {
	return r.dao.IsUserStandHolder(standID, userID)
}

func (r *KermesseRepository) chatMessageDomainToDAO(message domain.ChatMessage) dao.ChatMessage {
	return dao.ChatMessage{
		ID:         message.ID,
		KermesseID: message.KermesseID,
		StandID:    message.StandID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Message:    message.Message,
		Timestamp:  message.Timestamp,
	}
}

func (r *KermesseRepository) chatMessageDAOToDomain(message dao.ChatMessage) domain.ChatMessage {
	return domain.ChatMessage{
		ID:         message.ID,
		KermesseID: message.KermesseID,
		StandID:    message.StandID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Message:    message.Message,
		Timestamp:  message.Timestamp,
	}
}

func (r *KermesseRepository) IncrementParentTokens(transactionFromID uint, transactionAmount int) error {
	return r.dao.IncrementParentTokens(transactionFromID, transactionAmount)
}

func (r *KermesseRepository) IncrementKermesseTokensSold(transactionFromID uint, transactionAmount int) error {
	return r.dao.IncrementKermesseTokensSold(transactionFromID, transactionAmount)
}

func (r *KermesseRepository) UpdateTokenTransaction(transaction domain.TokenTransaction) (domain.TokenTransaction, error) {
	transactionDAO := r.domainToDAOTokenTransaction(transaction)
	updatedTransactionDAO, err := r.dao.UpdateTokenTransaction(transactionDAO)
	if err != nil {
		return domain.TokenTransaction{}, fmt.Errorf("r.dao.UpdateTokenTransaction -> %w", err)
	}
	return r.daoToDomainTokenTransaction(updatedTransactionDAO), nil
}

func (r *KermesseRepository) UpdateTokenBalances(parentUserID, studentUserID uint, amount int) error {
	err := r.dao.UpdateTokenBalances(parentUserID, studentUserID, amount)
	if err != nil {
		return fmt.Errorf("r.dao.UpdateTokenBalances -> %w", err)
	}
	return nil
}

func (r *KermesseRepository) GetStandByID(standID uint) (domain.Stand, error) {
	stand, err := r.dao.GetStandByID(standID)
	if err != nil {
		return domain.Stand{}, fmt.Errorf("r.dao.GetStandByID -> %w", err)
	}
	return r.standDaoToDomain(stand), nil
}

func (r *KermesseRepository) GetStockByID(ctx context.Context, stockID uint) (domain.Stock, error) {
	stock, err := r.dao.GetStockByID(ctx, stockID)
	if err != nil {
		return domain.Stock{}, fmt.Errorf("r.dao.GetStockByID -> %w", err)
	}
	return r.daoToDomainStock(stock), nil
}

func (r *KermesseRepository) UpdateStock(ctx context.Context, updatedStock domain.Stock) (domain.Stock, error) {
	stockDAO := r.domainToDaoStock(updatedStock)
	updatedStockDAO, err := r.dao.UpdateStock(ctx, stockDAO)
	if err != nil {
		return domain.Stock{}, fmt.Errorf("r.dao.UpdateStock -> %w", err)
	}
	return r.daoToDomainStock(updatedStockDAO), nil
}

func (r *KermesseRepository) CreateStock(ctx context.Context, stock domain.Stock) (domain.Stock, error) {
	stockDAO := r.domainToDaoStock(stock)
	createdStock, err := r.dao.CreateStock(ctx, stockDAO)
	if err != nil {
		return domain.Stock{}, fmt.Errorf("r.dao.CreateStock -> %w", err)
	}
	return r.daoToDomainStock(createdStock), nil
}

func (r *KermesseRepository) GetStockItem(standID uint, stockID uint) (domain.Stock, error) {
	stock, err := r.dao.GetStockItem(standID, stockID)
	if err != nil {
		return domain.Stock{}, fmt.Errorf("r.dao.GetStockItem -> %w", err)
	}
	return r.daoToDomainStock(stock), nil
}

func (r *KermesseRepository) UpdateTransactionStatus(transactionID uint, status string) error {
	transaction, err := r.GetTokenTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("r.GetTokenTransactionByID -> %w", err)
	}

	if status != "Approved" && status != "Rejected" {
		return ErrInvalidTransactionStatus
	}

	transaction.Status = status
	_, err = r.UpdateTokenTransaction(transaction)
	if err != nil {
		return fmt.Errorf("r.UpdateTokenTransaction -> %w", err)
	}

	return nil
}

func (r *KermesseRepository) UpdateStandTokensSpent(ctx context.Context, standID uint, transactionAmount int) error {
	stand, err := r.GetStandByID(standID)
	if err != nil {
		return fmt.Errorf("r.GetStandByID -> %w", err)
	}

	stand.TokensSpent += transactionAmount
	_, err = r.UpdateStand(ctx, stand)
	if err != nil {
		return fmt.Errorf("r.UpdateStand -> %w", err)
	}

	return nil
}

func (r *KermesseRepository) UpdateStockQuantity(ctx context.Context, standID uint, stockID uint, quantityChange int) error {
	stock, err := r.GetStockItem(standID, stockID)
	if err != nil {
		return fmt.Errorf("r.GetStockItem -> %w", err)
	}

	stock.Quantity += quantityChange
	if stock.Quantity < 0 {
		return ErrInsufficientStock
	}

	daoStock := r.domainToDaoStock(stock)

	_, err = r.dao.UpdateStock(ctx, daoStock)
	if err != nil {
		return fmt.Errorf("r.UpdateStock -> %w", err)
	}

	return nil
}

func (r *KermesseRepository) GetChildrenTransactions(parentID uint) ([]domain.TokenTransaction, error) {
	// First, get all children of the parent
	childrenDAOs, err := r.dao.GetChildrenByParentID(parentID)
	if err != nil {
		return nil, fmt.Errorf("r.dao.GetChildrenByParentID -> %w", err)
	}

	if len(childrenDAOs) == 0 {
		return []domain.TokenTransaction{}, nil // Return empty slice if no children found
	}

	// Collect all children IDs
	childrenIDs := make([]uint, len(childrenDAOs))
	for i, childDAO := range childrenDAOs {
		childrenIDs[i] = childDAO.UserID
	}

	// Fetch transactions for all children
	transactionDAOs, err := r.dao.GetChildrenTransactions(childrenIDs)
	if err != nil {
		return nil, fmt.Errorf("r.dao.GetChildrenTransactions -> %w", err)
	}

	// Convert DAOs to domain models
	transactions := make([]domain.TokenTransaction, len(transactionDAOs))
	for i, transactionDAO := range transactionDAOs {
		transactions[i] = r.daoToDomainTokenTransaction(transactionDAO)
	}

	return transactions, nil
}

func (r *KermesseRepository) GetStandsByKermesseID(kermesseID uint) ([]domain.Stand, error) {
	standsDAO, err := r.dao.GetStandsByKermesseID(kermesseID)
	if err != nil {
		return nil, fmt.Errorf("r.dao.GetStandsByKermesseID -> %w", err)
	}

	stands := make([]domain.Stand, len(standsDAO))
	for i, standDAO := range standsDAO {
		stands[i] = r.standDaoToDomain(standDAO)
	}

	return stands, nil
}

func (r *KermesseRepository) AttributePointsToStudent(ctx context.Context, studentID uint, points int) (domain.PointAttributionResult, error) {
	result, err := r.dao.AttributePointsToStudent(ctx, studentID, points)
	if err != nil {
		return domain.PointAttributionResult{}, fmt.Errorf("r.dao.AttributePointsToStudent -> %w", err)
	}
	return r.daoToDomainPointAttributionResult(result), nil
}

func (r *KermesseRepository) IncrementStandPointsGiven(ctx context.Context, standID uint, points int) error {
	err := r.dao.IncrementStandPointsGiven(ctx, standID, points)
	if err != nil {
		return fmt.Errorf("r.dao.IncrementStandPointsGiven -> %w", err)
	}
	return nil
}

func (r *KermesseRepository) daoToDomainPointAttributionResult(result dao.PointAttributionResult) domain.PointAttributionResult {
	return domain.PointAttributionResult{
		StudentID:   result.StudentID,
		PointsAdded: result.PointsAdded,
		TotalPoints: result.TotalPoints,
	}
}
