package domain

type Item struct {
	ID          int64   `db:"id"`
	ChrtID      *int64  `db:"chart_id"` // can be NULL
	TrackNumber string  `db:"track_number"`
	Price       float64 `db:"price"`
	RID         string  `db:"rid"`
	Name        string  `db:"name"`
	Sale        *int64  `db:"sale"`
	Size        *string `db:"size"`
	TotalPrice  float64 `db:"total_price"`
	NmID        int64   `db:"nm_id"`
	Brand       string  `db:"brand"`
	Status      int     `db:"status"`
}
