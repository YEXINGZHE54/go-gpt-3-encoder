package gpt3encoder

import (
	"math"

	"github.com/samber/lo"
)

type (
	BytePair struct {
		startPos int
		rank     int
	}
)

/*
See: https://github.com/openai/tiktoken/blob/f19feecd071e22e02bf567ed12ccf161ce6db661/src/lib.rs#L14
*/
func bpe_merge(piece string, ranks map[string]int) (tokens []int) {
	//init parts, each item is {startPosInPiece, rank}
	parts := make([]BytePair, len(piece)+1)
	for i := range parts {
		parts[i].startPos = i
		parts[i].rank = math.MaxInt
	}

	//calc rank func
	get_rank := func(start int, skip int) int {
		if start+skip+2 < len(parts) {
			startIdx := parts[start].startPos
			endIdx := parts[start+skip+2].startPos
			r, ok := ranks[piece[startIdx:endIdx]]
			if !ok {
				return math.MaxInt
			}
			return r
		} else {
			return math.MaxInt
		}
	}

	// calc rank once
	for i := 0; i < len(parts)-2; i++ {
		res := get_rank(i, 0)
		if res != math.MaxInt {
			parts[i].rank = res
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
			if v.rank < minRank.B {
				minRank.A = i
				minRank.B = v.rank
			}
		}
		if minRank.B == math.MaxInt { //not found?
			break
		}
		i := minRank.A
		//we found a min part, we need to update rank and remove parts[i+1]
		parts[i].rank = get_rank(i, 1)
		if i > 0 {
			parts[i-1].rank = get_rank(i-1, 1)
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
		startIdx := parts[i].startPos
		endIdx := parts[i+1].startPos
		tokens = append(tokens, ranks[piece[startIdx:endIdx]])
	}
	return
}
