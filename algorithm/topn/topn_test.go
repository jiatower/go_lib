package topn

import (
	"testing"
)

type Data struct {
	num  int
	size int64
}

func less(small interface{}, large interface{}) bool {
	return small.(Data).num < large.(Data).num
}
func TestTopn(t *testing.T) {
	a := []interface{}{Data{1, 2}, Data{2, 4}, Data{6, 3}, Data{9, 1}, Data{32, 2}, Data{9, 12}}
	r := TopN(a, less, 3)
	if len(r) != 3 {
		t.Error("len(r)!=3")
	}
}
