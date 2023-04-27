package gpt3encoder

import (
	"bytes"
	"embed"
	"fmt"

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

type Encoder struct {
	pat     *regexp2.Regexp
	encoder map[string]int
	decoder map[int]string
}

func NewEncoderFromConfig(conf *encoderConfig) (*Encoder, error) {
	encoder, err := loadTokenBPE(conf.filename)
	if err != nil {
		err = fmt.Errorf("gpt3encoder: failed to load token file: %v", err)
		return nil, err
	}
	decoder := lo.Invert(encoder)

	enc := Encoder{
		pat:     regexp2.MustCompile(conf.pattern, 0),
		encoder: encoder,
		decoder: decoder,
	}

	return &enc, nil
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
