package credible

import (
	"testing"
)

func TestGenUid(t *testing.T) {

	uid := genUid()
	t.Error(uid)

}

func TestCheckHead(t *testing.T) {
	// head := []byte{'N', 'T', 'E', 27}
	// tf, df := checkHead(head...)
	// t.Error(tf, df)
}

func TestNewFile(t *testing.T) {
	// NewFile("test.log")
}

func TestCheckJson(t *testing.T) {
	// str := []byte(`{"test":1,"hello":"world"}{"test":1}`)
	// real, rem := checkJson(str)
	// t.Error(string(real), string(rem), "eer")
}

func TestCovertI2B(t *testing.T) {
	ts := []struct {
		indata  int
		outdata [4]byte
	}{
		{indata: 1,
			outdata: [4]byte{0, 0, 0, 1}},
		{indata: 12,
			outdata: [4]byte{0, 0, 0, 12}},
		{
			indata:  1111,
			outdata: [4]byte{0, 0, 4, 87},
		},
		{
			indata:  256,
			outdata: [4]byte{0, 0, 1, 0},
		},
		{
			indata:  512,
			outdata: [4]byte{0, 0, 2, 0},
		},
		{
			indata:  1024,
			outdata: [4]byte{0, 0, 4, 0},
		},
	}

	for _, tc := range ts {
		ret := convertI2B(tc.indata)
		if ret != tc.outdata {
			t.Error("real:", ret, "want:", tc.outdata)
		}
	}

}

func TestB2I(t *testing.T) {
	// b := []byte{0, 0, 0, 18}
	// i, err := convertB2I(b)
	// t.Error(i, err)
}

func TestLen(t *testing.T) {
	// bs := []byte("123142423542")
	// t.Error(len(bs))
	// t.Error(len(bs[4:9]))
}
