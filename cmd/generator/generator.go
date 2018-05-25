package main

import (
	"fmt"
	"github.com/devplayg/ipas-server"
	"github.com/icrowley/fake"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	AppName    = "IPAS Data Generator"
	AppVersion = "2.0.1804.32201"
)

var companies = []string{"us1", "us2", "kr1", "kr2", "jp1", "jp2"}

func main() {

	// 옵션 설정
	var (
		version  = ipasserver.CmdFlags.Bool("version", false, "Version")
		count    = ipasserver.CmdFlags.Int("count", 10, "Event count")
		addr     = ipasserver.CmdFlags.String("addr", "127.0.0.1:8080", "Address")
		loop     = ipasserver.CmdFlags.Bool("loop", false, "Loop")
		interval = ipasserver.CmdFlags.Duration("interval", 10*time.Second, "Interval")
	)
	ipasserver.CmdFlags.Usage = ipasserver.PrintHelp
	ipasserver.CmdFlags.Parse(os.Args[1:])

	// 버전 출력
	if *version {
		ipasserver.DisplayVersion(AppName, AppVersion)
		return
	}

	if *loop {
		fmt.Printf("loop=true, interval=%3.0fs\n", (*interval).Seconds())
	}
	start := time.Now()
	for {
		for i := 0; i < *count; i++ {
			t := time.Now().Add(time.Duration(fake.Year(1, 3600)) * time.Second * -1)
			dt := t.Format(ipasserver.DateDefault)
			cstid := getRandomOrgCode()
			srcid := getRandomTag(cstid)
			lat := getLatitude("kr")
			lon := getLongitude("kr")
			spd := strconv.Itoa(fake.Year(-1, 33))
			snr := strconv.Itoa(fake.Year(0, 12))
			ctn := fake.Phone()
			sesid := fmt.Sprintf("%s_%s", srcid, t.Format("20060102150405"))

			// Status
			values := url.Values{
				"dt":    {dt},
				"cstid": {cstid},
				"srcid": {srcid},
				"lat":   {lat},
				"lon":   {lon},
				"spd":   {spd},
				"snr":   {snr},
				"ctn":   {ctn},
				"sesid": {sesid},
			}
			_, err := http.PostForm("http://"+*addr+"/status", values)
			if err != nil {
				panic(err)
			}

			// Event
			eventType := NumberRange(1, 4)
			if eventType == 3 {
				spd = strconv.Itoa(NumberRange(13, 33))
			}
			values = url.Values{
				"dt":    {dt},
				"cstid": {cstid},
				"srcid": {srcid},
				"dstid": {getRandomTag(cstid)},
				"lat":   {lat},
				"lon":   {lon},
				"spd":   {spd},
				"snr":   {snr},
				"ctn":   {ctn},
				"type":  {strconv.Itoa(eventType)},
				"dist":  {fake.DigitsN(1)},
				"sesid": {sesid},
			}

			_, err = http.PostForm("http://"+*addr+"/event", values)
			if err != nil {
				panic(err)
			}
		}
		dur := time.Since(start).Seconds()
		fmt.Printf("count=%d, time=%3.2fs, EPS=%3.1f\n", *count*2, dur, (float64(*count) * 2 / dur))
		if !*loop {
			break
		}

		time.Sleep(*interval)
	}
}

func getRandomOrgCode() string {
	c := len(companies)
	return companies[NumberRange(1, c)-1]
}

func getDstid(orgid string) string {
	count := NumberRange(1, 3)

	arr := make([]string, 0)
	for i := 0; i < count; i++ {
		arr = append(arr, getRandomTag(orgid))
	}

	return strings.Join(arr, ",")
}

func getRandomTag(orgid string) string {
	tagType := NumberRange(1, 3)
	prefix := ""

	if tagType == 1 {
		prefix = "VT_" + orgid + "_"
	} else if tagType == 2 {
		prefix = "ZT_" + orgid + "_"
	} else if tagType == 3 {
		prefix = "PT_" + orgid + "_"
	} else {
		log.Fatal("invalid tag type")
	}
	prefix += fake.DigitsN(1)

	return prefix
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

//
//func float32ToString(f32 float32) string {
//	return strconv.FormatFloat(float64(f32), 'f', 6, 64)
//}
