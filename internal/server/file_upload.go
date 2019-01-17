package server

import (
	"andre/tauros/api"
	"bytes"
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

	pwd, _ := os.Getwd()
	binDir := path.Join(pwd, "bin")

	resp = &api.NegotiateUploadFilesResp{}

	for i, file := range req.Files {
		filename := filepath.Join(binDir, file.Filename)
		if !bytes.Equal(fileSha256(filename), file.Sha256) {
			resp.Indices = append(resp.Indices, int32(i))
		}
	}

	return
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

	pwd, _ := os.Getwd()
	binDir := path.Join(pwd, "bin")

	var meta *api.UploadFileReqStream_Meta
	var chunk *api.UploadFileReqStream_Chunk
	var bytesReceived int

	var fd *os.File
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

		if bytesReceived != int(meta.Filesize) {
			err = errors.New("wrong byte count")
		} else {
			sha256 := hasher.Sum(nil)

			if meta != nil && !bytes.Equal(sha256, fileSha256(meta.Filename)) {
				err = errors.New("invalid hash")
			}
		}

		if err == nil {
			err = stream.SendAndClose(&api.UploadFileResp{})
		} else {
			if meta != nil {
				os.Remove(filepath.Join(binDir, meta.Filename))
			}
			err = stream.SendAndClose(&api.UploadFileResp{Err: err.Error()})
		}
	}()

EXIT_FOR:
	for {
		st, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		switch st.Value.(type) {
		case *api.UploadFileReqStream_Meta_:
			if meta != nil {
				err = errors.New("seeing meta 2nd time")
				break EXIT_FOR
			}
			meta = st.GetMeta()
			if fd, err = os.Create(filepath.Join(binDir, meta.Filename)); err != nil {
				break EXIT_FOR
			}

		case *api.UploadFileReqStream_Chunk_:
			if meta == nil {
				err = errors.New("no meta seen")
				break EXIT_FOR
			}

			chunk = st.GetChunk()
			bytesReceived += len(chunk.Data)
			if bytesReceived > int(meta.Filesize) {
				err = errors.New("too many bytes")
				break EXIT_FOR
			}
			hasher.Write(chunk.Data)

			if _, err = fd.Write(chunk.Data); err != nil {
				break EXIT_FOR
			}

		default:
			break EXIT_FOR
		}
	}

	return
}
