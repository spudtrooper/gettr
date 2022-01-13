package log

import reallog "log"

func Printf(tmpl string, args ...interface{}) {
	reallog.Printf("[gettr] "+tmpl, args...)
}

func Println(s string) {
	reallog.Println("[gettr] " + s)
}
