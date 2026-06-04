package cache

import "fmt"

func getKey(k string, t ValueType) string {
	return fmt.Sprintf("%s:%s", t, k)
}

type ValueType = string

const (
	CarsValueType ValueType = "cars"
)
