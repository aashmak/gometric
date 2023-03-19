package server

import "fmt"

func getObjectType(i interface{}) string {
	return fmt.Sprintf("%T", i)
}

func gaugeType(i interface{}) bool {
	return getObjectType(i) == "server.gauge"
}

func counterType(i interface{}) bool {
	return getObjectType(i) == "server.counter"
}
