package gopherbbs

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	"github.com/vutran/ansi"
)

// BBS is an instance of a BBS server
type BBS struct {
	lines []Line

	connections chan io.ReadWriter
}

// NewBBS gives an instance of BBS
func NewBBS() (*BBS, error) {
	return &BBS{
		lines:       []Line{},
		connections: make(chan io.ReadWriter),
	}, nil
}

// AddLine adds a line to listen on for the BBS
func (b *BBS) AddLine(line Line) {
	b.lines = append(b.lines, line)
}

// Start starts a BBS line
func (b *BBS) Start(ctx context.Context) {
	for _, line := range b.lines {
		line.Init()
		go b.listenForCalls(ctx, line)
	}
	<-ctx.Done()
}

func (b *BBS) listenForCalls(ctx context.Context, line Line) {
	callChan, errorChan := line.WaitForCall()
L:
	for {
		select {
		case <-ctx.Done():
			line.Close()
			break L
		case <-callChan:
			log.Println("Got call, connecting...")
			b.listenForConnection(ctx, line)
			break L
		case err := <-errorChan:
			log.Println(err)
			break L
		}
	}
}

func (b *BBS) listenForConnection(ctx context.Context, line Line) {
	connChan, errorChan := line.WaitForConnect()
	select {
	case <-ctx.Done():
		line.Close()
		break
	case <-connChan:
		log.Println("Connected")
		_, w, closed := line.GetReadWriter()
		log.Println("Got RW")
		w.Write([]byte(ansi.EraseDisplay(0)))
		w.Write([]byte("hello there\r\nthis is a demo of GoperBBS"))
		time.Sleep(10 * time.Second)
		w.Write([]byte(ansi.EraseDisplay(0)))
		w.Write([]byte("The magic gopher has greeted you with good luck\r\n"))
		w.Write([]byte(getGopher()))
		time.Sleep(10 * time.Second)
		log.Println("Sent")
		starWars(w)
		<-closed
		go b.listenForCalls(ctx, line) // hung op, back to waiting
		break
	case err := <-errorChan:
		log.Println(err)
		break
	}
}

func getGopher() string {
	out := ""
	out += "\r\n         ,_---~~~~~----._         "
	out += "\r\n  _,,_,*^____      _____``*g*\"*, "
	out += "\r\n / __/ /'     ^.  /      \\ ^@q   f "
	out += "\r\n[  @f | @))    |  | @))   l  0 _/  "
	out += "\r\n \\`/   \\~____ / __ \\_____/    \\   "
	out += "\r\n  |           _l__l_           I   "
	out += "\r\n  }          [______]           I  "
	out += "\r\n  ]            | | |            |  "
	out += "\r\n  ]             ~ ~             |  "
	out += "\r\n  |                            |   "
	out += "\r\n   |                           |   "

	return out
}

func starWars(out io.Writer) {
	conn, _ := net.Dial("tcp", "towel.blinkenlights.nl:23")
	for {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
		out.Write(buf)
	}
}
