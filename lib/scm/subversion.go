package scm

// Mercurial Information about the mercurial project for a given file.
type Subversion struct {
}

// Process Process
func (s Subversion) Process() bool {
	return false
}

// ProjectName ProjectName
func (s Subversion) ProjectName() *string {
	return nil
}

// BranchName BranchName
func (s Subversion) BranchName() *string {
	return nil
}
