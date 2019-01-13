package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// NegotiateDownloadFiles shall be called after any RunCommand() completed to possibly download (further) files via one or more DownloadFile().
func (s *TaurosServer) NegotiateDownloadFiles(ctx context.Context, req *api.NegotiateDownloadFilesReq) (resp *api.NegotiateDownloadFilesResp, err error) {
	log.Printf("NegotiateDownloadFiles")

	return &api.NegotiateDownloadFilesResp{}, nil
}

// DownloadFile downloads a single file from servers out dir.
func (s *TaurosServer) DownloadFile(req *api.DownloadFileReq, stream api.Tauros_DownloadFileServer) (err error) {
	log.Printf("DownloadFile")

	dlResp := api.DownloadFileRespStream{}
	if err := stream.Send(&dlResp); err != nil {
		return err
	}
	return nil
}
