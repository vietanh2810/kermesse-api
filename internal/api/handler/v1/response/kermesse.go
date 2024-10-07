package response

type PointsAttributionResponse struct {
	Message          string `json:"message"`
	StudentID        uint   `json:"student_id"`
	PointsAttributed int    `json:"points_attributed"`
	TotalPoints      int    `json:"total_points"`
}
