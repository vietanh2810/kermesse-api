package dao

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

var (
	ErrKermessNotFound          = errors.New("kermesse not found")
	ErrUserNotParticipant       = errors.New("user is not a participant")
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrUnauthorizedOrganizer    = errors.New("user is not an organizer of the kermesse")
	ErrInsufficientTokens       = errors.New("insufficient tokens")
	ErrNotParentOfStudent       = errors.New("user is not the parent of the student")
	ErrInvalidTransactionStatus = errors.New("invalid transaction status")
	ErrStandNotInKermesse       = errors.New("stand not in kermesse")
	ErrInsufficientStock        = errors.New("insufficient stock")
	ErrInvalidUserRole          = errors.New("invalid user role")
	ErrInvalidTransaction       = errors.New("invalid transaction")
	ErrStockNotFound            = errors.New("stock not found")
)

type Stand struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Type        string `gorm:"not null"` // "food", "drink", or "activity"
	Description string
	KermesseID  *uint    `gorm:"index"`
	Kermesse    Kermesse `gorm:"foreignKey:KermesseID"`
	Stock       []Stock  `gorm:"foreignKey:StandID"`
	TokensSpent int      `gorm:"default:0"`
	PointsGiven int      `gorm:"default:0"` // Only for activity stands
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Stock struct {
	ID        uint   `gorm:"primaryKey"`
	StandID   uint   `gorm:"not null"`
	ItemName  string `gorm:"not null"`
	Quantity  int    `gorm:"not null"`
	TokenCost int    `gorm:"not null"` // Cost in tokens
}

type Kermesse struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"not null"`
	Date         time.Time `gorm:"not null"`
	Location     string    `gorm:"not null"`
	Description  string
	Organizers   []Organizer `gorm:"many2many:organizer_kermesses;"`
	Participants []User      `gorm:"many2many:kermesse_participants;"`
	Stands       []Stand     `gorm:"foreignKey:KermesseID"`
	TokensSold   int         `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Tombola struct {
	ID         uint     `gorm:"primaryKey"`
	KermesseID uint     `gorm:"not null"`
	Kermesse   Kermesse `gorm:"foreignKey:KermessID"`
	Prizes     []Prize  `gorm:"foreignKey:TombolaID"`
	Tickets    []Ticket `gorm:"foreignKey:TombolaID"`
}

type Prize struct {
	ID        uint   `gorm:"primaryKey"`
	TombolaID uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
	Quantity  int    `gorm:"not null"`
}

type Ticket struct {
	ID        uint   `gorm:"primaryKey"`
	TombolaID uint   `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignKey:UserID"`
	Number    string `gorm:"not null"`
}

type ChatMessage struct {
	ID         uint `gorm:"primaryKey"`
	KermesseID uint `gorm:"index"`
	StandID    uint `gorm:"index"`
	SenderID   uint `gorm:"index"`
	ReceiverID uint `gorm:"index"`
	Message    string
	Timestamp  time.Time
}

type PointAttributionResult struct {
	StudentID   uint
	PointsAdded int
	TotalPoints int
}

type KermesseDao struct {
	db *gorm.DB
}

func NewKermesseDao(db *gorm.DB) *KermesseDao {
	return &KermesseDao{
		db: db,
	}
}

//func (d *KermesseDao) FindByUserID(ctx context.Context, user User) ([]Kermesse, error) {
//	var kermesses []Kermesse
//	err := d.db.Model(&Kermesse{}).
//		Joins("JOIN kermesse_participants ON kermesse_participants.kermesse_id = kermesses.id").
//		Where("kermesse_participants.user_id = ?", user.ID).
//		Find(&kermesses).Error
//	if err != nil {
//		return []Kermesse{}, err
//	}
//
//	return kermesses, nil
//}

func (d *KermesseDao) GetByID(id uint) (Kermesse, error) {
	var kermesse Kermesse
	err := d.db.First(&kermesse, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Kermesse{}, ErrKermessNotFound
		}
		return Kermesse{}, err
	}

	return kermesse, nil
}

//func (d *KermesseDao) GetStandByID(standID uint) (Stand, error) {
//	var stand Stand
//	err := d.db.Preload("Stock").First(&stand, standID).Error
//	if err != nil {
//		return Stand{}, err
//	}
//	return stand, nil
//}

func (d *KermesseDao) GetStandByID(standID uint) (Stand, error) {
	var stand Stand
	result := d.db.Preload("Stock").First(&stand, standID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Stand{}, fmt.Errorf("stand with ID %d not found", standID)
		}
		return Stand{}, fmt.Errorf("error fetching stand: %w", result.Error)
	}
	return stand, nil
}
func (d *KermesseDao) GetStockByID(ctx context.Context, stockID uint) (Stock, error) {
	var stock Stock
	err := d.db.WithContext(ctx).First(&stock, stockID).Error
	if err != nil {
		return Stock{}, err
	}
	return stock, nil
}

func (d *KermesseDao) GetStockItem(standID uint, stockID uint) (Stock, error) {
	var stock Stock
	err := d.db.Where("stand_id = ? AND id = ?", standID, stockID).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Stock{}, ErrStockNotFound
		}
		return Stock{}, err
	}
	return stock, nil
}

func (d *KermesseDao) GetAllKermesses() ([]Kermesse, error) {
	var kermesses []Kermesse
	err := d.db.
		Preload("Organizers").
		Preload("Participants").
		Preload("Stands").
		Preload("Stands.Stock").
		Find(&kermesses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kermesses: %w", err)
	}
	return kermesses, nil
}

func (d *KermesseDao) FindByUserID(user User) ([]Kermesse, error) {
	var kermesses []Kermesse
	query := d.db.Model(&Kermesse{})

	if user.Role == "organizer" {
		err := query.
			Joins("JOIN organizer_kermesses ON organizer_kermesses.kermesse_id = kermesses.id").
			Joins("JOIN organizers ON organizers.user_id = organizer_kermesses.organizer_user_id").
			Where("organizers.user_id = ?", user.ID).
			Find(&kermesses).Error
		if err != nil {
			return []Kermesse{}, err
		}
	} else {
		err := query.
			Joins("JOIN kermesse_participants ON kermesse_participants.kermesse_id = kermesses.id").
			Where("kermesse_participants.user_id = ?", user.ID).
			Find(&kermesses).Error
		if err != nil {
			return []Kermesse{}, err
		}
	}
	return kermesses, nil
}

func (d *KermesseDao) CreateKermess(ctx context.Context, kermesse Kermesse, organizerID uint) (Kermesse, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Kermesse{}, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the Kermesse
	if err := tx.Create(&kermesse).Error; err != nil {
		tx.Rollback()
		return Kermesse{}, err
	}

	// Fetch the Organizer
	var organizer Organizer
	if err := tx.First(&organizer, organizerID).Error; err != nil {
		tx.Rollback()
		return Kermesse{}, fmt.Errorf("failed to fetch organizer: %w", err)
	}

	// Associate the Organizer with the Kermesse
	if err := tx.Model(&kermesse).Association("Organizers").Append(&organizer); err != nil {
		tx.Rollback()
		return Kermesse{}, fmt.Errorf("failed to associate organizer with kermesse: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return Kermesse{}, err
	}

	return kermesse, nil
}

func (d *KermesseDao) AddParticipant(ctx context.Context, kermesseID, userID uint) error {
	var kermesse Kermesse
	if err := d.db.WithContext(ctx).First(&kermesse, kermesseID).Error; err != nil {
		return fmt.Errorf("failed to find kermesse: %w", err)
	}

	var user User
	if err := d.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if err := d.db.WithContext(ctx).Model(&kermesse).Association("Participants").Append(&user); err != nil {
		return fmt.Errorf("failed to add participant to kermesse: %w", err)
	}

	return nil
}

func (d *KermesseDao) UpdateTokenBalances(parentUserID, studentUserID uint, amount int) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Parent{}).Where("user_id = ?", parentUserID).Update("tokens", gorm.Expr("tokens - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&Student{}).Where("user_id = ?", studentUserID).Update("tokens", gorm.Expr("tokens + ?", amount)).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *KermesseDao) CreateStand(ctx context.Context, stand Stand, stock []Stock, standHolderID uint) (Stand, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return Stand{}, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&stand).Error; err != nil {
		tx.Rollback()
		return Stand{}, err
	}

	for i := range stock {
		stock[i].StandID = stand.ID
		if err := tx.Create(&stock[i]).Error; err != nil {
			tx.Rollback()
			return Stand{}, err
		}
	}

	// Update the StandHolder's StandID
	result := tx.Model(&StandHolder{}).
		Where("user_id = ?", standHolderID).
		Update("stand_id", stand.ID)

	if result.Error != nil {
		tx.Rollback()
		return Stand{}, result.Error
	}

	if result.RowsAffected == 0 {
		// If no rows were affected, it means there's no existing StandHolder
		tx.Rollback()
		return Stand{}, fmt.Errorf("standHolder not found for user ID: %d", standHolderID)
	}

	if err := tx.Commit().Error; err != nil {
		return Stand{}, err
	}

	return stand, nil
}

func (d *KermesseDao) CreateTokenTransaction(transaction TokenTransaction) (TokenTransaction, error) {
	err := d.db.Create(&transaction).Error
	if err != nil {
		return TokenTransaction{}, err
	}
	return transaction, nil
}

func (d *KermesseDao) GetTokenTransactionByID(transactionID uint) (TokenTransaction, error) {
	var transaction TokenTransaction
	err := d.db.First(&transaction, transactionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TokenTransaction{}, ErrTransactionNotFound
		}
		return TokenTransaction{}, err
	}
	return transaction, nil
}

func (d *KermesseDao) IsUserKermesseOrganizer(kermesseID uint, userID uint) (bool, error) {
	var count int64
	err := d.db.Model(&Organizer{}).
		Joins("JOIN organizer_kermesses ON organizer_kermesses.organizer_user_id = organizers.user_id").
		Where("organizer_kermesses.kermesse_id = ? AND organizers.user_id = ?", kermesseID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *KermesseDao) IncrementParentTokens(transactionFromID uint, transactionAmount int) error {
	var transactionFrom TokenTransaction
	if err := d.db.First(&transactionFrom, transactionFromID).Error; err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	var user User
	if err := d.db.First(&user, transactionFrom.FromID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	var parent Parent
	err := d.db.Where("user_id = ?", user.ID).First(&parent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("parent not found for user ID %d", user.ID)
		}
		return fmt.Errorf("failed to find parent: %w", err)
	}

	parent.Tokens += transactionAmount
	if err := d.db.Save(&parent).Error; err != nil {
		return fmt.Errorf("failed to update parent: %w", err)
	}

	return nil
}

func (d *KermesseDao) IncrementKermesseTokensSold(kermesseID uint, transactionAmount int) error {
	var kermesse Kermesse
	if err := d.db.First(&kermesse, kermesseID).Error; err != nil {
		return fmt.Errorf("failed to find kermesse: %w", err)
	}

	kermesse.TokensSold += transactionAmount
	if err := d.db.Save(&kermesse).Error; err != nil {
		return fmt.Errorf("failed to update kermesse: %w", err)
	}

	return nil
}

func (d *KermesseDao) UpdateTokenTransaction(transactionDAO TokenTransaction) (TokenTransaction, error) {
	if err := d.db.Save(&transactionDAO).Error; err != nil {
		return TokenTransaction{}, fmt.Errorf("failed to update transaction: %w", err)
	}
	return transactionDAO, nil
}

func (d *KermesseDao) UpdateStand(ctx context.Context, stand Stand) (Stand, error) {
	if err := d.db.WithContext(ctx).Save(&stand).Error; err != nil {
		return Stand{}, err
	}
	return stand, nil
}

func (d *KermesseDao) UpdateStock(ctx context.Context, stock Stock) (Stock, error) {
	if err := d.db.WithContext(ctx).Save(&stock).Error; err != nil {
		return Stock{}, err
	}
	return stock, nil
}

func (d *KermesseDao) CreateStock(ctx context.Context, stockDAO Stock) (Stock, error) {
	if err := d.db.WithContext(ctx).Create(&stockDAO).Error; err != nil {
		return Stock{}, err
	}
	return stockDAO, nil
}

func (d *KermesseDao) GetChildrenByParentID(parentID uint) ([]Student, error) {
	var students []Student
	err := d.db.Where("parent_id = ?", parentID).Find(&students).Error
	if err != nil {
		return []Student{}, err
	}
	return students, nil
}

//func (d *KermesseDao) GetChildrenTransactions(childrenIDs []uint, kermesseID uint) ([]TokenTransaction, error) {
//	var transactions []TokenTransaction
//	err := d.db.Where("from_id IN ? AND kermesse_id = ?", childrenIDs, kermesseID).Find(&transactions).Error
//	if err != nil {
//		return []TokenTransaction{}, err
//	}
//	return transactions, nil
//}

func (d *KermesseDao) GetChildrenTransactions(childrenIDs []uint) ([]TokenTransaction, error) {
	var transactions []TokenTransaction
	err := d.db.Where("from_id IN ?", childrenIDs).Find(&transactions).Error
	if err != nil {
		return []TokenTransaction{}, err
	}
	return transactions, nil
}

func (d *KermesseDao) GetStandsByKermesseID(kermesseID uint) ([]Stand, error) {
	var stands []Stand
	err := d.db.Where("kermesse_id = ?", kermesseID).
		Preload("Stock").
		Find(&stands).Error
	if err != nil {
		return []Stand{}, err
	}
	return stands, nil
}

func (d *KermesseDao) IsUserStandHolder(standID, userID uint) (bool, error) {
	var count int64
	result := d.db.Model(&StandHolder{}).
		Where("stand_id = ? AND user_id = ?", standID, userID).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check if user is stand holder: %w", result.Error)
	}
	return count > 0, nil
}

func (d *KermesseDao) SaveChatMessage(message ChatMessage) (ChatMessage, error) {
	result := d.db.Create(&message)
	if result.Error != nil {
		return ChatMessage{}, fmt.Errorf("failed to save chat message: %w", result.Error)
	}
	return message, nil
}

func (d *KermesseDao) GetChatMessages(kermesseID, standID uint, limit, offset int) ([]ChatMessage, error) {
	var messages []ChatMessage
	result := d.db.Where("kermesse_id = ? AND stand_id = ?", kermesseID, standID).
		Order("timestamp desc").
		Limit(limit).
		Offset(offset).
		Find(&messages)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", result.Error)
	}
	return messages, nil
}

func (d *KermesseDao) AttributePointsToStudent(ctx context.Context, studentID uint, points int) (PointAttributionResult, error) {
	var student Student
	err := d.db.WithContext(ctx).Where("user_id = ?", studentID).First(&student).Error
	if err != nil {
		return PointAttributionResult{}, fmt.Errorf("failed to find student: %w", err)
	}

	student.Points += points
	err = d.db.WithContext(ctx).Save(&student).Error
	if err != nil {
		return PointAttributionResult{}, fmt.Errorf("failed to update student points: %w", err)
	}

	return PointAttributionResult{
		StudentID:   studentID,
		PointsAdded: points,
		TotalPoints: student.Points,
	}, nil
}

func (d *KermesseDao) IncrementStandPointsGiven(ctx context.Context, standID uint, points int) error {
	result := d.db.WithContext(ctx).Model(&Stand{}).
		Where("id = ?", standID).
		UpdateColumn("points_given", gorm.Expr("points_given + ?", points))
	if result.Error != nil {
		return fmt.Errorf("failed to update stand points given: %w", result.Error)
	}
	return nil
}
