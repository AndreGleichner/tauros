package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// To be called if RunCommand() returned CommandRespStream.final.needs_reboot after any required
// download were completed.
func (s *TaurosServer) Reboot(ctx context.Context, req *api.RebootReq) (resp *api.RebootResp, err error) {
	log.Printf("Reboot")

	resp = &api.RebootResp{}
	err = Reboot()

	return
}
