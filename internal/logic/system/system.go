package system

import (
	"github.com/ayflying/p2p/internal/service"
)

type sSystem struct{}

func New() *sSystem {
	return &sSystem{}
}

func init() {
	service.RegisterSystem(New())
}

func (system *sSystem) Init() {}
