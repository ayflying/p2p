package http

import v1 "github.com/ayflying/p2p/api/p2p/v1"

type Http struct {
	Ip    string `json:"ip"`
	Port  int    `json:"port"`
	Cname string `json:"cname"`
	Data  []byte `json:"data"`
}

var New = Http{}

func (h Http) Message(msg *v1.Message) (err error) {
	//TODO implement me
	panic("implement me")
}
