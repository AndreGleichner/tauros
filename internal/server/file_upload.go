package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// message NegotiateUploadFilesReq {
//     message File {
//         string  filename = 1; // relative to bin dir
//         bytes   sha256 = 2;
//     }
//     repeated File   files = 1; // files the client has
// }
// message NegotiateUploadFilesResp {
//     repeated int32  indices = 1; // files the server doesn't have yet
// }
// NegotiateUploadFiles shall be called before every RunCommand() to possibly upload (further) files via one or more UploadFile().
func (s *TaurosServer) NegotiateUploadFiles(ctx context.Context, req *api.NegotiateUploadFilesReq) (resp *api.NegotiateUploadFilesResp, err error) {
	log.Printf("NegotiateUploadFiles")

	return &api.NegotiateUploadFilesResp{}, nil
}

// message UploadFileReqStream {
//     message Meta {
//         string  filename = 1; // relative to bin dir
//         int32   filesize = 2;
//     }
//     message Chunk {
//         bytes   data = 1;
//         int32   offset = 2;
//     }
//     oneof value {
//         Meta    meta = 1;
//         Chunk   chunk = 2;
//     }
// }

// message UploadFileResp {
//     string  err = 1; // empty : success
// }
// UploadFile uploads a single file to servers bin dir.
func (s *TaurosServer) UploadFile(stream api.Tauros_UploadFileServer) (err error) {
	log.Printf("UploadFile")

	return stream.SendAndClose(&api.UploadFileResp{
		Err: "Failed",
	})
}
