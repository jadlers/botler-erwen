package main

import "fmt"
import "log"

func logInfoln(a ...interface{}) {
	log.Printf("[INFO] %v\n", a...)
}
func logInfof(format string, a ...interface{}) {
	log.Printf("[INFO] %s", fmt.Sprintf(format, a...))
}

func logDebugln(a ...interface{}) {
	log.Printf("[DEBUG] %v\n", a...)
}

func logDebugf(format string, a ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(format, a...))
}
