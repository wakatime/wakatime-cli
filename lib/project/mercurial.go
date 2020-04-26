package project

// Mercurial Information about the mercurial project for a given file.
type Mercurial struct {
	Name   *string
	Branch *string
}

// Process Process
func (p Mercurial) Process() bool {
	return false
}

// SetEntity SetEntity
func (p Mercurial) SetEntity(entity string) {

}

// SetConfigItems SetConfigItems
func (p Mercurial) SetConfigItems(ci map[string]string) {

}

// ProjectName ProjectName
func (p Mercurial) ProjectName() *string {
	return p.Name
}

// BranchName BranchName
func (p Mercurial) BranchName() *string {
	return p.Branch
}
