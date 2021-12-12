package main

import (
	"context"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func Background() Context {
	ws, err := GetWinSize()
	if err != nil {
		panic(err)
	}
	return Context{
		WinSize: ws,
		Width:   int(ws.Cols),
		Height:  int(ws.Rows),
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	ret := ctx
	ret.cancelChan = make(chan struct{})
	return ret, func() {
		close(ret.cancelChan)
	}
}

type Context struct {
	WinSize
	X, Y          int
	Width, Height int
	cancelChan    chan struct{}
}

func (ctx Context) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (ctx Context) Done() <-chan struct{} {
	return ctx.cancelChan
}

func (ctx Context) Err() error {
	return nil
}

func (ctx Context) Value(key interface{}) interface{} {
	return nil
}

func (ctx Context) YCellPixels() int {
	return int(ctx.YPixel) / int(ctx.Rows)
}

func (ctx Context) XCellPixels() int {
	return int(ctx.XPixel) / int(ctx.Cols)
}

const tiocgwinsz = 0x5413

func ioctl(fd, op, arg uintptr) error {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, fd, op, arg)
	if ep != 0 {
		return syscall.Errno(ep)
	}
	return nil
}

type WinSize struct {
	Rows   int16 /* rows, in characters */
	Cols   int16 /* columns, in characters */
	XPixel int16 /* horizontal size, pixels */
	YPixel int16 /* vertical size, pixels */
}

func GetWinSize() (sz WinSize, err error) {
	//TIOCGWINSZ syscall
	for fd := uintptr(0); fd < 3; fd++ {
		if err = ioctl(fd, tiocgwinsz, uintptr(unsafe.Pointer(&sz))); err == nil && sz.XPixel != 0 && sz.YPixel != 0 {
			return
		}
	}
	//if pixels are 0, try CSI
	if sz.XPixel == 0 || sz.YPixel == 0 {
		fmt.Printf("\033[18t")
		fmt.Scanf("\xb1[%d;%dt", &sz.Rows, &sz.Cols)
		//get terminal resolution
		fmt.Printf("\033[14t")
		fmt.Scanf("\033[4;%d;%dt", &sz.YPixel, &sz.XPixel)
	}
	return
}
