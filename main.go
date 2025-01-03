package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	ticker := time.NewTicker(300 * time.Millisecond)

	client := resty.New()

	errsCount := 0

	for range ticker.C {
		if err := fetchMetrics(client); err != nil {
			errsCount++
			if errsCount >= 3 {
				log.Println(err)
			}
		}
	}
}

func fetchMetrics(client *resty.Client) error {
	resp, err := client.R().
		Get("http://srv.msk01.gigacorp.local")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	body := resp.String()
	parts := strings.Split(body, ",")

	var numbers []int64
	for _, part := range parts {
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		numbers = append(numbers, num)
	}

	if len(numbers) != 7 {
		return fmt.Errorf("unexpected number of numbers: %d", len(numbers))
	}

	loadAverage := numbers[0]
	if loadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", loadAverage)
	}

	totalMemBytes := numbers[1]
	usedMemBytes := numbers[2]
	memUsage := (totalMemBytes - usedMemBytes) / totalMemBytes * 100
	if memUsage > 80 {
		fmt.Printf("Memory usage too high: %d\n", memUsage)
	}

	totalDiskBytes := numbers[3]
	usedDiskBytes := numbers[4]
	diskUsage := (totalDiskBytes - usedDiskBytes) / totalDiskBytes * 100
	if diskUsage > 90 {
		fmt.Printf("Free disk space is too low: %d Mb left", (totalDiskBytes-usedDiskBytes)/1_024/1_024)
	}

	totalNetworkBandwidth := numbers[5]
	currentNetworkBandwidth := numbers[6]
	networkBandwidthUsage := (totalNetworkBandwidth - currentNetworkBandwidth) / totalNetworkBandwidth * 100
	if networkBandwidthUsage > 90 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available", (totalNetworkBandwidth-currentNetworkBandwidth)/1_024/1_024*8)
	}

	return nil
}
