package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// message NegotiateDownloadFilesReq {
// }
// message NegotiateDownloadFilesResp {
//     message File {
//         string  filename = 1; // relative to out dir
//         bytes   sha256 = 2;
//     }
//     repeated File   files = 1; // files the server wants the client to download
// }
// NegotiateDownloadFiles shall be called after any RunCommand() completed to possibly download (further) files via one or more DownloadFile().
func (s *TaurosServer) NegotiateDownloadFiles(ctx context.Context, req *api.NegotiateDownloadFilesReq) (resp *api.NegotiateDownloadFilesResp, err error) {
	log.Printf("NegotiateDownloadFiles")

	return &api.NegotiateDownloadFilesResp{}, nil
}

// message DownloadFileReq {
//     message Meta {
//         string  filename = 1; // relative to out dir
//     }
// }

// message DownloadFileRespStream {
//     message Meta {
//         int32   filesize = 1;
//     }
//     message Chunk {
//         bytes   data = 1;
//         int32   offset = 2;
//     }
//     message FinalStatus {
//         string  err = 1; // empty : success
//     }
//     oneof value {
//         Meta            meta = 1;
//         Chunk           chunk = 2;
//         FinalStatus     final_status = 3;
//     }
// }
// DownloadFile downloads a single file from servers out dir.
func (s *TaurosServer) DownloadFile(req *api.DownloadFileReq, stream api.Tauros_DownloadFileServer) (err error) {
	log.Printf("DownloadFile")

	dlResp := api.DownloadFileRespStream{}
	if err := stream.Send(&dlResp); err != nil {
		return err
	}
	return nil
}
