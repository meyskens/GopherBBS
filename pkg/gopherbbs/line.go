package gopherbbs

import "io"

// Line is an interface for Modem, can be replaced with a debug line or a future implementation like telnet
type Line interface {
	Init() error
	WaitForCall() (chan struct{}, chan error)
	WaitForConnect() (chan struct{}, chan error)
	GetReadWriter() (io.Reader, io.Writer, chan struct{})
	Close() error
}
