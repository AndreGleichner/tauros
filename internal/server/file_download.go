package server

import (
	"andre/tauros/api"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

// message NegotiateDownloadFilesReq {
// }
// message NegotiateDownloadFilesResp {
//     message File {
//         string  filename = 1; // relative to bin/out dir
//         bytes   sha256 = 2;
//     }
//     repeated File   files = 1; // files the server wants the client to download
// }
// NegotiateDownloadFiles shall be called after any RunCommand() completed to possibly download (further) files via one or more DownloadFile().
func (s *TaurosServer) NegotiateDownloadFiles(ctx context.Context, req *api.NegotiateDownloadFilesReq) (resp *api.NegotiateDownloadFilesResp, err error) {
	log.Printf("NegotiateDownloadFiles")

	pwd, _ := os.Getwd()
	outDir := path.Join(pwd, "bin", "out")
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
//     string  filename = 1; // relative to bin/out dir
// }
// message DownloadFileRespStream {
//     message Meta {
//         int32   filesize = 1;
//     }
//     message Chunk {
//         bytes   data = 1;
//     }
//     message FinalStatus {
//         string  err = 1; // empty : success
//         bytes   sha256 = 2;
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

	pwd, _ := os.Getwd()
	filename := path.Join(pwd, "bin", "out", req.Filename)

	var fd *os.File
	var bytesSent int

	hasher := sha256.New()

	defer func() {
		defer func() {
			r := recover()

			if r != nil {
				err = errors.New(fmt.Sprint("Recoverd ", r))
			}
		}()

		r := recover()

		if r != nil {
			err = errors.New(fmt.Sprint("Recoverd ", r))
		}

		if fd != nil {
			fd.Close()
		}

		err = dlSendFinal(stream, err, hasher.Sum(nil))
	}()

	if fd, err = os.Open(filename); err != nil {
		return
	}

	fi, err := fd.Stat()
	if err != nil {
		return
	}

	filesize := int(fi.Size())

	if err = dlSendMeta(stream, filesize); err != nil {
		return
	}

	const BufferSize = 64 * 1024
	buffer := make([]byte, BufferSize)

	for {
		bytesread, err := fd.Read(buffer)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		bytesSent += bytesread
		if bytesSent > filesize {
			err = errors.New("too many bytes read")
			break
		}

		if _, err = hasher.Write(buffer); err != nil {
			break
		}

		if err = dlSendChunk(stream, buffer); err != nil {
			break
		}
	}

	return
}

func dlSendMeta(stream api.Tauros_DownloadFileServer, filesize int) error {
	resp := api.DownloadFileRespStream{Value: &api.DownloadFileRespStream_Meta_{
		Meta: &api.DownloadFileRespStream_Meta{Filesize: int32(filesize)}}}

	return stream.Send(&resp)
}

func dlSendChunk(stream api.Tauros_DownloadFileServer, data []byte) error {
	resp := api.DownloadFileRespStream{Value: &api.DownloadFileRespStream_Chunk_{
		Chunk: &api.DownloadFileRespStream_Chunk{Data: data}}}

	return stream.Send(&resp)
}

func dlSendFinal(stream api.Tauros_DownloadFileServer, err error, sha256 []byte) error {
	if err != nil {
		sha256 = nil
	}
	resp := api.DownloadFileRespStream{Value: &api.DownloadFileRespStream_FinalStatus_{
		FinalStatus: &api.DownloadFileRespStream_FinalStatus{Err: err.Error(), Sha256: sha256}}}

	return stream.Send(&resp)
}
