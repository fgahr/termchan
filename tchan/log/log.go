package log

import "log"

// Error adds an error description to the application log.
func Error(err error) {
	log.Println(err)
}
