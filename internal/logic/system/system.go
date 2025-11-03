package system

type sSystem struct{}

func New() *sSystem {
	return &sSystem{}
}

func init() {

}

func (s *sSystem) Init() {}
