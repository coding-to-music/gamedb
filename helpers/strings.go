package helpers

func TruncateString(str string, size int) string {
	ret := str
	if len(str) > size {
		if size > 3 {
			size -= 3
		}
		ret = str[0:size] + "..."
	}
	return ret
}
