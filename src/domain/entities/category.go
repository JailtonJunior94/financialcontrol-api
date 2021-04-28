package entities

type Category struct {
	Name     string `db:"Name"`
	Sequence int    `db:"Sequence"`
	Entity
}
