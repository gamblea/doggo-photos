package main

// Photo structure for working with uploading photos
type Photo struct {
	User   string
	ID     string
	Data   []byte
	Width  int
	Height int
}
