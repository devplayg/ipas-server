package main

import (
	"github.com/icrowley/fake"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"fmt"
	"strings"
)

const (
	DefaultDateFormat = "2006-01-02 15:04:05"
)



func main() {
	t := time.Now()

	dt := t.Format(DefaultDateFormat)
	srcid := randTag(fake.CharactersN(2))
	lat := getLatitude("kr")
	lon := getLongitude("kr")
	spd := strconv.Itoa(fake.Year(-1,33))
	snr := strconv.Itoa(fake.Year(0,12))
	ctn := fake.Phone()
	sesid := fmt.Sprintf("%s_%s_1", srcid, t.Format("20060102150405"))

	// Status
	values := url.Values{
		"dt":    {dt},
		"srcid": {srcid},
		"lat":   {lat},
		"lon":   {lon},
		"spd":   {spd},
		"snr":   {snr},
		"ctn":   {ctn},
		"sesid": {sesid},
	}
	_, err := http.PostForm("http://127.0.0.1:8080/status", values)
	if err != nil {
		panic(err)
	}

	// Event
	values = url.Values{
		"dt":    {dt},
		"srcid": {srcid},
		"dstid": {getDstid()},
		"lat":   {lat},
		"lon":   {lon},
		"spd":   {spd},
		"snr":   {snr},
		"ctn":   {ctn},
		"type":  {strconv.Itoa(fake.Year(0, 3))},
		"dist":  {fake.DigitsN(1)},
		"sesid": {sesid},
	}

	_, err = http.PostForm("http://127.0.0.1:8080/event", values)
	if err != nil {
		panic(err)
	}
}

func getDstid() string{
	count := NumberRange(1, 3)

	arr := make([]string, 0)
	for i:= 0; i<count; i++ {
		arr = append(arr,randTag(fake.CharactersN(2)))
	}

	return strings.Join(arr, ",")
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
	prefix += name
	return prefix
	//prefix += name + "_"
	//return prefix + fake.DigitsN(2)
}

func NumberRange(from, to int) int {
	return fake.Year(from-1, to)
}


func getLatitude(loc string) string {
	i := NumberRange(35, 37)
	d := NumberRange(100, 99999)

	return fmt.Sprintf("%d.%d", i, d)
}

func getLongitude(loc string) string {
	i := NumberRange(127, 128)
	d := NumberRange(100, 99999)

	return fmt.Sprintf("%d.%d", i, d)
}

func float32ToString(f32 float32) string {
	return strconv.FormatFloat(float64(f32), 'f', 6, 64)
}