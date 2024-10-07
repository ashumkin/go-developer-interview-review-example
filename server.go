package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Location struct {
	Lat float64
	Lon float64
}

func main() {
	var addr string
	if os.Args[1] == "-listen.addr" {
		if addr = os.Args[2]; addr == "" {
			log.Fatal("You must specify an address to listen on.")
		}
	} else {
		log.Fatal("You must specify an address to listen on.")
	}

	http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		srCh := make(chan time.Time)
		if request.Method == "POST" {
			if request.URL.Path == "/sunrise/at" {
				loc := getLocation(request)
				go getSunriseFromFirstURL(srCh, loc)
				go getSunriseFromSecondURL(srCh, loc)

				sunrise := <-srCh

				writer.Write([]byte(sunrise.Format(time.RFC3339)))
			}
		} else {
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func getSunriseFromFirstURL(ch chan time.Time, location Location) {
	type SunriseIOTime struct {
		Time time.Time
	}
	r, err := http.Get("http://sunrise.sunrise.io/?location=" + strconv.FormatFloat(location.Lat, 'f', 10, 64) +
		"," + strconv.FormatFloat(location.Lon, 'f', 10, 64))
	if err == nil {
		var sunriseTime SunriseIOTime
		b, _ := io.ReadAll(r.Body)
		err = json.Unmarshal(b, &sunriseTime)
		if err == nil {
			ch <- sunriseTime.Time
		}
	}
}

func getSunriseFromSecondURL(ch chan time.Time, location Location) {
	type SunriseTime struct {
		Time time.Time
	}
	body := bytes.NewBuffer([]byte("lat=" + strconv.FormatFloat(location.Lat, 'f', 10, 64) +
		"&lon=" + strconv.FormatFloat(location.Lon, 'f', 10, 64)))
	r, err := http.Post("http://sun.ri.se/at", "application/x-www-form-urlencoded", body)
	if err != nil {
		return
	}
	var sunriseTime SunriseTime
	b, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(b, &sunriseTime)
	if err != nil {
		log.Println("Error parsing sunrise time:", err)
		return
	}
	ch <- sunriseTime.Time
}

func getLocation(request *http.Request) Location {
	// assume this function gets Location passed via request
	return Location{}
}
