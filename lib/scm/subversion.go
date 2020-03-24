package scm

// Subversion Information about the mercurial project for a given file.
type Subversion struct {
	Name   *string
	Branch *string
}

// Process Process
func (s Subversion) Process() bool {
	return false
}

// ProjectName ProjectName
func (s Subversion) ProjectName() *string {
	return s.Name
}

// BranchName BranchName
func (s Subversion) BranchName() *string {
	return s.Branch
}
