package interaction

type Store interface {
	Create(i *Interaction) error
	Get(token string) (*Interaction, error)
	Update(i *Interaction) error
	Delete(i *Interaction) error
}
