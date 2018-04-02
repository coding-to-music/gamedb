package helpers

func SliceHasInt(slice []int, i int) bool {
	for _, v := range slice {
		if v == i {
			return true
		}
	}
	return false
}

func Unique(arg []int) []int {

	tempMap := make(map[int]uint8)

	for idx := range arg {
		tempMap[arg[idx]] = 0
	}

	tempSlice := make([]int, 0)
	for key := range tempMap {
		tempSlice = append(tempSlice, key)
	}
	return tempSlice
}
