package main

import (
	"net/http"
	"time"
	"net/url"
	"fmt"
)

func main() {

	encodedDate := url.QueryEscape(time.Now().Format("2006-01-04 15:04:05"))
	u := "http://127.0.0.1:8080/status?dt=" + encodedDate + "&srcid=VTSAMPLE&lat=126.886559&lon=37.480888&spd=1234.1"
	fmt.Println(u)
	resp, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	u = "http://127.0.0.1:8080/event?dt=" + encodedDate + "&srcid=VTSAMPLE&lat=126.886559&lon=37.480888&spd=1234.1"
	fmt.Println(u)
	resp, err = http.Get(u)
	if err != nil {
		panic(err)
	}
}
