package gpt3encoder

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewEncoder_e2e(t *testing.T) {
	is := assert.New(t)

	encoder := NewEncoder("gpt-3.5-turbo")

	cases := []lo.Tuple2[string, []int]{
		// lo.T2("", []int{}),	// @TODO
		lo.T2(" ", []int{220}),
		lo.T2("\t", []int{197}),
		lo.T2("tiktoken is great!", []int{83, 1609, 5963, 374, 2294, 0}),
		lo.T2("indivisible", []int{485, 344, 23936}),
		lo.T2("hello 👋 world 🌍", []int{15339, 62904, 233, 1917, 11410, 234, 235}),
		lo.T2("hello, 世界", []int{15339, 11, 220, 3574, 244, 98220}),
	}

	for _, c := range cases {
		encoded, err := encoder.Encode(c.A)
		is.Nil(err)
		is.EqualValues(c.B, encoded, c.A)
		result := encoder.Decode(encoded)
		is.EqualValues(c.A, result, c.A)
	}
}

func TestEncodeLen(t *testing.T) {
	encoder := NewEncoder("gpt-3.5-turbo")
	text := ``
	encoded, err := encoder.Encode(text)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", encoded)
	t.Log(len(encoded))
}
