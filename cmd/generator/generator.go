package main

import (
	"github.com/icrowley/fake"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultDateFormat = "2006-01-02 15:04:05"
)

func float32ToString(f32 float32) string {
	return strconv.FormatFloat(float64(f32), 'f', 6, 64)
}

func main() {
	t := time.Now()

	dt := t.Format(DefaultDateFormat)
	srcid := randTag(fake.CharactersN(2))
	lat := float32ToString(fake.Latitude())
	lon := float32ToString(fake.Longitude())
	spd := fake.DigitsN(2)
	snr := fake.DigitsN(1)
	ctn := fake.Phone()

	// Status
	values := url.Values{
		"dt":    {dt},
		"srcid": {srcid},
		"lat":   {lat},
		"lon":   {lon},
		"spd":   {spd},
		"snr":   {snr},
		"ctn":   {ctn},
	}
	_, err := http.PostForm("http://127.0.0.1:8080/status", values)
	if err != nil {
		panic(err)
	}

	// Event
	values = url.Values{
		"dt":    {dt},
		"srcid": {srcid},
		"dstid": {randTag(fake.CharactersN(2)) + "," + randTag(fake.CharactersN(2))},
		"lat":   {lat},
		"lon":   {lon},
		"spd":   {spd},
		"snr":   {snr},
		"ctn":   {ctn},
		"type":  {strconv.Itoa(fake.Year(1, 2))},
		"dist":  {fake.DigitsN(1)},
	}

	_, err = http.PostForm("http://127.0.0.1:8080/event", values)
	if err != nil {
		panic(err)
	}
}

func randTag(name string) string {
	tagType := NumberRange(1, 3)
	prefix := ""

	if tagType == 1 {
		prefix = "VT_"
	} else if tagType == 2 {
		prefix = "ZT_"
	} else if tagType == 3 {
		prefix = "PT_"
	}
	prefix += name + "_"
	//prefix += idPools[NumberRange(0, len(idPools)-1)].(string)
	return prefix + fake.DigitsN(2)
}

func NumberRange(from, to int) int {
	return fake.Year(from-1, to)
}
