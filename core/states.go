package core

// State representa os 3 estados do filósofo (na sequência)
// tranquilo -> com sede -> bebendo.
type State int

const (
	Tranquilo State = iota // pensando / ocioso
	ComSede                // tentando adquirir todas as garrafas
	Bebendo                // seção crítica (1 segundo)
)

func (s State) String() string {
	switch s {
	case Tranquilo:
		return "tranquilo"
	case ComSede:
		return "com sede"
	case Bebendo:
		return "bebendo"
	default:
		return "desconhecido"
	}
}
