package project

// Subversion Information about the mercurial project for a given file.
type Subversion struct {
	Name   *string
	Branch *string
}

// Process Process
func (p Subversion) Process() bool {
	return false
}

// SetEntity SetEntity
func (p Subversion) SetEntity(entity string) {

}

// SetConfigItems SetConfigItems
func (p Subversion) SetConfigItems(ci map[string]string) {

}

// ProjectName ProjectName
func (p Subversion) ProjectName() *string {
	return p.Name
}

// BranchName BranchName
func (p Subversion) BranchName() *string {
	return p.Branch
}
