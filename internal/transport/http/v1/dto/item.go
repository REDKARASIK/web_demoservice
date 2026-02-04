package dto

type ItemDTO struct {
	ChrtID      int64   `json:"chrt_id"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	RID         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        int64   `json:"sale"`
	Size        string  `json:"size"`
	TotalPrice  float64 `json:"total_price"`
	NmID        int64   `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      int     `json:"status"`
}
