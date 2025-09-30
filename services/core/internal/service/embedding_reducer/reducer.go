package embedding_reducer

import "github.com/humanbelnik/kinoswap/core/internal/model"

type EmbeddingReducer struct{}

/*
Make this function panic only in learning purposes
*/
func (r *EmbeddingReducer) Reduce(embs []*model.Embedding) model.Embedding {
	if embs == nil || *embs[0] == nil {
		panic("no data")
	}

	for i := 1; i < len(embs); i++ {
		if len(*embs[i-1]) != len(*embs[i]) {
			panic("not equal vectors")
		}
	}

	e := make([]float32, len(*embs[0]))
	n := float32(len(embs))
	for i := range len(*embs[0]) {
		for k := range len(embs) {
			e[i] += (*embs[k])[i]
		}
		e[i] /= n
	}

	return e
}
