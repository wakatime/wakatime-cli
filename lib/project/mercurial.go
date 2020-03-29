package project

// Mercurial Information about the mercurial project for a given file.
type Mercurial struct {
	Name   *string
	Branch *string
}

// Process Process
func (s Mercurial) Process() bool {
	return false
}

// ProjectName ProjectName
func (s Mercurial) ProjectName() *string {
	return s.Name
}

// BranchName BranchName
func (s Mercurial) BranchName() *string {
	return s.Branch
}
