package tranx2

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestHandler chan interface{}

func (h TestHandler) OnPassing(rec Passing) { h <- rec }
func (h TestHandler) OnNoise(noise uint16)  { h <- noise }
func (h TestHandler) OnError(err error)     { h <- err }

type StringRC struct {
	*strings.Reader
}

func (s *StringRC) Close() error { return nil }

func TestNewClient(t *testing.T) {
	const input = "#0970\r\n$09000235BF3436E021358700\r\n\r\n$09000235BF3436\r\n#097j\r\n"
	expected := []interface{}{
		uint16(2416),
		Passing{TransponderID: 144831, PassingTicks: 876011553, Hits: 53, Strength: 135, Prefix: 2304, Trailing: 0},
	}
	src := &StringRC{strings.NewReader(input)}
	h := make(TestHandler)

	c := NewClient("dummy/port", h)
	c.Conn = src

	done := make(chan struct{})
	go func() {
		for _, v := range expected {
			assert.Equal(t, v, <-h)
		}
		assert.Error(t, (<-h).(error))
		assert.Error(t, (<-h).(error))
		done <- struct{}{}
	}()

	err := c.Serve()
	assert.Equal(t, io.EOF, err)
	c.Close()
	<-done
}
