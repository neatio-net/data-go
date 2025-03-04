package data_test

import (
	"encoding/json"
	"testing"

	data "github.com/neatio-net/data-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoders(t *testing.T) {
	assert := assert.New(t)

	hex := data.HexEncoder
	b64 := data.B64Encoder
	rb64 := data.RawB64Encoder
	cases := []struct {
		encoder         data.ByteEncoder
		input, expected []byte
	}{
		// hexidecimal
		{hex, []byte(`"1a2b3c4d"`), []byte{0x1a, 0x2b, 0x3c, 0x4d}},
		{hex, []byte(`"DE14"`), []byte{0xde, 0x14}},
		// these are errors
		{hex, []byte(`0123`), nil},     // not in quotes
		{hex, []byte(`"dewq12"`), nil}, // invalid chars
		{hex, []byte(`"abc"`), nil},    // uneven length

		// base64
		{b64, []byte(`"Zm9v"`), []byte("foo")},
		{b64, []byte(`"RCEuM3M="`), []byte("D!.3s")},
		// make sure url encoding!
		{b64, []byte(`"D4_a--1="`), []byte{0x0f, 0x8f, 0xda, 0xfb, 0xed}},
		// these are errors
		{b64, []byte(`"D4/a++1="`), nil}, // non-url encoding
		{b64, []byte(`0123`), nil},       // not in quotes
		{b64, []byte(`"hey!"`), nil},     // invalid chars
		{b64, []byte(`"abc"`), nil},      // length%4 != 0

		// raw base64
		{rb64, []byte(`"Zm9v"`), []byte("foo")},
		{rb64, []byte(`"RCEuM3M"`), []byte("D!.3s")},
		// make sure url encoding!
		{rb64, []byte(`"D4_a--1"`), []byte{0x0f, 0x8f, 0xda, 0xfb, 0xed}},
		// these are errors
		{rb64, []byte(`"D4/a++1"`), nil}, // non-url encoding
		{rb64, []byte(`0123`), nil},      // not in quotes
		{rb64, []byte(`"hey!"`), nil},    // invalid chars
		{rb64, []byte(`"abc="`), nil},    // with padding

	}

	for _, tc := range cases {
		var output []byte
		err := tc.encoder.Unmarshal(&output, tc.input)
		if tc.expected == nil {
			assert.NotNil(err, tc.input)
		} else if assert.Nil(err, "%s: %+v", tc.input, err) {
			assert.Equal(tc.expected, output, tc.input)
		}
	}
}

// BData can be encoded/decoded
type BData struct {
	Count int
	Data  data.Bytes
}

// BView is to unmarshall and check the encoding
type BView struct {
	Count int
	Data  string
}

func TestBytes(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		encoder  data.ByteEncoder
		data     data.Bytes
		expected string
	}{
		{data.HexEncoder, []byte{0x1a, 0x2b, 0x3c, 0x4d}, "1A2B3C4D"},
		{data.B64Encoder, []byte("D!.3s"), "RCEuM3M="},
		{data.RawB64Encoder, []byte("D!.3s"), "RCEuM3M"},
	}

	for i, tc := range cases {
		data.Encoder = tc.encoder
		// encode the data
		in := BData{Count: 15, Data: tc.data}
		d, err := json.Marshal(in)
		require.Nil(err, "%d: %+v", i, err)
		// recover the data
		out := BData{}
		err = json.Unmarshal(d, &out)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(in.Count, out.Count, "%d", i)
		assert.Equal(in.Data, out.Data, "%d", i)
		// check the encoding
		view := BView{}
		err = json.Unmarshal(d, &view)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.expected, view.Data)
	}
}

/*** this is example code for the byte array ***/

type Dings [5]byte

func (d Dings) MarshalJSON() ([]byte, error) {
	return data.Encoder.Marshal(d[:])
}

func (d *Dings) UnmarshalJSON(enc []byte) error {
	var ref []byte
	err := data.Encoder.Unmarshal(&ref, enc)
	copy(d[:], ref)
	return err
}

func TestByteArray(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	d := Dings{}
	copy(d[:], []byte("D!.3s"))

	cases := []struct {
		encoder  data.ByteEncoder
		data     Dings
		expected string
	}{
		{data.HexEncoder, Dings{0x1a, 0x2b, 0x3c, 0x4d, 0x5e}, "1A2B3C4D5E"},
		{data.B64Encoder, d, "RCEuM3M="},
		{data.RawB64Encoder, d, "RCEuM3M"},
	}

	for i, tc := range cases {
		data.Encoder = tc.encoder
		// encode the data
		d, err := json.Marshal(tc.data)
		require.Nil(err, "%d: %+v", i, err)
		// recover the data
		out := Dings{}
		err = json.Unmarshal(d, &out)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.data, out, "%d", i)
		// check the encoding
		view := ""
		err = json.Unmarshal(d, &view)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.expected, view)
	}

	// Test invalid data
	invalid := []byte(`"food"`)
	data.Encoder = data.HexEncoder
	ding := Dings{1, 2, 3, 4, 5}
	parsed := ding
	require.Equal(ding, parsed)
	// on a failed parsing, we don't overwrite any data
	err := json.Unmarshal(invalid, &parsed)
	require.NotNil(err)
	assert.Equal(ding, parsed)
}
