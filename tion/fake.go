package tion

type fakeTion struct {
	s Status
}

func NewFake() Tion {
	return &fakeTion{}
}

func (ft fakeTion) ReadState(tmt int) (*Status, error) {
	return &ft.s, nil
}
func (ft *fakeTion) Update(s *Status) error {
	ft.s = *s
	return nil
}
