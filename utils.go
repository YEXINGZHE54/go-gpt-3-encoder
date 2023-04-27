package gpt3encoder

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	"github.com/samber/lo"
)

func loadTokenBPE(filename string) (data map[string]int, err error) {
	var bpeData []byte
	if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		resp, err := http.Get(filename)
		if err != nil {
			return nil, err
		}
		bpeData, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	} else {
		bpeData, err = files.ReadFile(filename)
	}
	if err != nil {
		return
	}
	bpeLines := bytes.Split(bpeData, []byte{'\n'})
	bpeMerges := lo.Map(bpeLines[0:len(bpeLines)-1], func(line []byte, _ int) lo.Tuple2[[]byte, int] {
		parts := bytes.SplitN(line, []byte{' '}, 2)
		decoded, err := base64.StdEncoding.DecodeString(b2s(parts[0]))
		if err != nil {
			panic(err)
		}
		rank, err := strconv.Atoi(b2s(parts[1]))
		if err != nil {
			panic(err)
		}
		return lo.T2(decoded, rank)
	})
	return lo.SliceToMap(bpeMerges, func(item lo.Tuple2[[]byte, int]) (string, int) {
		return string(item.A), item.B
	}), nil
}

// carefully use it
func b2s(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&b))
}
