package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateBookingRequest struct {
	ConcertID string `json:"concert_id" validate:"required,uuid"`
	Seat      int    `json:"seat" validate:"required,min=1"`
}
