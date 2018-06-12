package algoutil

import "strconv"

func RetriveIntOrDefault(s string, d int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return d
	}

	return i
}

func RetriveInt64OrDefault(s string, d int64) int64 {
	i, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return d
	}

	return i
}
