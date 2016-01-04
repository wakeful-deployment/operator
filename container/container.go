package container

type Container struct {
	Name    string
	Image   string
	Ports   []string
	Env     map[string]string
	Restart string
	Tags    []string
}

func Diff(left []Container, right []Container) []Container {
	var result []Container

	for _, leftItem := range left {
		// Let's assume at first it is missing
		isMissing := true

		for _, rightItem := range right {
			if leftItem.Name == rightItem.Name {
				// If we find a match, then it's not missing
				isMissing = false
				break
			}
		}

		// If we found it to be missing in the end, then append to the diff
		if isMissing {
			result = append(result, leftItem)
		}
	}

	return result
}
