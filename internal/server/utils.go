package server

import "fmt"

func getObjectType(i interface{}) string {
	return fmt.Sprintf("%T", i)
}

func gaugeType(i interface{}) bool {
	return getObjectType(i) == "float64"
}

func counterType(i interface{}) bool {
	return getObjectType(i) == "int64"
}

func contentEncodingContains(a []string, x string) bool {
	for _, s := range a {
		if s == x {
			return true
		}
	}

	return false
}
