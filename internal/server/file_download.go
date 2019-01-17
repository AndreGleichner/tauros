package server

import (
	"andre/tauros/api"
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
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

	pwd, _ := os.Getwd()
	outDir := path.Join(pwd, "bin/out")
	lenPrefix := len(outDir) + 1

	resp = &api.NegotiateDownloadFilesResp{}

	err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		resp.Files = append(resp.Files, &api.NegotiateDownloadFilesResp_File{Filename: path[lenPrefix:], Sha256: fileSha256(path)})

		return nil
	})

	return
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
