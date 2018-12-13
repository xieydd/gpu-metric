package util

import (
	"github.com/golang/glog"
	"time"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"strings"
)

//Must is candy func
func Must(err error) {
	if err != nil {
		glog.Fatalln(err)
	}
}

func MustE(err error) {
	Must(err)
	if(err != nil) {
		os.Exit(0)
	}
}

// Return readeable Duration
func ShortHumanDuration(d time.Duration) string{
	if seconds := int(d.Seconds()); seconds < -1 {
		fmt.Sprintf("<invalid>")
	}else if seconds < 0 {
		return fmt.Sprintf("0s")
	}else if seconds < 60 {
		return fmt.Sprintf("%ds",seconds)
	}else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm",minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh",hours)
	}else if hours < 24*365 {
		return fmt.Sprintf("%dd",hours/24)
	}
	return fmt.Sprintf("%dy",int(d.Hours()/24/365))
}

// SetLogLevel sets the logrus logging level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("Unknown level: %s", level)
	}
}
