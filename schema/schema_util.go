package schema

func FindFirstSpecifiedTag(tags *[][]TagElem, tag string) *[]TagElem {
	for _, t := range *tags {
		if string(t[0]) == tag {
			tmpVal := t
			return &tmpVal
		}
	}
	return nil
}
