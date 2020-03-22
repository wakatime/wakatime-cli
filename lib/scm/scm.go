package scm

// Scm Scm interface
type Scm interface {
	Process() bool
	ProjectName() *string
	BranchName() *string
}
