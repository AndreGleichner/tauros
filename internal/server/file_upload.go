package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// NegotiateUploadFiles shall be called before every RunCommand() to possibly upload (further) files via one or more UploadFile().
func (s *TaurosServer) NegotiateUploadFiles(ctx context.Context, req *api.NegotiateUploadFilesReq) (*api.NegotiateUploadFilesResp, error) {
	log.Printf("NegotiateUploadFiles")

	return &api.NegotiateUploadFilesResp{}, nil
}

// UploadFile uploads a single file to servers bin dir.
func (s *TaurosServer) UploadFile(stream api.Tauros_UploadFileServer) error {
	log.Printf("UploadFile")

	return stream.SendAndClose(&api.UploadFileResp{
		ErrorMessage: "Failed",
	})
}
