package linter

// Linter name constants.
const (
	LinterVersions    = "versions"
	LinterPermissions = "permissions"
	LinterFormat      = "format"
	LinterSecrets     = "secrets"
	LinterInjection   = "injection"
	LinterStyle       = "style"
)

// lintersWithAutoFix lists linters that support automatic fixing.
var lintersWithAutoFix = map[string]bool{
	LinterVersions: true,
	LinterFormat:   true,
}

// SupportsAutoFix returns true if the linter supports automatic fixing.
func SupportsAutoFix(linterName string) bool {
	return lintersWithAutoFix[linterName]
}
