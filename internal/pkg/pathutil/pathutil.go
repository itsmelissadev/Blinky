package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

func Normalize(path string) string {
	if path == "" {
		return ""
	}
	p := filepath.ToSlash(filepath.Clean(path))
	if os.PathSeparator == '\\' && len(p) == 2 && p[1] == ':' {
		return p + "/"
	}
	return p
}

func GetRoot() string {
	if os.PathSeparator == '\\' {
		return "C:/"
	}
	return "/"
}

func GetParent(path string) string {
	if path == "" {
		return ""
	}

	normalized := Normalize(path)

	if os.PathSeparator == '\\' {
		trimmed := strings.TrimSuffix(normalized, "/")
		if len(trimmed) == 2 && trimmed[1] == ':' {
			return ""
		}
	}

	parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(normalized)))

	if parent == "." || parent == normalized {

		if normalized == "/" || normalized == "" {
			return ""
		}

		if os.PathSeparator == '\\' && len(parent) <= 3 && strings.Contains(parent, ":") {
			return ""
		}
		return parent
	}

	return parent
}

func Join(base string, elem string) string {
	if base == "" {
		return elem
	}
	return filepath.ToSlash(filepath.Join(filepath.FromSlash(base), elem))
}

func FromSlash(path string) string {
	if path == "" {
		return "."
	}
	res := filepath.FromSlash(path)
	if os.PathSeparator == '\\' && len(res) == 2 && res[1] == ':' {
		return res + "\\"
	}
	return res
}

func ToSlash(path string) string {
	return filepath.ToSlash(path)
}
