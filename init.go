package gpt3encoder

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type encoderConfig struct {
	pattern  string
	filename string
}

var modelMapping = map[string]string{
	// chat
	"gpt-4":         "cl100k_base",
	"gpt-3.5-turbo": "cl100k_base",
	// text
	"text-davinci-003": "p50k_base",
	"text-davinci-002": "p50k_base",
	"text-davinci-001": "r50k_base",
	"text-curie-001":   "r50k_base",
	"text-babbage-001": "r50k_base",
	"text-ada-001":     "r50k_base",
	"davinci":          "r50k_base",
	"curie":            "r50k_base",
	"babbage":          "r50k_base",
	"ada":              "r50k_base",
	// code
	"code-davinci-002": "p50k_base",
	"code-davinci-001": "p50k_base",
	"code-cushman-002": "p50k_base",
	"code-cushman-001": "p50k_base",
	"davinci-codex":    "p50k_base",
	"cushman-codex":    "p50k_base",
	// edit
	"text-davinci-edit-001": "p50k_edit",
	"code-davinci-edit-001": "p50k_edit",
	// embeddings
	"text-embedding-ada-002": "cl100k_base",
	// old embeddings
	"text-similarity-davinci-001":  "r50k_base",
	"text-similarity-curie-001":    "r50k_base",
	"text-similarity-babbage-001":  "r50k_base",
	"text-similarity-ada-001":      "r50k_base",
	"text-search-davinci-doc-001":  "r50k_base",
	"text-search-curie-doc-001":    "r50k_base",
	"text-search-babbage-doc-001":  "r50k_base",
	"text-search-ada-doc-001":      "r50k_base",
	"code-search-babbage-code-001": "r50k_base",
	"code-search-ada-code-001":     "r50k_base",
}

/*
https://github.com/openai/tiktoken/blob/f19feecd071e22e02bf567ed12ccf161ce6db661/tiktoken_ext/openai_public.py
*/
var configMapping = map[string]*encoderConfig{
	"cl100k_base": &encoderConfig{
		filename: "cl100k_base.tiktoken",
		pattern:  `/(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+/`,
	},
	"p50k_base": &encoderConfig{
		filename: "p50k_base.tiktoken",
		pattern:  `/'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+/`,
	},
	"r50k_base": &encoderConfig{
		filename: "r50k_base.tiktoken",
		pattern:  `/'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+/`,
	},
	"p50k_edit": &encoderConfig{
		filename: "p50k_edit.tiktoken",
		pattern:  `/'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+/`,
	},
}

var lock sync.RWMutex
var encoderMapping = make(map[string]*Encoder)

func NewEncoder(model string) *Encoder {
	underlying := modelMapping[model]
	if len(underlying) == 0 {
		return nil
	}

	lock.RLock()
	enc, ok := encoderMapping[underlying]
	lock.RUnlock()
	if ok {
		return enc
	}

	lock.Lock()
	defer lock.Unlock()
	enc, ok = encoderMapping[underlying]
	if ok {
		return enc
	}

	config, ok := configMapping[underlying]
	if !ok {
		return nil
	}
	enc, err := NewEncoderFromConfig(config)
	if err != nil {
		log.Errorf("new encoder for %s %s failed: %v", model, underlying, err)
	}
	encoderMapping[underlying] = enc
	return enc
}
