package gpt3encoder

import (
	"bytes"
	"embed"
	"encoding/base64"
	"math"
	"strconv"

	"github.com/dlclark/regexp2"
	"github.com/samber/lo"
)

const (
	ENDOFTEXT   = "<|endoftext|>"
	FIM_PREFIX  = "<|fim_prefix|>"
	FIM_MIDDLE  = "<|fim_middle|>"
	FIM_SUFFIX  = "<|fim_suffix|>"
	ENDOFPROMPT = "<|endofprompt|>"
)

//go:embed cl100k_base.tiktoken
var files embed.FS

// var pat = regexp2.MustCompile(`/'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+/`, 0)

func loadTokenBPE(filename string) (data map[string]int, err error) {
	bpeData, err := files.ReadFile(filename)
	if err != nil {
		return
	}
	bpeLines := bytes.Split(bpeData, []byte{'\n'})
	bpeMerges := lo.Map(bpeLines[0:len(bpeLines)-1], func(line []byte, _ int) lo.Tuple2[[]byte, int] {
		parts := bytes.SplitN(line, []byte{' '}, 2)
		decoded, err := base64.StdEncoding.DecodeString(string(parts[0]))
		if err != nil {
			panic(err)
		}
		rank, err := strconv.Atoi(string(parts[1]))
		if err != nil {
			panic(err)
		}
		return lo.T2(decoded, rank)
	})
	return lo.SliceToMap(bpeMerges, func(item lo.Tuple2[[]byte, int]) (string, int) {
		return string(item.A), item.B
	}), nil
}

func specialTokens() map[string]int {
	return map[string]int{
		ENDOFTEXT:   100257,
		FIM_PREFIX:  100258,
		FIM_MIDDLE:  100259,
		FIM_SUFFIX:  100260,
		ENDOFPROMPT: 100276,
	}
}

type Encoder struct {
	pat     *regexp2.Regexp
	encoder map[string]int
	decoder map[int]string
}

func NewEncoderFromConfig(conf *encoderConfig) (*Encoder, error) {
	encoder, err := loadTokenBPE(conf.filename)
	if err != nil {
		return nil, err
	}
	decoder := lo.Invert(encoder)

	enc := Encoder{
		// pat:     regexp2.MustCompile(`/(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+/`, 0),
		pat:     regexp2.MustCompile(conf.pattern, 0),
		encoder: encoder,
		decoder: decoder,
	}

	return &enc, nil
}

func bpe_merge(piece string, ranks map[string]int) (tokens []int) {
	//init parts, each item is {startPosInPiece, rank}
	parts := make([]*lo.Tuple2[int, int], len(piece)+1)
	for i := range parts {
		parts[i] = &lo.Tuple2[int, int]{A: i, B: math.MaxInt}
	}

	//calc rank func
	emptyRank := EmptyOption[int]()
	get_rank := func(start int, skip int) Optional[int] {
		if start+skip+2 < len(parts) {
			startIdx := parts[start].A
			endIdx := parts[start+skip+2].A
			r, ok := ranks[piece[startIdx:endIdx]]
			if !ok {
				return emptyRank
			}
			return Optional[int]{&r}
		} else {
			return emptyRank
		}
	}

	// calc rank once
	for i := 0; i < len(parts)-2; i++ {
		res := get_rank(i, 0)
		if res.IsPresent() {
			parts[i].B = res.Get()
		}
	}
	// now, loop and update rank if merge needed
	for {
		if len(parts) == 1 {
			break
		}
		//search for min rank in parts
		minRank := lo.Tuple2[int, int]{A: 0, B: math.MaxInt} // {idxInParts, rank}
		for i, v := range parts {
			if v.B < minRank.B {
				minRank.A = i
				minRank.B = v.B
			}
		}
		if minRank.B == math.MaxInt { //not found?
			break
		}
		i := minRank.A
		//we found a min part, we need to update rank and remove parts[i+1]
		parts[i].B = get_rank(i, 1).GetOr(math.MaxInt)
		if i > 0 {
			parts[i-1].B = get_rank(i-1, 1).GetOr(math.MaxInt)
		}
		// remove i+1
		if i+2 < len(parts) {
			parts = append(parts[:i+1], parts[i+2:]...)
		} else {
			parts = parts[:i+1]
		}
	}

	tokens = make([]int, 0, len(parts)-1)
	for i := range parts[:len(parts)-1] {
		startIdx := parts[i].A
		endIdx := parts[i+1].A
		tokens = append(tokens, ranks[piece[startIdx:endIdx]])
	}
	return
}

func (e *Encoder) splitToken(token string) ([]string, error) {
	var matches []string

	m, err := e.pat.FindStringMatch(token)
	if err != nil {
		return nil, err
	}

	for m != nil {
		matches = append(matches, m.String())

		m, err = e.pat.FindNextMatch(m)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

/*
See: https://github.com/openai/tiktoken/blob/f19feecd071e22e02bf567ed12ccf161ce6db661/src/lib.rs#L14
*/

func (e *Encoder) Encode(text string) ([]int, error) {
	bpeTokens := []int{}

	matches, err := e.splitToken(text)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		bpeTokens = append(bpeTokens, bpe_merge(match, e.encoder)...)
	}

	return bpeTokens, nil
}

func (e *Encoder) Decode(tokens []int) string {
	var buf bytes.Buffer
	for _, tok := range tokens {
		buf.WriteString(e.decoder[tok])
	}
	return buf.String()
}
