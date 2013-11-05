package hash

import (
	"fmt"
	"testing"
  "sort"
)

func TestRand64(t *testing.T) {
	var i interface{}
	i = Rand64()
	v, ok := i.(int64)
	if !ok {
		t.Errorf("Rand64 returned wrong type")
	}
	if v < 0 {
		t.Errorf("Rand64 returned a too small number [%d]", v)
	}
}

func TestMd5Bytes(t *testing.T) {
	var blob = []byte("Hurp til you Derp")
	var expected = "3ef08fa896a154eee3c97f037c9d6dfc"
	var actual = fmt.Sprintf("%x", GetMd5FromBytes(blob))
	if actual != expected {
		t.Errorf("Md5FromBytes sum did not match! %s != %s", actual, expected)
	}
}

func TestMd5String(t *testing.T) {
	var blob = "Hurp til you Derp"
	var expected = "3ef08fa896a154eee3c97f037c9d6dfc"
	var actual = fmt.Sprintf("%x", GetMd5FromString(blob))
	if actual != expected {
		t.Errorf("Md5FromString sum did not match! %s != %s", actual, expected)
	}
}

func TestHash(t *testing.T) {
  seen := []string{}
  for i := 0; i < 10000; i++ {
    h := GetSmallHash()
    j := sort.SearchStrings(seen, h)
    if len(seen) != 0 && j < len(seen) && h == seen[j] {
      t.Errorf("there is a non-unique hash after %d attempts [%s and %s]", i, h, seen[j])
    } else {
      seen = append(seen, h)
      sort.Strings(seen)
    }
  }
}
