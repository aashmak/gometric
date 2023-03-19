package server

import "fmt"

func getObjectType(i interface{}) string {
	return fmt.Sprintf("%T", i)
}

func gaugeType(i interface{}) bool {
	if getObjectType(i) == "server.gauge" {
		return true
	}
	return false
}

func counterType(i interface{}) bool {
	if getObjectType(i) == "server.counter" {
		return true
	}
	return false
}
