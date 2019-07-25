package helpers

func FlattenMap(m map[string]interface{}) map[string]interface{} {

	o := make(map[string]interface{})
	for k, v := range m {

		switch child := v.(type) {
		case map[string]interface{}:
			nm := FlattenMap(child)
			for nk, nv := range nm {
				o[k+" / "+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}
