# go-gpt-3-encoder

Go BPE tokenizer (Encoder+Decoder) for openai models(not including GPT2)

## About

This code is inspired by and forked from [samber/go-gpt-3-encoder](https://github.com/samber/go-gpt-3-encoder)

but main logic is rewrite from [openai/tiktoken](https://github.com/openai/tiktoken)

encode logic see `_byte_pair_merge` in [lib.rs](https://github.com/openai/tiktoken/blob/main/src/lib.rs)

## Install

```bash
go get github.com/YEXINGZHE54/go-gpt-3-encoder
```

## Usage

```go
import tokenizer "github.com/YEXINGZHE54/go-gpt-3-encoder"

encoder, err := tokenizer.NewEncoder("gpt-3.5-turbo")
if err != nil {
    log.Fatal(err)
}

str := "This is an example sentence to try encoding out on!"

encoded, err := encoder.Encode(str)
if err != nil {
    log.Fatal(err)
}

fmt.Println("We can look at each token and what it represents:")
for _, token := range encoded {
    fmt.Printf("%d -- %s\n", token, encoder.Decode([]int{token}))
}

decoded := encoder.Decode(encoded)
fmt.Printf("We can decode it back into: %s\n", decoded)
```

## Contribute

Some corner cases are not covered by this library. See `@TODO` in tests.
