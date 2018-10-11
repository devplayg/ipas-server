package main

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"time"
	"github.com/icrowley/fake"
	"github.com/devplayg/ipas-server"
	"encoding/json"
	"fmt"
)

func main() {
	url := "http://127.0.0.1:8000/event"
	jsonData := getJsonData()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	//fmt.Println("response Body:", string(body))
}

func getJsonData() []byte {
	t := time.Now().Add(time.Duration(fake.Year(1, 3600)) * time.Second * -1)
	data := map[string]interface{} {
		"dt": t.Add(time.Duration(fake.Year(1, 360)) * time.Second).Format(ipasserver.DateDefault),
		"cstid": "won",
		"spd": 3434,
	}
	jsonData, _ := json.Marshal(data)
	return jsonData
}