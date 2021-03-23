package main

import "fmt"
import "log"

func logInfoln(a ...interface{}) {
	log.Println("[INFO] ", a)
}
func logInfof(format string, a ...interface{}) {
	log.Printf("[INFO] %s", fmt.Sprintf(format, a...))
}

func logDebugln(a ...interface{}) {
	log.Println("[DEBUG] ", a)
}

func logDebugf(format string, a ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(format, a...))
}
