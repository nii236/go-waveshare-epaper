package epd

import "log"

func MustRead(val int, err error) int {
	if err != nil {
		log.Println(err)
		return 0
	}
	return val
}
