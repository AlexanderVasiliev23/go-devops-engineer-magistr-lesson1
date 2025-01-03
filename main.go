package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	ticker := time.NewTicker(100 * time.Millisecond)

	client := resty.New()

	errsCount := 0

	for range ticker.C {
		select {
		case <-quit:
			return
		default:
		}
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
		Get("http://srv.msk01.gigacorp.local/_stats")

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
	if loadAverage >= 30 {
		fmt.Printf("Load Average is too high: %d\n", loadAverage)
	}

	totalMemBytes := numbers[1]
	usedMemBytes := numbers[2]
	memUsage := float64(usedMemBytes) / float64(totalMemBytes) * 100
	if memUsage >= 80 {
		fmt.Printf("Memory usage too high: %d%%\n", int(memUsage))
	}

	totalDiskBytes := numbers[3]
	usedDiskBytes := numbers[4]
	diskUsage := float64(usedDiskBytes) / float64(totalDiskBytes) * 100
	if diskUsage >= 90 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", (totalDiskBytes-usedDiskBytes)/1_024/1_024)
	}

	totalNetworkBandwidth := numbers[5]
	currentNetworkBandwidth := numbers[6]
	networkBandwidthUsage := float64(currentNetworkBandwidth) / float64(totalNetworkBandwidth) * 100
	if networkBandwidthUsage > 90 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", (totalNetworkBandwidth-currentNetworkBandwidth)/1_000/1_000)
	}

	return nil
}
