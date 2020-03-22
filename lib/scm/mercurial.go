package scm

// Mercurial Information about the mercurial project for a given file.
type Mercurial struct {
}

// Process Process
func (s Mercurial) Process() bool {
	return false
}

// ProjectName ProjectName
func (s Mercurial) ProjectName() *string {
	return nil
}

// BranchName BranchName
func (s Mercurial) BranchName() *string {
	return nil
}
