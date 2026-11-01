package cinema

import "time"

type Screening struct {
	ID             int64     `json:"id"`
	FilmTitle      string    `json:"film_title"`
	StartsAt       time.Time `json:"starts_at"`
	AvailableSeats int       `json:"available_seats"`
}

type ScreeningFilter struct {
	Search     string
	StartsFrom *time.Time
	MinSeats   *int
	Limit      uint64
}

type Booking struct {
	ID          int64     `json:"id"`
	ScreeningID int64     `json:"screening_id"`
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}
