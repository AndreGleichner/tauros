package server

import (
	"andre/tauros/api"
	"log"
)

// RunCommand executes a command-line relative to bin dir, streaming back stdout/stderr.
func (s *TaurosServer) RunCommand(req *api.CommandReq, stream api.Tauros_RunCommandServer) error {
	log.Printf("NegotiateUploadFiles")

	cmdResp := api.CommandRespStream{}
	if err := stream.Send(&cmdResp); err != nil {
		return err
	}
	return nil
}
