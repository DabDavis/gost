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

// globalPTY stores the active PTY file and mutex.
var globalPTY struct {
	f  *os.File
	mu sync.Mutex
}

// startResizeWatcher resizes the PTY dynamically with the window.
func startResizeWatcher(f *os.File, charWidth, charHeight int) {
	go func() {
		for {
			w, h := ebiten.WindowSize()
			if err := pty.Setsize(f, &pty.Winsize{
				Rows: uint16(h / charHeight),
				Cols: uint16(w / charWidth),
			}); err != nil {
				log.Println("[PTY] resize error:", err)
			}
			time.Sleep(time.Second)
		}
	}()
}

// readLoop reads PTY output and emits events until the shell exits.
func (s *System) readLoop(f *os.File, cmd *exec.Cmd) {
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
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
	_ = cmd.Wait()
	s.bus.Publish("system_exit", nil)
	log.Println("[PTY] published system_exit event.")
}

