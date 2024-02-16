package util

func isMap(a any) bool {
	_, isMap := a.(map[string]any)
	return isMap
}

// JsonMerge performs a JSON Merge Patch as defined in RFC7386.
//
// The operation happens *in place* that means that current is updated.
//
// Notes:
//
// - if new is nil, the output is an empty object - this allows for in-place
//
//   - if a key is not a map, it will be treated as scalar according to the
//     JSON Merge Patch strategy. This includes structs and slices.
func JsonMerge(current, new map[string]any) map[string]any {
	if new == nil {
		return nil
	}
	for k := range new {
		if _, exists := current[k]; exists {
			if new[k] != nil {
				if isMap(current[k]) && isMap(new[k]) {
					current[k] = JsonMerge(current[k].(map[string]any), new[k].(map[string]any))
				} else {
					current[k] = new[k]
				}
			} else {
				delete(current, k)
			}
		} else {
			current[k] = new[k]
		}
	}
	return current
}
