package tranx2

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	NoisePrefix      byte   = '#'
	PassingPrefix    byte   = '$'
	MaxTransponderID uint32 = ^uint32(0) >> 8

	// passingMsgLength does not include "\r\n"
	passingMsgLength = 25
	// noiseMsgLength does not include "\r\n"
	noiseMsgLength = 5
)

// Passing struct contains fields present in passing recordsend by TranX2 devices.
// Note that TransponderID is actually uint24 but is presented as uint32 so Marshalling message
// with something bigger than MaxTransponderID will result in TransponderIDOverflow error.
type Passing struct {
	TransponderID uint32 // ID of passing device
	PassingTicks  uint32 // milliseconds elapsed since the device was started.
	Hits          uint8  // number of reads while transponder passed the loop
	Strength      uint8  // signal strenght
	Prefix        uint16 // TODO: figure what is the meaning of this
	Trailing      uint8  // TODO: figure what is the meaning of this
}

// Reader provides API for reading tranx2 encoded data
type Reader struct {
	reader *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(r),
	}
}

// ReadPassing is high level API for reading only Passing records and discarding everything else.
// It is suitable for applications that are only interested on passings and do not care about noise level records.
func (r *Reader) ReadPassing() (rec Passing, err error) {
	var msg []byte
	for len(msg) == 0 || msg[0] != PassingPrefix {
		if msg, _, err = r.reader.ReadLine(); err != nil {
			return
		}
	}
	return UnmarshalPassing(msg)
}

// ReadNoise is high level API for reading only noise levels and discarding everything else.
// It is suitable for applications that are only interested on noise levels and do not care about transponder data.
func (r *Reader) ReadNoise() (noise uint16, err error) {
	var msg []byte
	for len(msg) == 0 || msg[0] != NoisePrefix {
		if msg, _, err = r.reader.ReadLine(); err != nil {
			return
		}
	}
	return UnmarshalNoise(msg)
}

// MessageLenghtError is returned when message byte slice is not what Unmarshal functions are expecting.
var MessageLenghtError = errors.New("invalid message lenght")

type UnmarshalError struct {
	Msg    []byte
	Detail string
	Err    error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("could not unmarshal message '%s: %s: %s", e.Msg, e.Detail, e.Err)
}

// UnmarshalPassing creates Passing struct from TranX2 encoded data.
// TranX2 lines are terminated with "\r\n". Unmarshaller allows you to pass lines with or without "\r\n".
func UnmarshalPassing(msg []byte) (Passing, error) {
	var rec Passing

	msg = bytes.TrimLeft(msg, "\r\n")
	if len(msg) != passingMsgLength {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: fmt.Sprintf("message length %d", len(msg)),
			Err:    fmt.Errorf("invalid message length"),
		}
	}

	var hexErr error
	if rec.Prefix, hexErr = hexBytesToUint16(msg[1:5]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse prefix",
			Err:    hexErr,
		}
	}
	if rec.TransponderID, hexErr = hexBytesToUint24(msg[5:11]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse transponder id",
			Err:    hexErr,
		}
	}
	if rec.PassingTicks, hexErr = hexBytesToUint32(msg[11:19]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse passing ticks",
			Err:    hexErr,
		}
	}
	if rec.Hits, hexErr = hexBytesToUint8(msg[19:21]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse hit",
			Err:    hexErr,
		}
	}
	if rec.Strength, hexErr = hexBytesToUint8(msg[21:23]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse strength",
			Err:    hexErr,
		}
	}
	if rec.Trailing, hexErr = hexBytesToUint8(msg[23:25]); hexErr != nil {
		return rec, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse trailing",
			Err:    hexErr,
		}
	}

	return rec, nil
}

// UnmarshalNoise parses noise level from TranX2 encoded data.
// TranX2 lines are terminated with "\r\n". Unmarshaller allows you to pass lines with or without "\r\n".
func UnmarshalNoise(msg []byte) (uint16, error) {
	msg = bytes.TrimLeft(msg, "\r\n")
	if len(msg) != noiseMsgLength {
		return 0, &UnmarshalError{
			Msg:    msg,
			Detail: fmt.Sprintf("message length %d", len(msg)),
			Err:    fmt.Errorf("invalid message length"),
		}
	}

	noise, err := hexBytesToUint16(msg[1:noiseMsgLength])
	if err != nil {
		return 0, &UnmarshalError{
			Msg:    msg,
			Detail: "failed to parse noise level",
			Err:    err,
		}
	}

	return noise, nil
}

// Writer provides API for writing tranx2 encoded data
type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

// TransponderIDOverflow is returned when transponder id is bigger than MaxUint24.
var TransponderIDOverflow = errors.New("transponder id overflow")

// WritePassing writes given Passing record and returns the number of bytes writen or error.
func (w *Writer) WritePassing(rec Passing) (n int, err error) {
	msg, err := MarshalPassing(rec)
	if err != nil {
		return
	}
	return w.writer.Write(msg)
}

// WriteNoise writes noise level and returns the number of bytes writen or error.
func (w *Writer) WriteNoise(noise uint16) (n int, err error) {
	msg, err := MarshalNoise(noise)
	if err != nil {
		return
	}
	return w.writer.Write(msg)
}

// MarshalPassing marshals Passing struct to byte array
func MarshalPassing(rec Passing) ([]byte, error) {
	if rec.TransponderID > MaxTransponderID {
		return nil, TransponderIDOverflow
	}

	buf := bytes.NewBuffer(make([]byte, 0, 27))
	buf.WriteByte(PassingPrefix)
	buf.WriteString(uint16ToHexStr(rec.Prefix))
	buf.WriteString(uint24ToHexStr(rec.TransponderID))
	buf.WriteString(uint32ToHexStr(rec.PassingTicks))
	buf.WriteString(uint8ToHexStr(rec.Hits))
	buf.WriteString(uint8ToHexStr(rec.Strength))
	buf.WriteString(uint8ToHexStr(rec.Trailing))
	buf.WriteString("\r\n")
	return buf.Bytes(), nil
}

// MarshalNoise marshals uint16 noise value to byte array
func MarshalNoise(noise uint16) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 7))
	buf.WriteByte(NoisePrefix)
	buf.WriteString(uint16ToHexStr(noise))
	buf.WriteString("\r\n")
	return buf.Bytes(), nil
}

func hexBytesToUint8(in []byte) (uint8, error) {
	out, err := hex.DecodeString(string(in))
	if err != nil {
		return 0, err
	}
	return out[0], err
}

func hexBytesToUint16(in []byte) (uint16, error) {
	out, err := hex.DecodeString(string(in))
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(out), err
}

func hexBytesToUint24(in []byte) (uint32, error) {
	out, err := hex.DecodeString("00" + string(in))
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(out), err
}

func hexBytesToUint32(in []byte) (uint32, error) {
	out, err := hex.DecodeString(string(in))
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(out), err
}

func uint8ToHexStr(in uint8) string {
	return strings.ToUpper(hex.EncodeToString([]byte{in}))
}

func uint16ToHexStr(in uint16) string {
	return strings.ToUpper(hex.EncodeToString([]byte{byte(in >> 8), byte(in)}))
}

func uint24ToHexStr(in uint32) string {
	return strings.ToUpper(hex.EncodeToString([]byte{byte(in >> 16), byte(in >> 8), byte(in)}))
}

func uint32ToHexStr(in uint32) string {
	return strings.ToUpper(hex.EncodeToString([]byte{byte(in >> 24), byte(in >> 16), byte(in >> 8), byte(in)}))
}
