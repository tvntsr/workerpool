package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pool "github.com/tvntsr/workerpool"
)

const (
	pool_size = 512
)

// result of the download:
// - download time
// - page size
func downloading(uri string) (time.Duration, int, error) {
	start := time.Now()
	res, err := http.Get(uri)
	if err != nil {
		return 0, 0, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return 0, 0, fmt.Errorf("Response failed, %d", res.StatusCode)
	}

	return time.Since(start), len(body), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:")
		fmt.Println(os.Args[0], "file_name")
		fmt.Println()
		fmt.Println("File could be retireved by the following commands:")
		fmt.Println("\twget https://s3.amazonaws.com/alexa-static/top-1m.csv.zip")
		fmt.Println("\tunzip top-1m.csv.zip")
		fmt.Println("\tcat top-1m.csv| awk 'BEGIN{ FS=\",\"} {print $2}' > top-1m.list.txt")

		return
	}

	fmt.Println("Parsing", os.Args[1])

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	defer file.Close()

	pool, err := pool.NewUnmanagedPool(pool_size)
	if err != nil {
		fmt.Printf("Error in pool creating %v, terminating...\n", err)
		return
	}

	err = pool.Start()
	if err != nil {
		fmt.Printf("Error in pool start %v, terminating...\n", err)
		return
	}

	var download_time_total time.Duration = 0
	var page_size_total int = 0
	var records int = 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		records++
		rec := scanner.Text()
		_, err = pool.PushTask(func() (interface{}, error) {
			fmt.Println("Checking", rec)
			download_time, page_size, err := downloading("http://" + rec)
			if err != nil {
				fmt.Printf("Page %s failed, error %v\n", rec, err)
				return nil, err
			}
			fmt.Printf("Status %s, size: %d\n", rec, page_size)
			download_time_total += download_time
			page_size_total += page_size
			return nil, nil
		})
		if err != nil {
			fmt.Printf("Cannot add tesk to the pool, %v", err)
		}
	}
	//_ = pool.WaitAllStarted()

	fmt.Println("Scanner done")

	if err := scanner.Err(); err != nil {
		fmt.Printf("Got error %v\n", err)
	}

	_ = pool.Stop()

	fmt.Println("")
	fmt.Println("Total records", records)
	if records > 0 {
		fmt.Printf("Average download time: %v\n", download_time_total/time.Duration(records))
		fmt.Printf("Average page size: %v bytes\n", page_size_total/records)
	}
	fmt.Println("Done")
}
