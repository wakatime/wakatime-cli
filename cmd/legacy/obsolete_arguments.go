package legacy

// ObsoleteArguments ObsoleteArguments
type ObsoleteArguments struct {
	File           string //file
	HideFilenames1 bool   //hide-filenames
	HideFilenames2 bool   //hidefilenames
	LogFile        string //log-file
	APIURL         string //apiurl
}

// NewObsoleteArgs NewObsoleteArgs
func NewObsoleteArgs() *ObsoleteArguments {
	return &ObsoleteArguments{}
}
