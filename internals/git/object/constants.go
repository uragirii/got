package object

type ObjectType string

const _GitBlobHeader string = "blob %d\u0000"

const (
	BlobObj   ObjectType = "blob"
	TreeObj   ObjectType = "tree"
	CommitObj ObjectType = "commit"
)
