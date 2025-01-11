package errors

import "dt/utils/log"

func Error(e error) {
	log.Error(e.Error())
}

func ErrorCheck(e error) {
	if e != nil {
		Error(e)
	}
}
