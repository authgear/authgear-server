package loginid

type ReservedNameData []string

type ReservedNameChecker struct {
	data map[string]struct{}
}

func NewReservedNameChecker(data ReservedNameData) *ReservedNameChecker {
	wordMap := make(map[string]struct{})
	for _, word := range data {
		wordMap[word] = struct{}{}
	}
	return &ReservedNameChecker{
		data: wordMap,
	}
}

func (c *ReservedNameChecker) IsReserved(name string) bool {
	_, isReserved := c.data[name]
	return isReserved
}
