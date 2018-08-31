package tranx2

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReaderPassings(t *testing.T) {
	const input = "#0970\r\n#096F\r\n#0971\r\n#096E\r\n#096C\r\n#0968\r\n#096D\r\n#096E\r\n#0970\r\n#0971\r\n#096F\r\n$09000235BF3436E021358700\r\n#096F\r\n#0970\r\n#096E\r\n#096C\r\n#0974\r\n$09000235BF34374BDD318700\r\n#0974\r\n#096D\r\n#096F\r\n#0971\r\n#0973\r\n#096D\r\n$09000235BF3437BEAC3E8800\r\n"
	var err error
	var rec Passing
	expected := []Passing{
		{TransponderID: 144831, PassingTicks: 876011553, Hits: 53, Strength: 135, Prefix: 2304, Trailing: 0},
		{TransponderID: 144831, PassingTicks: 876039133, Hits: 49, Strength: 135, Prefix: 2304, Trailing: 0},
		{TransponderID: 144831, PassingTicks: 876068524, Hits: 62, Strength: 136, Prefix: 2304, Trailing: 0},
	}
	got := make([]Passing, 0, 3)
	r := NewReader(strings.NewReader(input))
	for {
		if rec, err = r.ReadPassing(); err != nil {
			break
		}
		got = append(got, rec)
	}

	require.Equal(t, io.EOF, err, err.Error())
	require.Equal(t, expected, got)
}
func TestNewReaderNoise(t *testing.T) {
	const input = "#0970\r\n#096F\r\n#0971\r\n#096E\r\n#096C\r\n#0968\r\n#096D\r\n#096E\r\n#0970\r\n"
	expected := []uint16{2416, 2415, 2417, 2414, 2412, 2408, 2413, 2414, 2416}
	got := make([]uint16, 0, 9)
	r := NewReader(strings.NewReader(input))
	var noise uint16
	var err error
	for {
		if noise, err = r.ReadNoise(); err != nil {
			break
		}
		got = append(got, noise)
	}

	require.Equal(t, io.EOF, err, err.Error())
	require.Equal(t, expected, got)
}

func TestNewReaderPassingErrors(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{name: "invalid lenght", data: "$09000235BF3436E0213587\r\n"},
		{name: "invalid hex character for prefix", data: "$09M00235BF3036E021358700\r\n"},
		{name: "invalid hex character for transponder id", data: "$090002M5BF3036E021358700\r\n"},
		{name: "invalid hex character for ticks", data: "$09000235BF3M36E021358700\r\n"},
		{name: "invalid hex character for hits", data: "$09000235BF3036E021M58700\r\n"},
		{name: "invalid hex character for strength", data: "$09000235BF3036E02155M700\r\n"},
		{name: "invalid hex character for trailing", data: "$09000235BF3036E02135870M\r\n"},
	}
	buf := &bytes.Buffer{}
	r := NewReader(buf)
	for _, tc := range testCases {
		buf.WriteString(tc.data)
		_, err := r.ReadPassing()
		err.Error()
		assert.Error(t, err, "did not receive "+tc.name+" error")
	}
}

func TestNewReaderNoiseErrors(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{name: "invalid lenght", data: "#090\r\n"},
		{name: "invalid hex character", data: "#097J\r\n"},
	}
	buf := &bytes.Buffer{}
	r := NewReader(buf)
	for _, tc := range testCases {
		buf.WriteString(tc.data)
		_, err := r.ReadNoise()
		assert.Error(t, err, "did not receive "+tc.name+" error")
	}
}

func TestNewWriterPassings(t *testing.T) {
	const expected = "$09000235BF3436E021358700\r\n$09000235BF34374BDD318700\r\n$09000235BF3437BEAC3E8800\r\n"
	input := []Passing{
		{TransponderID: 144831, PassingTicks: 876011553, Hits: 53, Strength: 135, Prefix: 2304, Trailing: 0},
		{TransponderID: 144831, PassingTicks: 876039133, Hits: 49, Strength: 135, Prefix: 2304, Trailing: 0},
		{TransponderID: 144831, PassingTicks: 876068524, Hits: 62, Strength: 136, Prefix: 2304, Trailing: 0},
	}
	s := &strings.Builder{}
	w := NewWriter(s)
	for _, rec := range input {
		if _, err := w.WritePassing(rec); err != nil {
			require.NoError(t, err)
		}
	}
	require.Equal(t, expected, s.String())
}

func TestNewWriterNoise(t *testing.T) {
	const expected = "#0970\r\n#096F\r\n#0971\r\n#096E\r\n#096C\r\n#0968\r\n#096D\r\n#096E\r\n#0970\r\n"
	input := []uint16{2416, 2415, 2417, 2414, 2412, 2408, 2413, 2414, 2416}
	s := &strings.Builder{}
	w := NewWriter(s)
	for _, noise := range input {
		if _, err := w.WriteNoise(noise); err != nil {
			require.NoError(t, err)
		}
	}
	require.Equal(t, expected, s.String())
}

func TestNewWriterErrors(t *testing.T) {
	rec := Passing{
		TransponderID: ^uint32(0),
		PassingTicks:  876011553,
		Hits:          53,
		Strength:      135,
		Prefix:        2304,
		Trailing:      0,
	}
	s := &strings.Builder{}
	w := NewWriter(s)
	_, err := w.WritePassing(rec)
	require.Equal(t, TransponderIDOverflow, err)
}
