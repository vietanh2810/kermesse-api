package domain

import "time"

type ChatMessage struct {
	ID         uint      `json:"id"`
	KermesseID uint      `json:"kermesse_id"`
	StandID    uint      `json:"stand_id"`
	SenderID   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}
