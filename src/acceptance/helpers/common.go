package helpers

import "strings"

func getNames(resources cfResourceObject) []string {
	var names []string
	for _, item := range resources.Resources {
		names = append(names, item.Name)
	}
	return names
}

func filterByPrefix(prefix string, in []string) []string {
	var filtered []string
	for _, item := range in {
		if strings.HasPrefix(item, prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
