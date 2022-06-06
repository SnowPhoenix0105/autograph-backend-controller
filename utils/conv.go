package utils

import "strconv"

func MustAtoi(integer string) int {
	ret, err := strconv.Atoi(integer)
	if err != nil {
		panic(err)
	}

	return ret
}
