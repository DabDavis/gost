package pty

import (
    "os"
    "os/exec"
    "github.com/creack/pty"
    "github.com/hajimehoshi/ebiten/v2"
    "log"
)

// after you’ve started the PTY (inside readLoop or just after start) add:
go func() {
    for {
        w, h := ebiten.WindowSize()
        err := pty.Setsize(globalPTY.f, &pty.Winsize{
            Rows: uint16(h / charHeight),
            Cols: uint16(w / charWidth),
        })
        if err != nil {
            log.Println("[PTY] setsiz error:", err)
        }
        ebiten.Sleep(time.Second)  // or some interval
    }
}()

