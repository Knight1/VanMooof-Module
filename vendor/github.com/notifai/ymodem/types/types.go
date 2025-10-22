package ytypes

type Bar interface {
	Add(int) error
}

type Progress interface {
	Create(string, int) Bar
	Shutdown()
}

type dummyProgress struct{}
type dummyBar struct{}

var _ Progress = (*dummyProgress)(nil)
var _ Bar = (*dummyBar)(nil)

func DummyProgress() Progress {
	return &dummyProgress{}
}

func (b *dummyProgress) Create(name string, ln int) Bar {
	return &dummyBar{}
}

func (b *dummyProgress) Shutdown() {
}

func (b *dummyBar) Add(n int) error {
	return nil
}
