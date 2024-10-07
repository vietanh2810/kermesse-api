package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/request"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/service"
	"net/http"
	"strconv"
	"time"
)

type KermesseService interface {
	IsParticipating(kermessID uint, userID uint) (bool, error)
	GetKermesses(user domain.User) ([]domain.Kermesse, error)
	CreateKermesse(ctx context.Context, kermesse domain.Kermesse, organizerID uint) (domain.Kermesse, error)
	AddParticipantToKermesse(ctx context.Context, kermesseID, userID uint) error
	CreateStand(ctx context.Context, stand domain.Stand, stock []domain.Stock, standHolderID uint) (domain.Stand, error)
	CreateTokenTransaction(transaction domain.TokenTransaction, user domain.User) (domain.TokenTransaction, error)
	ValidateTokenTransaction(transactionID uint, user domain.User) (domain.TokenTransaction, error)
	CreateParentToChildTokenTransaction(ctx context.Context, transaction domain.TokenTransaction, user domain.User) (domain.TokenTransaction, error)
	GetStandByID(standID uint) (domain.Stand, error)
	PerformPurchase(ctx context.Context, userID, kermesseID, standID uint, itemName string, quantity int, totalCost int) (domain.TokenTransaction, error)
	GetStockItem(standID uint, itemName string) (domain.Stock, error)
	GetTokenTransactionByID(transactionID uint) (domain.TokenTransaction, error)
	IsStandHolderAssociatedWithStand(ctx context.Context, standHolderID, standID uint) (bool, error)
	ApproveTransaction(ctx context.Context, transactionID uint, standHolderID uint, itemName string, quantity int) error
	RejectTransaction(ctx context.Context, transactionID uint, standHolderID uint) error
	GetChildrenTransactions(ctx context.Context, userID uint, kermesseID uint) ([]domain.TokenTransaction, error)
	UpdateStock(ctx context.Context, req request.StockUpdateRequest, userID uint, standID uint) error
	IsKermesseOrganizer(kermesseID, userID uint) (bool, error)
	GetStandsByKermesseID(kermesseID uint) ([]domain.Stand, error)
	IsStandHolder(userID, standID uint) (bool, error)
}

type KermesseHandler struct {
	svc  KermesseService
	uSvc UserService
}

func NewKermesseHandler(svc KermesseService, uSvc UserService) *KermesseHandler {
	return &KermesseHandler{
		svc:  svc,
		uSvc: uSvc,
	}
}

// HandleGetKermesses godoc
// @Summary      Get kermesses for user
// @Description  Retrieves all kermesses associated with the authenticated user
// @Tags         kermesses
// @Produce      json
// @Success      200  {array}   domain.Kermesse
// @Failure      401  {object}  response.Err
// @Failure      404  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses [get]
// @Security BearerAuth
func (h *KermesseHandler) HandleGetKermesses(ctx *gin.Context) {
	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	kermesses, err := h.svc.GetKermesses(user)
	if err != nil {
		if errors.Is(err, service.ErrKermesseNotFound) {
			response.RenderErr(ctx, response.ErrNotFound("kermesse", "userID", user.ID))
			return
		}

		err = fmt.Errorf("HandleGetKermesse -> h.svc.GetKermesses -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusOK, kermesses)
}

// HandleCreateKermesse godoc
// @Summary      Create a new kermesse
// @Description  Creates a new kermesse event. Only users with the "organizer" role can create kermesses.
// @Tags         kermesses
// @Accept       json
// @Produce      json
// @Param        input  body      request.CreateKermesseRequest  true  "Kermesse details"
// @Success      201    {object}  domain.Kermesse
// @Failure      400    {object}  response.Err
// @Failure      401    {object}  response.Err
// @Failure      403    {object}  response.Err
// @Failure      500    {object}  response.Err
// @Router       /kermesses [post]
// @Security BearerAuth
func (h *KermesseHandler) HandleCreateKermesse(ctx *gin.Context) {
	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	// Check if the user is an organizer
	if user.Role != "organizer" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not an organizer", user.ID)))
		return
	}

	var input request.CreateKermesseRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	if err := input.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	parsedDate, err := time.Parse("02/01/2006", input.Date)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid date format: %v", err)))
		return
	}

	kermesse := domain.Kermesse{
		Name:        input.Name,
		Date:        parsedDate,
		Location:    input.Location,
		Description: input.Description,
	}

	createdKermesse, err := h.svc.CreateKermesse(ctx.Request.Context(), kermesse, user.ID)
	if err != nil {
		err = fmt.Errorf("HandleCreateKermesse -> h.svc.CreateKermesse -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusCreated, createdKermesse)
}

// HandleKermesseParticipation godoc
// @Summary      Participate in a kermesse
// @Description  Adds the authenticated user as a participant to the specified kermesse
// @Tags         kermesses
// @Produce      json
// @Param        kermesseID  path      int  true  "Kermesse ID"
// @Success      200
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/participate [get]
// @Security     BearerAuth
func (h *KermesseHandler) HandleKermesseParticipation(ctx *gin.Context) {
	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 64)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID: %w", err)))
		return
	}

	err = h.svc.AddParticipantToKermesse(ctx.Request.Context(), uint(kermesseID), user.ID)
	if err != nil {
		err = fmt.Errorf("HandleKermesseParticipation -> h.svc.AddParticipantToKermesse -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully participated in the kermesse"})
}

// HandleGetStands godoc
// @Summary      Get stands for a kermesse
// @Description  Retrieves all stands for a specific kermesse. The user must be a participant, organizer, or stand holder associated with the kermesse to access this information.
// @Tags         kermesses,stands
// @Produce      json
// @Param        kermesseID  path      int  true  "Kermesse ID"
// @Success      200  {array}   domain.Stand
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      403  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/stand [get]
// @Security BearerAuth
func (h *KermesseHandler) HandleGetStands(ctx *gin.Context) {
	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 64)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID: %w", err)))
		return
	}

	// Check if the user is a participant or organizer of the kermesse
	isParticipant, err := h.svc.IsParticipating(uint(kermesseID), user.ID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check user participation: %w", err)))
		return
	}

	isOrganizer, err := h.svc.IsKermesseOrganizer(uint(kermesseID), user.ID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check user organizer status: %w", err)))
		return
	}

	if !isParticipant && !isOrganizer {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to view stands for this kermesse", user.ID)))
		return
	}

	stands, err := h.svc.GetStandsByKermesseID(uint(kermesseID))
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to get stands: %w", err)))
		return
	}

	var result []map[string]interface{}

	for _, stand := range stands {
		standInfo := map[string]interface{}{
			"id":          stand.ID,
			"name":        stand.Name,
			"type":        stand.Type,
			"description": stand.Description,
		}

		if isOrganizer {
			// Organizers can see all details
			standInfo["tokens_spent"] = stand.TokensSpent
			standInfo["points_given"] = stand.PointsGiven
			standInfo["stock"] = stand.Stock
		} else if user.Role == "stand_holder" {
			isStandHolder, err := h.svc.IsStandHolderAssociatedWithStand(ctx, user.ID, stand.ID)
			if err != nil {
				response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check stand holder association: %w", err)))
				return
			}
			if isStandHolder {
				standInfo["tokens_spent"] = stand.TokensSpent
				standInfo["points_given"] = stand.PointsGiven
				standInfo["stock"] = stand.Stock
			}
		}

		result = append(result, standInfo)
	}

	ctx.JSON(http.StatusOK, result)
}

// HandleCreateStand godoc
// @Summary      Create a new stand for a kermesse
// @Description  Creates a new stand for a specific kermesse. Only organizers, admins, or stand holders can perform this action.
// @Tags         kermesses,stands
// @Accept       json
// @Produce      json
// @Param        kermesseID  path      int                     true  "Kermesse ID"
// @Param        stand       body      request.CreateStandRequest  true  "Stand details"
// @Success      201  {object}  domain.Stand
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      403  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/stand [post]
// @Security     BearerAuth
func (h *KermesseHandler) HandleCreateStand(ctx *gin.Context) {
	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	//// Check if the user is an organizer or admin
	//if user.Role != "organizer" && user.Role != "admin" && user.Role != "stand_holder" {
	//	response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to create stands", user.ID)))
	//	return
	//}

	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 64)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID: %w", err)))
		return
	}

	if user.Role == "stand_holder" {
		isParticipant, err := h.svc.IsParticipating(uint(kermesseID), user.ID)
		if err != nil {
			response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check user participation: %w", err)))
			return
		}
		if !isParticipant {
			response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not participating in kermesse %v", user.ID, kermesseID)))
			return
		}
	} else {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to create stands", user.ID)))
		return
	}

	var req request.CreateStandRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	if err := req.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	stand := domain.Stand{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		KermesseID:  uint(kermesseID),
	}

	stock := make([]domain.Stock, len(req.Stock))
	for i, s := range req.Stock {
		stock[i] = domain.Stock{
			ItemName:  s.ItemName,
			Quantity:  s.Quantity,
			TokenCost: s.TokenCost,
		}
	}

	createdStand, err := h.svc.CreateStand(ctx.Request.Context(), stand, stock, user.ID)
	if err != nil {
		err = fmt.Errorf("HandleCreateStand -> h.svc.CreateStand -> %w", err)
		response.RenderErr(ctx, response.ErrInternalServerError(err))
		return
	}

	ctx.JSON(http.StatusCreated, createdStand)
}

// HandleTokenPurchase godoc
// @Summary      Purchase tokens for a kermesse
// @Description  Allows a parent to purchase tokens for a specific kermesse. Only parents who are participating in the kermesse can perform this action.
// @Tags         kermesses,tokens
// @Accept       json
// @Produce      json
// @Param        kermesseID  path      int                          true  "Kermesse ID"
// @Param        purchase    body      request.TokenPurchaseRequest true  "Token purchase details"
// @Success      201  {object}  domain.TokenTransaction
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      403  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/token/purchase [post]
// @Security     BearerAuth
func (h *KermesseHandler) HandleTokenPurchase(ctx *gin.Context) {
	// Get kermesseID from URL params
	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid kermesse ID"})
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	if user.Role != "parent" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to buy token", user.ID)))
		return
	}

	// Validate that the user is a parent and has participated in the kermesse
	isParticipating, err := h.svc.IsParticipating(uint(kermesseID), user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify participation"})
		return
	}
	if !isParticipating {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Parent is not participating in this kermesse"})
		return
	}

	// Parse purchaseRequest body
	var purchaseRequest request.TokenPurchaseRequest
	if err := ctx.ShouldBindJSON(&purchaseRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := purchaseRequest.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create token transaction
	transaction := domain.TokenTransaction{
		KermesseID: uint(kermesseID),
		FromID:     user.ID,
		FromType:   "parent",
		ToID:       uint(kermesseID),
		ToType:     "kermess",
		Amount:     purchaseRequest.Amount,
		Type:       domain.TokenPurchase,
		Status:     "Pending",
	}

	// Submit token purchase purchaseRequest
	createdTransaction, err := h.svc.CreateTokenTransaction(transaction, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit token purchase purchaseRequest"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":     "Token purchase purchaseRequest submitted successfully",
		"transaction": createdTransaction,
	})
}

// HandleValidateTokenTransaction godoc
// @Summary      Validate a token transaction
// @Description  Validates a token transaction for a specific kermesse. Only organizers can perform this action.
// @Tags         kermesses,transactions
// @Produce      json
// @Param        kermesseID      path  int  true  "Kermesse ID"
// @Param        transactionID   path  int  true  "Transaction ID"
// @Success      200
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      403  {object}  response.Err
// @Failure      404  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/transaction/{transactionID}/validate [get]
// @Security BearerAuth
func (h *KermesseHandler) HandleValidateTokenTransaction(ctx *gin.Context) {
	// Get transactionID from URL params
	transactionID, err := strconv.ParseUint(ctx.Param("transactionID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid transaction ID: %w", err)))
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	if user.Role != "organizer" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to validate token transactions", user.ID)))
		return
	}

	validatedTransaction, err := h.svc.ValidateTokenTransaction(uint(transactionID), user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTransactionNotFound):
			response.RenderErr(ctx, response.ErrNotFound("transaction", "ID", transactionID))
		case errors.Is(err, service.ErrUnauthorizedOrganizer):
			response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not an organizer of this kermesse", user.ID)))
		case errors.Is(err, service.ErrInvalidTransactionStatus):
			response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("transaction is not in a valid state for validation")))
		default:
			err = fmt.Errorf("HandleValidateTokenTransaction -> h.svc.ValidateTokenTransaction -> %w", err)
			response.RenderErr(ctx, response.ErrInternalServerError(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Token transaction validated successfully",
		"transaction": validatedTransaction,
	})
}

// HandleParentSendTokensToChild godoc
// @Summary Send tokens from parent to child
// @Description Allows a parent to send tokens to their child
// @Tags kermesses
// @Accept json
// @Produce json
// @Param kermesseID path int true "Kermesse ID"
// @Param sendTokensRequest body request.SendTokensRequest true "Send tokens request"
// @Success 201
// @Failure 400 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/token/transferToChild [post]
func (h *KermesseHandler) HandleParentSendTokensToChild(ctx *gin.Context) {
	// Get kermesseID from URL params
	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID")))

		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	if user.Role != "parent" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to send tokens", user.ID)))
		return
	}

	//// Parse request body
	//var sendTokensRequest struct {
	//	StudentID uint `json:"student_id" binding:"required"`
	//	Amount    int  `json:"amount" binding:"required,min=1"`
	//}
	//if err := ctx.ShouldBindJSON(&sendTokensRequest); err != nil {
	//	response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request body: %w", err)))
	//	return
	//}

	var sendTokensRequest request.SendTokensRequest
	if err := ctx.ShouldBindJSON(&sendTokensRequest); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	if err := sendTokensRequest.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	// Create token transaction
	transaction := domain.TokenTransaction{
		KermesseID: uint(kermesseID),
		FromID:     user.ID,
		FromType:   "parent",
		ToID:       sendTokensRequest.StudentID,
		ToType:     "student",
		Amount:     sendTokensRequest.Amount,
		Type:       domain.TokenDistribution,
		Status:     "Pending",
	}

	// Submit token transfer request
	createdTransaction, err := h.svc.CreateParentToChildTokenTransaction(ctx, transaction, user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientTokens):
			response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("insufficient tokens")))
		case errors.Is(err, service.ErrNotParentOfStudent):
			response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not the parent of student %v", user.ID, sendTokensRequest.StudentID)))
		case errors.Is(err, service.ErrUserNotParticipant):
			response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not participating in kermesse %v", user.ID, kermesseID)))
		default:
			response.RenderErr(ctx, response.ErrInternalServerError(err))
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":     "Token transfer request submitted successfully",
		"transaction": createdTransaction,
	})
}

// HandleStandPurchase godoc
// @Summary Make a purchase from a stand
// @Description Allows a user to make a purchase from a stand in a kermesse
// @Tags kermesses
// @Accept json
// @Produce json
// @Param kermesseID path int true "Kermesse ID"
// @Param standID path int true "Stand ID"
// @Param purchaseRequest body request.StandPurchaseRequest true "Purchase request"
// @Success 200
// @Failure 400 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 404 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/stand/{standID}/purchase [post]
func (h *KermesseHandler) HandleStandPurchase(ctx *gin.Context) {
	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID")))
		return
	}

	standID, err := strconv.ParseUint(ctx.Param("standID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid stand ID")))
		return
	}

	var purchaseRequest request.StandPurchaseRequest
	if err := ctx.ShouldBindJSON(&purchaseRequest); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	if err := purchaseRequest.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	// Check if user is a participant of the kermesse
	isParticipant, err := h.svc.IsParticipating(user.ID, uint(kermesseID))
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check user participation: %w", err)))
		return
	}
	if !isParticipant {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user is not a participant of this kermesse")))
		return
	}

	// Get stand details
	stand, err := h.svc.GetStandByID(uint(standID))
	if err != nil {
		response.RenderErr(ctx, response.ErrNotFound("stand", "ID", standID))
		return
	}

	// Get stock item
	stockItem, err := h.svc.GetStockItem(uint(standID), purchaseRequest.ItemName)
	if err != nil {
		response.RenderErr(ctx, response.ErrNotFound("item", "name", purchaseRequest.ItemName))
		return
	}

	// Check stock availability (skip for activity stands)
	if stand.Type != "activity" && stockItem.Quantity < purchaseRequest.Quantity {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("not enough stock available")))
		return
	}

	// Calculate total cost
	totalCost := stockItem.TokenCost * purchaseRequest.Quantity

	userTokens, err := h.uSvc.GetUserTokens(ctx, user.ID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to get user tokens: %w", err)))
		return
	}
	if userTokens < totalCost {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("not enough tokens for this purchase")))
		return
	}

	// Perform the purchase
	purchase, err := h.svc.PerformPurchase(ctx, user.ID, uint(kermesseID), uint(standID), purchaseRequest.ItemName, purchaseRequest.Quantity, totalCost)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to perform purchase: %w", err)))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "Purchase successful",
		"purchase":        purchase,
		"remainingTokens": userTokens - totalCost,
	})
}

// HandleValidatePurchase godoc
// @Summary Validate a purchase transaction
// @Description Allows a stand holder to validate a purchase transaction
// @Tags kermesses
// @Accept json
// @Produce json
// @Param kermesseID path int true "Kermesse ID"
// @Param transactionID path int true "Transaction ID"
// @Param approvalRequest body request.StandTransactionApprovalRequest true "Approval request"
// @Success 200
// @Failure 400 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 404 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/transaction/{transactionID} [post]
func (h *KermesseHandler) HandleValidatePurchase(ctx *gin.Context) {
	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID")))
		return
	}

	transactionID, err := strconv.ParseUint(ctx.Param("transactionID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid transaction ID")))
		return
	}

	var approvalRequest request.StandTransactionApprovalRequest
	if err := ctx.ShouldBindJSON(&approvalRequest); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	if err := approvalRequest.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid request: %w", err)))
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	// Check if user is a participant of the kermesse
	isParticipant, err := h.svc.IsParticipating(user.ID, uint(kermesseID))
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check user participation: %w", err)))
		return
	}
	if !isParticipant {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user is not a participant of this kermesse")))
		return
	}

	if user.Role != "stand_holder" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to validate purchases", user.ID)))
		return
	}

	standHolder, err := h.uSvc.GetStandHolderByUserID(ctx.Request.Context(), user.ID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to get standholder: %w", err)))
		return
	}

	// Fetch the transaction
	transaction, err := h.svc.GetTokenTransactionByID(uint(transactionID))
	if err != nil {
		response.RenderErr(ctx, response.ErrNotFound("transaction", "ID", transactionID))
		return
	}

	// Check if the transaction is pending
	if transaction.Status != "Pending" {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("transaction is not in a pending state")))
		return
	}

	// Verify that the standHolder is associated with the stand
	isAssociated, err := h.svc.IsStandHolderAssociatedWithStand(ctx, standHolder.UserID, *transaction.StandID)
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to check standholder association: %w", err)))
		return
	}
	if !isAssociated {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("standholder is not associated with this stand")))
		return
	}

	var actionMsg string

	if approvalRequest.Approved {
		err = h.svc.ApproveTransaction(ctx, uint(transactionID), standHolder.UserID, approvalRequest.ItemName, approvalRequest.Quantity)
		actionMsg = "approved"
	} else {
		err = h.svc.RejectTransaction(ctx, uint(transactionID), standHolder.UserID)
		actionMsg = "rejected"
	}

	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to %s transaction: %w", actionMsg, err)))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Purchase %s successfully", actionMsg),
	})
}

// HandleGetChildrenTransactions godoc
// @Summary      Get children's transactions for a kermesse
// @Description  Retrieves all transactions made by the children of the authenticated parent user for a specific kermesse. Only parents can access this endpoint.
// @Tags         kermesses,transactions
// @Produce      json
// @Param        kermesseID  path      int  true  "Kermesse ID"
// @Success      200  {array}   domain.TokenTransaction
// @Failure      400  {object}  response.Err
// @Failure      401  {object}  response.Err
// @Failure      403  {object}  response.Err
// @Failure      500  {object}  response.Err
// @Router       /kermesses/{kermesseID}/children_transactions [get]
// @Security BearerAuth
func (h *KermesseHandler) HandleGetChildrenTransactions(ctx *gin.Context) {
	kermesseID, err := strconv.ParseUint(ctx.Param("kermesseID"), 10, 32)
	if err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(fmt.Errorf("invalid kermesse ID")))
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	if user.Role != "parent" {
		response.RenderErr(ctx, response.ErrPermissionDenied(fmt.Errorf("user %v is not authorized to view children transactions", user.ID)))
		return
	}

	transactions, err := h.svc.GetChildrenTransactions(ctx.Request.Context(), user.ID, uint(kermesseID))
	if err != nil {
		response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to get children transactions: %w", err)))
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

// HandleUpdateStock godoc
// @Summary Update stock for a stand
// @Description Allows updating the stock for items in a stand
// @Tags kermesses
// @Accept json
// @Produce json
// @Param standID path int true "Stand ID"
// @Param stockUpdateRequest body request.StockUpdateRequest true "Stock update request"
// @Success 200
// @Failure 400 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/stand/{standID}/stock/update [post]
func (h *KermesseHandler) HandleUpdateStock(ctx *gin.Context) {
	var req request.StockUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	if err := req.Validate(); err != nil {
		response.RenderErr(ctx, response.ErrBadRequest(err))
		return
	}

	standID, err := strconv.ParseUint(ctx.Param("standID"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stand ID"})
		return
	}

	user, respErr := getUserFromContext(ctx, h.uSvc)
	if respErr != nil {
		response.RenderErr(ctx, respErr)
		return
	}

	err = h.svc.UpdateStock(ctx.Request.Context(), req, user.ID, uint(standID))
	if err != nil {
		if errors.Is(err, service.ErrUnauthorizedOrganizer) {
			response.RenderErr(ctx, response.ErrPermissionDenied(err))
		} else {
			response.RenderErr(ctx, response.ErrInternalServerError(fmt.Errorf("failed to update stock: %w", err)))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}
