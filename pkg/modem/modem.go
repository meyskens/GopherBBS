package modem

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jacobsa/go-serial/serial"
)

// Modem exposes an interface to talk to a serial modem
type Modem struct {
	port     string
	baudRate uint

	rw     io.ReadWriteCloser
	reader *bufio.Reader
}

// NewModem gives an initialized Modem struct for a serial address and baud rate
func NewModem(port string, baudRate uint) (*Modem, error) {
	return &Modem{
		port:     port,
		baudRate: baudRate,
	}, nil
}

// Init starts up the modem
func (m *Modem) Init() error {
	options := serial.OpenOptions{
		PortName:        m.port,
		BaudRate:        m.baudRate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	var err error
	m.rw, err = serial.Open(options)
	if err != nil {
		return err
	}

	m.reader = bufio.NewReader(m.rw)

	return nil
}

// WaitForCall waits for a ring signal from the modem, it will pick up once it sees one
func (m *Modem) WaitForCall() (chan struct{}, chan error) {
	callChan := make(chan struct{})
	errorChan := make(chan error)

	go func() {
		for {
			line, _, err := m.reader.ReadLine()
			log.Println(string(line))
			if err != nil {
				errorChan <- err
				close(errorChan)
				close(callChan)
				break
			}
			if m.cleanLine(string(line)) == "RING" {
				m.rw.Write([]byte("\r\nATA\r\n")) // AT ANSWER
				callChan <- struct{}{}
				close(errorChan)
				close(callChan)
				break
			}
		}
	}()

	return callChan, errorChan
}

// WaitForConnect will wait till the modem confirms a connection is made after a call is picked up
func (m *Modem) WaitForConnect() (chan struct{}, chan error) {
	connChan := make(chan struct{})
	errorChan := make(chan error)
	go func() {
		for {
			line, _, err := m.reader.ReadLine()
			log.Println(string(line))
			if err != nil {
				errorChan <- err
				close(errorChan)
				close(connChan)
				break
			}
			if strings.Contains(m.cleanLine(string(line)), "CONNECT") {
				connChan <- struct{}{}
				close(errorChan)
				close(connChan)
				break
			} else if m.cleanLine(string(line)) == "NO CARRIER" {
				errorChan <- errors.New("No carrier on line")
				break
			}
		}
	}()

	return connChan, errorChan
}

// GetReadWriter exposes the raw serial line with an end channel that resturns a value when the caller hangs up
func (m *Modem) GetReadWriter() (io.Reader, io.Writer, chan struct{}) {
	closedChan := make(chan struct{})
	return m.checkReader(m.rw, closedChan), m.checkWriter(m.rw, closedChan), closedChan
}

func (m *Modem) checkReader(in io.Reader, end chan struct{}) io.Reader {
	r, w := io.Pipe()

	go func() {
		messageBuffer := []byte{}
		for {
			buf := make([]byte, 1)
			in.Read(buf)
			messageBuffer = append(messageBuffer, buf...)
			go w.Write(buf) //TODO: remove this routine here!
			fmt.Print(string(buf))
			if len(messageBuffer) > 100 {
				messageBuffer = messageBuffer[1:]
			}

			if strings.Contains(string(messageBuffer), "NO CARRIER") { // end on NO CARRIER message
				end <- struct{}{}
				m.rw.Write([]byte("\r\nATH\r\n")) // AT HANG UP
				break
			}
		}
	}()

	return r
}

func (m *Modem) checkWriter(in io.Writer, end chan struct{}) io.Writer {
	r, w := io.Pipe()

	go func() {
		messageBuffer := []byte{}
		for {
			buf := make([]byte, 1)
			r.Read(buf)
			messageBuffer = append(messageBuffer, buf...)
			in.Write(buf)
			//fmt.Print(string(buf))
			if len(messageBuffer) > 100 {
				messageBuffer = messageBuffer[1:]
			}
			// TODO do something with this buffer
		}
	}()

	return w
}

// Close closes the connection to the modem
func (m *Modem) Close() error {
	m.rw.Write([]byte("\r\nATH\r\n")) // AT HANG UP
	return m.rw.Close()
}

func (m *Modem) cleanLine(in string) string {
	in = strings.Trim(in, "\n")
	in = strings.Trim(in, "\r")
	in = strings.Trim(in, " ")

	return in
}
