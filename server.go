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

type poolItem struct {
	responseCh chan time.Time
	location   Location
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

	poolCh := make(chan poolItem)
	workerPool(poolCh, 2)
	http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		srCh := make(chan time.Time)
		if request.Method == "POST" {
			if request.URL.Path == "/sunrise/at" {
				loc := getLocation(request)
				poolCh <- poolItem{responseCh: srCh, location: loc}

				sunrise := <-srCh

				writer.Write([]byte(sunrise.Format(time.RFC3339)))
			}
		} else {
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func workerPool(ch chan poolItem, size int) {
	for i := 0; i < size; i++ {
		go poolForFirstURL(ch)
		go poolForSecondURL(ch)
	}
}

func poolForFirstURL(ch chan poolItem) {
	for {
		item := <-ch
		tm, err := getSunriseFromFirstURL(item.location)
		if err == nil {
			item.responseCh <- tm
		}
	}
}

func poolForSecondURL(ch chan poolItem) {
	for {
		item := <-ch
		tm, err := getSunriseFromSecondURL(item.location)
		if err == nil {
			item.responseCh <- tm
		}
	}
}

func getSunriseFromFirstURL(location Location) (time.Time, error) {
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
			return sunriseTime.Time, nil
		}
	}

	return time.Time{}, err
}

func getSunriseFromSecondURL(location Location) (time.Time, error) {
	type SunriseTime struct {
		Time time.Time
	}
	body := bytes.NewBuffer([]byte("lat=" + strconv.FormatFloat(location.Lat, 'f', 10, 64) +
		"&lon=" + strconv.FormatFloat(location.Lon, 'f', 10, 64)))
	r, err := http.Post("http://sun.ri.se/at", "application/x-www-form-urlencoded", body)
	if err != nil {
		return time.Time{}, err
	}
	var sunriseTime SunriseTime
	b, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(b, &sunriseTime)
	if err != nil {
		log.Println("Error parsing sunrise time:", err)
		return time.Time{}, err
	}

	return sunriseTime.Time, nil
}

func getLocation(request *http.Request) Location {
	// assume this function gets Location passed via request
	return Location{}
}
