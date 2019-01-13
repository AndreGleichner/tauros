package server

import (
	"andre/tauros/api"
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
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
//         bool        needs_reboot = 2;
//     }
//     oneof value {
//         Output          output = 1;
//         FinalStatus     final_status = 2;
//     }
// }

// RunCommand executes a command-line relative to bin dir, streaming back stdout/stderr.
func (s *TaurosServer) RunCommand(req *api.CommandReq, stream api.Tauros_RunCommandServer) error {
	log.Printf("RunCommand " + req.Cmd)

	cmd := exec.Command(req.Cmd)

	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, req.Env...)
	cmd.Stdin = os.Stdin

	time.AfterFunc(time.Duration(req.Timeout.Seconds)*time.Second, func() { cmd.Process.Kill() })

	var err error

	defer func() {
		r := recover()

		if err != nil || r != nil {
			cmd.Process.Kill()
		}

		if r != nil {
			panic(r)
		}
	}()

	var wg sync.WaitGroup

	// Start goroutines that are going to be reading the lines out of stdout/stderr piping into returned channel
	stdoutCh, err := stdStreamChannel(cmd, &wg, false)
	if err != nil {
		return err
	}
	stderrCh, err := stdStreamChannel(cmd, &wg, true)
	if err != nil {
		return err
	}

	// Make sure after we exit we read the lines from stdout forever
	// so they don't block since it is a pipe.
	// The scanner goroutine above will close this, but track it with a wait
	// group for completeness.
	wg.Add(1)
	defer func() {
		go func() {
			defer wg.Done()
			for range stdoutCh {
			}
			for range stderrCh {
			}
		}()
	}()

	go func() {
		for {
			select {
			case line := <-stdoutCh:
				if err := sendOutput(stream, line, false); err != nil {
					return
				}
			case line := <-stderrCh:
				if err := sendOutput(stream, line, true); err != nil {
					return
				}
			}
		}
	}()

	debugMsgArgs := []interface{}{
		"path", cmd.Path,
	}

	var exitCode int32
	err = cmd.Run()
	if err != nil {
		debugMsgArgs = append(debugMsgArgs,
			[]interface{}{"error", err.Error()}...)

		log.Printf("Cmd exited ", debugMsgArgs...)
		//exitCode = err
	}

	wg.Wait()

	sendFinal(stream, exitCode)

	return err
}

func sendOutput(stream api.Tauros_RunCommandServer, line string, isErr bool) error {
	cmdResp := api.CommandRespStream{Value: &api.CommandRespStream_Output_{
		Output: &api.CommandRespStream_Output{Timestamp: time.Now().Unix(), IsErr: isErr, Line: line}}}

	return stream.Send(&cmdResp)
}

func sendFinal(stream api.Tauros_RunCommandServer, exitCode int32) error {
	cmdResp := api.CommandRespStream{Value: &api.CommandRespStream_FinalStatus_{
		FinalStatus: &api.CommandRespStream_FinalStatus{Exitcode: exitCode, NeedsReboot: false}}}

	return stream.Send(&cmdResp)
}

func stdStreamChannel(cmd *exec.Cmd, wg *sync.WaitGroup, errStream bool) (chan string, error) {
	var pipe io.ReadCloser
	var err error

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

	c := make(chan string)
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
