package logs

import (
	"fmt"
	"log"
	"os"
	"time"
)

func Info(msg string, args ...interface{}) {
	output := fmt.Sprintf("[INFO] "+msg, args...)
	log.Output(2, output)
}

func Debug(msg string, args ...interface{}) {
	output := fmt.Sprintf("[DEBUG] "+msg, args...)
	log.Output(2, output)
}

func Warn(msg string, args ...interface{}) {
	output := fmt.Sprintf("[WARN] "+msg, args...)
	log.Output(2, output)
}

func Error(msg string, args ...interface{}) {
	output := fmt.Sprintf("[ERROR] "+msg, args...)
	log.Output(2, output)
}
func Critical(msg string, args ...interface{}) {
	output := fmt.Sprintf("[CRITICAL] "+msg, args...)
	log.Output(2, output)
	os.Exit(1)
}
func TimeOut(msg string, args ...interface{}) {
	output := fmt.Sprintf("[TIMEOUT] "+msg, args...)
	log.Output(2, output)
}

func Success(msg string, args ...interface{}) {
	output := fmt.Sprintf("[SUCCESS] "+msg, args...)
	log.Output(2, output)
}
func Router(msg string, args ...interface{}) {
	output := fmt.Sprintf("[ROUTER] "+msg, args...)
	log.Output(2, output)
}

func Trace(msg string, args ...interface{}) {
	output := fmt.Sprintf("[TRACE] "+msg, args...)
	log.Output(2, output)
}

func TrackWithId(name string, cycle uint64, gid uint64) func() {

	start := time.Now()

	Trace("INICIO [cycle: %d | gid: %d] - %s", cycle, gid, name)

	return func() {
		elapsed := time.Since(start)
		Trace("FIM [cycle: %d | gid: %d] - %s | duração: %s", cycle, gid, name, elapsed)
	}
}
func Track(name string) func() {

	start := time.Now()

	Trace("INICIO - %s", name)

	return func() {
		elapsed := time.Since(start)
		Trace("FIM - %s | duração: %s", name, elapsed)
	}
}
