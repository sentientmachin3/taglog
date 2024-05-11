package internal

func MaxStringLength(list []string) int {
	maxLength := len(list[0])
	for _, str := range list {
		if len(str) > maxLength {
			maxLength = len(str)
		}
	}
	return maxLength
}
