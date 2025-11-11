package model

import "github.com/google/uuid"

const EmptyTitle string = ""

type MovieMeta struct {
	ID         uuid.UUID
	PosterLink string
	Title      string
	Genres     []string
	Year       int
	Rating     float64

	Overview string
}

type Poster struct {
	Filename string
	Content  []byte

	MovieID string
}

func (r Poster) GetFilename() string {
	return r.Filename
}

func (r Poster) GetContent() []byte {
	return r.Content
}

func (r Poster) GetParent() string {
	return r.MovieID
}

func (r *Poster) NewFromData(content []byte, name string) FileObject {
	return &Poster{
		Content:  content,
		Filename: name,
	}
}

type Movie struct {
	MM     *MovieMeta
	Poster *Poster
}
