package server

import (
	"andre/tauros/api"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// message CommandReq {
//     // executable in cmd shall be relative to bin dir, as it gets executed with cur dir set to bin:
//     // Rooted paths are rejected.
//     // Paths containing "/../" are rejected.
//     // Paths not beginning with ./ get prepended with ./
//     string          cmd = 1;
//     Duration        timeout = 2;
//     repeated string env = 3;
// }

// message CommandRespStream {
//     message Output {
//         int64   timestamp = 1;
//         bool    is_err = 2;
//         string  line = 3;
//     }
//     message FinalStatus {
//         int32       exitcode = 1;
//         string      err = 2;
//         bool        needs_reboot = 3;
//     }
//     oneof value {
//         Output          output = 1;
//         FinalStatus     final_status = 2;
//     }
// }

// RunCommand executes a command-line relative to bin dir, streaming back stdout/stderr.
func (s *TaurosServer) RunCommand(req *api.CommandReq, stream api.Tauros_RunCommandServer) (err error) {
	log.Printf("RunCommand " + req.Cmd)

	cmd := exec.Command(req.Cmd)

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, req.Env...)
	cmd.Stdin = os.Stdin

	doneCtx, ctxCancel := context.WithCancel(context.Background())
	if req.Timeout.Seconds > 0 {
		time.AfterFunc(time.Duration(req.Timeout.Seconds)*time.Second, func() { cmd.Process.Kill() })
	}

	defer func() {
		r := recover()

		if r != nil {
			err = errors.New(fmt.Sprint("Recoverd ", r))
		}
	}()

	var wg sync.WaitGroup

	// Start 2 goroutines that are going to be reading the lines out of stdout/stderr piping into returned channel.
	stdoutCh, err := stdStreamChannel(cmd, &wg, false)
	if err != nil {
		return err
	}
	stderrCh, err := stdStreamChannel(cmd, &wg, true)
	if err != nil {
		return err
	}

	// Forward stdout/stderr from launched command back to the client.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case line := <-stdoutCh:
				if err = sendOutput(stream, line, false); err != nil {
					return
				}
			case line := <-stderrCh:
				if err = sendOutput(stream, line, true); err != nil {
					return
				}
			case <-doneCtx.Done():
				return
			}
		}
	}()

	err = cmd.Run()

	var exitCode int
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = -1
		}
	}

	// Wait some short time for possibly remaining stdout/stderr.
	wg.Add(1)
	time.AfterFunc(time.Second, func() { defer wg.Done(); ctxCancel() })

	wg.Wait()

	sendFinal(stream, exitCode, err)

	return err
}

func sendOutput(stream api.Tauros_RunCommandServer, line string, isErr bool) error {
	cmdResp := api.CommandRespStream{Value: &api.CommandRespStream_Output_{
		Output: &api.CommandRespStream_Output{Timestamp: time.Now().Unix(), IsErr: isErr, Line: line}}}

	return stream.Send(&cmdResp)
}

func sendFinal(stream api.Tauros_RunCommandServer, exitCode int, err error) error {
	pwd, _ := os.Getwd()
	rebootMarkerFile := filepath.Join(pwd, "out/NeedReboot.marker")
	needsReboot := fileExists(rebootMarkerFile)

	cmdResp := api.CommandRespStream{Value: &api.CommandRespStream_FinalStatus_{
		FinalStatus: &api.CommandRespStream_FinalStatus{Exitcode: int32(exitCode), Err: err.Error(), NeedsReboot: needsReboot}}}

	return stream.Send(&cmdResp)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func stdStreamChannel(cmd *exec.Cmd, wg *sync.WaitGroup, errStream bool) (c chan string, err error) {
	var pipe io.ReadCloser

	if errStream {
		pipe, err = cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
	} else {
		pipe, err = cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
	}

	c = make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(c)

		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()
	return c, nil
}
