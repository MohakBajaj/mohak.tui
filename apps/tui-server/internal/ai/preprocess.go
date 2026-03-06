package ai

import "strings"

// PreprocessMessage normalizes common shorthand and one-word portfolio queries.
func PreprocessMessage(message string) string {
	processed := strings.TrimSpace(message)
	replacer := strings.NewReplacer(
		" u ", " you ",
		" ur ", " your ",
		" pls ", " please ",
		" thx ", " thanks ",
	)
	processed = replacer.Replace(" " + processed + " ")
	processed = strings.TrimSpace(processed)

	if len(strings.Fields(processed)) == 1 {
		word := strings.ToLower(processed)
		switch word {
		case "skills", "experience", "projects", "contact", "education":
			processed = "Tell me about Mohak's " + word
		}
	}

	return processed
}

func toLower(value string) string {
	return strings.ToLower(value)
}
