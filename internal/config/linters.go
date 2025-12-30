package config

// Linter name constants.
const (
	LinterVersions    = "versions"
	LinterPermissions = "permissions"
	LinterFormat      = "format"
	LinterSecrets     = "secrets"
	LinterInjection   = "injection"
	LinterStyle       = "style"
)

// allLinters lists all available linters.
var allLinters = []string{
	LinterVersions,
	LinterPermissions,
	LinterFormat,
	LinterSecrets,
	LinterInjection,
	LinterStyle,
}
