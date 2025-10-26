package pty

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/hajimehoshi/ebiten/v2"
)

// globalPTY stores the active PTY file and protects it via a mutex.
var globalPTY struct {
	f  *os.File
	mu sync.Mutex
}

// startResizeWatcher continuously updates the PTY size based on Ebiten's window.
func startResizeWatcher(f *os.File, charWidth, charHeight int) {
	go func() {
		for {
			w, h := ebiten.WindowSize()
			err := pty.Setsize(f, &pty.Winsize{
				Rows: uint16(h / charHeight),
				Cols: uint16(w / charWidth),
			})
			if err != nil {
				log.Println("[PTY] resize error:", err)
			}
			time.Sleep(time.Second)
		}
	}()
}

// readLoop continuously reads data from the PTY and publishes output events.
// When the shell process exits, it gracefully signals ECS with "system_exit".
func (s *System) readLoop(f *os.File, cmd *exec.Cmd) {
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			// Copy slice to prevent re-use of underlying buffer
			data := append([]byte(nil), buf[:n]...)
			s.bus.Publish("pty_output", data)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("[PTY] shell exited cleanly.")
			} else {
				log.Println("[PTY] read error:", err)
			}
			break
		}
	}

	// Ensure child process terminates fully
	_ = cmd.Wait()

	// Notify ECS to shut down
	s.bus.Publish("system_exit", nil)
	log.Println("[PTY] published system_exit event.")
}

