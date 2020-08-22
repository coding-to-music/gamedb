package helpers

func StringsToInterfaces(s []string) (o []interface{}) {
	for _, v := range s {
		o = append(o, v)
	}
	return o
}

func IntsToInt32s(s []int) (o []int32) {
	for _, v := range s {
		o = append(o, int32(v))
	}
	return o
}
