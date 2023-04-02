package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func QueryInt(ctx *gin.Context, key string, def int, minMax ...int) int {
	v := ctx.Query(key)
	if v == "" {
		return def
	}

	vi, err := strconv.Atoi(v)
	if err != nil {
		return def
	}

	if len(minMax) > 0 && vi < minMax[0] {
		vi = minMax[0]
	}
	if len(minMax) > 1 && vi > minMax[1] {
		vi = minMax[1]
	}

	return vi
}

func DoPaged[T any](s []T, ps int, f func([]T) error) (err error) {
	pages := len(s) / ps
	if len(s)%ps > 0 {
		pages++
	}

	for i := 0; i < pages; i++ {
		c := s[i*ps:]
		if i < pages-1 {
			c = c[:ps]
		}
		if err = f(c); err != nil {
			return err
		}
	}

	return nil
}
