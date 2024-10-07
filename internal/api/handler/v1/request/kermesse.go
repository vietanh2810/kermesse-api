package request

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
)

type StockItem struct {
	ItemName  string `json:"item_name"`
	Quantity  int    `json:"quantity"`
	TokenCost int    `json:"token_cost"`
}

type CreateKermesseRequest struct {
	Name        string `json:"name" binding:"required"`
	Date        string `json:"date" binding:"required" format:"DD/MM/YYYY"`
	Location    string `json:"location" binding:"required"`
	Description string `json:"description"`
}

type SendTokensRequest struct {
	StudentID uint `json:"student_id" binding:"required"`
	Amount    int  `json:"amount" binding:"required,min=1"`
}

func (req *SendTokensRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.StudentID, validation.Required, validation.Min(uint(1))),
		validation.Field(&req.Amount, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (req *CreateKermesseRequest) Validate() error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&req.Date, validation.Required),
		validation.Field(&req.Location, validation.Required, validation.Length(2, 50)),
		validation.Field(&req.Description, validation.Length(0, 200)),
	)
}

type CreateStandRequest struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Stock       []StockItem `json:"stock"`
}

type TokenPurchaseRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

type StandPurchaseRequest struct {
	ItemName string `json:"itemName" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

type StandTransactionApprovalRequest struct {
	Approved      bool   `json:"approved" binding:"required"`
	TransactionID uint   `json:"transaction_id" binding:"required"`
	ItemName      string `json:"itemName" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=1"`
}

func (req *StandTransactionApprovalRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.TransactionID, validation.Required, validation.Min(uint(1))),
		validation.Field(&req.ItemName, validation.Required, validation.Length(1, 50)),
		validation.Field(&req.Quantity, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return err
	}
	return nil
}

type StockUpdateRequest struct {
	StockID   uint   `json:"stock_id"`
	ItemName  string `json:"item_name"`
	Quantity  int    `json:"quantity"`
	TokenCost int    `json:"token_cost"`
}

func (req *StockUpdateRequest) Validate() error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.StockID, validation.Required, validation.Min(uint(1))),
		validation.Field(&req.ItemName, validation.Required, validation.Length(1, 50)),
		validation.Field(&req.Quantity, validation.Required, validation.Min(0)),
		validation.Field(&req.TokenCost, validation.Required, validation.Min(1)),
	)
}

func (req *StandPurchaseRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.ItemName, validation.Required, validation.Length(1, 50)),
		validation.Field(&req.Quantity, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (req *TokenPurchaseRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.Amount, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return err
	}
	return nil
}

func (req *CreateStandRequest) Validate() error {
	err := validation.ValidateStruct(
		req,
		validation.Field(&req.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&req.Type, validation.Required, validation.In("food", "drink", "activity")),
		validation.Field(&req.Description, validation.Length(0, 200)),
		validation.Field(&req.Stock, validation.Required, validation.Each(validation.By(validateStockItem))),
	)
	if err != nil {
		return err
	}

	return nil
}

func validateStockItem(value interface{}) error {
	item, ok := value.(StockItem)
	if !ok {
		return fmt.Errorf("invalid stock item")
	}

	return validation.ValidateStruct(&item,
		validation.Field(&item.ItemName, validation.Required, validation.Length(1, 50)),
		validation.Field(&item.Quantity, validation.Required, validation.Min(0)),
		validation.Field(&item.TokenCost, validation.Required, validation.Min(1)),
	)
}
