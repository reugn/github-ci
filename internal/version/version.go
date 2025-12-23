package version

import (
	"strconv"
	"strings"
)

// Normalize normalizes a version string by removing leading/trailing whitespace
// and the "v" prefix if present (e.g., "v1.2.3" becomes "1.2.3").
func Normalize(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}

// ExtractMajor extracts the major version number from a version string.
// For example, "v1.2.3" returns 1.
func ExtractMajor(v string) int {
	v = Normalize(v)
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		major, _ := strconv.Atoi(parts[0])
		return major
	}
	return 0
}

// ExtractMajorMinor extracts the major and minor version numbers from a version string.
// For example, "v1.2.3" returns (1, 2).
func ExtractMajorMinor(v string) (int, int) {
	v = Normalize(v)
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		major, _ := strconv.Atoi(parts[0])
		minor, _ := strconv.Atoi(parts[1])
		return major, minor
	}

	if len(parts) == 1 {
		major, _ := strconv.Atoi(parts[0])
		return major, 0
	}

	return 0, 0
}

// ToMajorTag converts a version string to just the major version tag.
// For example, "v5.2.0" returns "v5", "5.2.0" returns "v5".
func ToMajorTag(v string) string {
	major := ExtractMajor(v)
	return "v" + strconv.Itoa(major)
}

// Compare compares two semantic version strings.
// Returns -1 if v1 < v2, 0 if v1 == v2, or 1 if v1 > v2.
// Handles versions with different numbers of components by treating missing components as 0.
func Compare(v1, v2 string) int {
	v1 = Normalize(v1)
	v2 = Normalize(v2)

	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := max(len(v1Parts), len(v2Parts))

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int
		if i < len(v1Parts) {
			v1Num, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			v2Num, _ = strconv.Atoi(v2Parts[i])
		}

		if v1Num < v2Num {
			return -1
		}
		if v1Num > v2Num {
			return 1
		}
	}

	return 0
}
