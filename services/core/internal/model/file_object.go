package model

type FileObject interface {
	GetFilename() string
	GetParent() string
	GetContent() []byte
}

type FileObjectBuilder interface {
	NewFromData(content []byte, filename string) FileObject
}
