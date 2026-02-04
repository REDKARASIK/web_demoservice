package domain

type Bank struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}
