package event

type BookingCreatedEvent struct {
	BookingID string  `json:"booking_id"`
	UserID    string  `json:"user_id"`
	ConcertID string  `json:"concert_id"`
	Seat      int     `json:"seat"`
	Amount    float64 `json:"amount"`
}
