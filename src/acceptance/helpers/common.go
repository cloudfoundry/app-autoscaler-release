package helpers

import "strings"

func getNames(resources []cfResource) []string {
	var names []string
	for _, item := range resources {
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


func filterByState(apps []cfResource, state string) []string {
	var filtered []string

	for _, item := range apps {
		if item.State == state {
			filtered = append(filtered, item.Name)
		}
	}

	return filtered

}