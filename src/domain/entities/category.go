package entities

type Category struct {
	Name     string `db:"Name"`
	Sequence string `db:"Sequence"`
	Entity
}
