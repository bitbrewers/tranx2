package tranx2

import (
	"bufio"
	"io"

	"github.com/jacobsa/go-serial/serial"
)

type SerialOptions serial.OpenOptions

// Tranx2Handler interface requiresall callback functions
// that are needed for handling tranx2 events.
type Tranx2Handler interface {
	OnPassing(rec Passing)
	OnNoise(noise uint16)
	OnError(err error)
}

// Client provides high level callback API for TranX-2 serial connection.
type Client struct {
	Opts    SerialOptions
	Handler Tranx2Handler
	Conn    io.ReadCloser
}

// NewClient returns TranX2 client with default configuration.
// Configuration can be modified before calling the Listen receiver.
func NewClient(portName string, handler Tranx2Handler) *Client {
	return &Client{
		Opts: SerialOptions{
			PortName:        portName,
			BaudRate:        9600,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 2,
		},
		Handler: handler,
	}
}

// Listen opens serial connections usign options in Client.
func (c *Client) Listen() (err error) {
	c.Conn, err = serial.Open(serial.OpenOptions(c.Opts))
	return
}

// Serve starts reading from connection and calling handler callbacks.
func (c *Client) Serve() (err error) {
	r := bufio.NewReader(c.Conn)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case NoisePrefix:
			if noise, err := UnmarshalNoise(line); err != nil {
				c.Handler.OnError(err)
			} else {
				c.Handler.OnNoise(noise)
			}
		case PassingPrefix:
			if rec, err := UnmarshalPassing(line); err != nil {
				c.Handler.OnError(err)
			} else {
				c.Handler.OnPassing(rec)
			}
		}
	}
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
