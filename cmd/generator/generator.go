package main

import (
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultDateFormat = "2006-01-02 15:04:05"
)

func main() {
	t := time.Now()

	// Status
	values := url.Values{
		"dt":    {t.Format(DefaultDateFormat)},
		"srcid": {"VT-SAMPLE"},
		"lat":   {"126.886633"},
		"lon":   {"38.1488088"},
		"spd":   {"34.1"},
	}
	u := "http://127.0.0.1:8080/status?" + values.Encode()
	_, err := http.Get(u)
	if err != nil {
		panic(err)
	}

	// Event
	values = url.Values{
		"dt":      {t.Format(DefaultDateFormat)},
		"target":  {"PT-SAMPLE,ZT-SAMPLE"},
		"wardist": {"1"},
		"caudist": {"3"},
		"v2vdist": {"5"},
	}
	_, err = http.PostForm("http://127.0.0.1:8080/event", values)
	if err != nil {
		panic(err)
	}
}
