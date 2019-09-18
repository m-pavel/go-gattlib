package tion

type Tion interface {
	ReadState(tmt int) (*Status, error)
	Update(s *Status) error
}
