package util

import "regexp"

func NamedStringSubMatch(r *regexp.Regexp, text string) map[string]string {

	result := map[string]string{}

	match := r.FindStringSubmatch(text)
	if match == nil {
		return result
	}

	for i, name := range r.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}

	return result
}
