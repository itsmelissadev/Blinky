package postgresql

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type ConfField struct {
	Name        string   `json:"name"`
	Value       string   `json:"value"`
	IsCommented bool     `json:"is_commented"`
	Description string   `json:"description"`
	Options     []string `json:"options,omitempty"`
}

type ConfSection struct {
	Title  string      `json:"title"`
	Fields []ConfField `json:"fields"`
}

var (
	sectionRegex = regexp.MustCompile(`^#\s*--+`)
	settingRegex = regexp.MustCompile(`^(#)?\s*([a-zA-Z0-9_]+)\s*=\s*([^#\n]*)?(#.*)?$`)
	optionsRegex = regexp.MustCompile(`([a-zA-Z0-9_]+(\s*,\s*[a-zA-Z0-9_]+)+\s*or\s*[a-zA-Z0-9_]+)|([a-zA-Z0-9_]+\s*or\s*[a-zA-Z0-9_]+)`)
)

func ParseConf(content string) []ConfSection {
	var sections []ConfSection
	var currentSection *ConfSection

	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if sectionRegex.MatchString(trimmed) {
			if scanner.Scan() {
				titleLine := scanner.Text()
				title := strings.Trim(titleLine, "# ")
				if title != "" && !strings.Contains(title, "---") {
					if currentSection != nil && len(currentSection.Fields) > 0 {
						sections = append(sections, *currentSection)
					}
					currentSection = &ConfSection{Title: title, Fields: []ConfField{}}
				}
			}
			continue
		}

		matches := settingRegex.FindStringSubmatch(trimmed)
		if matches != nil {
			isCommented := matches[1] == "#"
			name := matches[2]
			value := strings.TrimSpace(matches[3])
			value = strings.Trim(value, "'")
			description := ""
			var options []string

			if matches[4] != "" {
				description = strings.TrimSpace(strings.TrimPrefix(matches[4], "#"))
				if strings.Contains(description, " or ") {
					cleanDesc := strings.ReplaceAll(description, " (change requires restart)", "")
					cleanDesc = strings.ReplaceAll(cleanDesc, "(change requires restart)", "")
					optMatch := optionsRegex.FindString(cleanDesc)
					if optMatch != "" {
						normalized := strings.ReplaceAll(optMatch, " or ", ",")
						parts := strings.Split(normalized, ",")
						for _, p := range parts {
							p = strings.TrimSpace(p)
							if p != "" {
								options = append(options, p)
							}
						}
					}
				}
				if strings.Contains(strings.ToLower(description), "on, off") && len(options) == 0 {
					options = []string{"on", "off"}
				}
			}

			if currentSection == nil {
				currentSection = &ConfSection{Title: "General", Fields: []ConfField{}}
			}

			currentSection.Fields = append(currentSection.Fields, ConfField{
				Name:        name,
				Value:       value,
				IsCommented: isCommented,
				Description: description,
				Options:     options,
			})
		}
	}

	if currentSection != nil && len(currentSection.Fields) > 0 {
		sections = append(sections, *currentSection)
	}

	return sections
}

func GenerateConf(sections []ConfSection) string {
	var sb strings.Builder
	sb.WriteString("# -----------------------------\n")
	sb.WriteString("# PostgreSQL configuration file\n")
	sb.WriteString("# -----------------------------\n\n")

	for _, section := range sections {
		if section.Title != "" {
			sb.WriteString(fmt.Sprintf("\n# %s\n", section.Title))
			sb.WriteString("# " + strings.Repeat("-", len(section.Title)+2) + "\n")
		}

		for _, field := range section.Fields {
			prefix := ""
			if field.IsCommented {
				prefix = "#"
			}
			value := field.Value

			if strings.Contains(value, " ") || strings.Contains(value, ",") || value == "" {
				value = "'" + value + "'"
			}

			line := fmt.Sprintf("%s%s = %s", prefix, field.Name, value)
			if field.Description != "" {
				line = fmt.Sprintf("%-40s # %s", line, field.Description)
			}
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}
