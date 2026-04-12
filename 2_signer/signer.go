package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var inputArr map[int]int = make(map[int]int)

type hashPair struct {
	hash      string
	firstIdx  int
	secondIdx int
}

func StartJob(job job, in, out chan any) {
	var wg sync.WaitGroup
	wg.Go(func() {
		job(in, out)
	})
	wg.Wait()
	close(out)
}

func SingleHash(in, out chan any) {
	var wg sync.WaitGroup
	var mainWg sync.WaitGroup
	var inputArrMutex sync.Mutex
	var md5Mutex sync.Mutex
	for range 7 {
		mainWg.Go(func() {
			for num := range in {
				parsedInt, ok := num.(int)
				if !ok {
					continue
				}

				inputArrMutex.Lock()
				inputArr[parsedInt]++
				inputArrMutex.Unlock()

				var tmpCh chan string = make(chan string, 1)
				val := strconv.Itoa(parsedInt)
				wg.Go(func() {
					CalculateCrc32(val, tmpCh)
				})
				md5Mutex.Lock()
				md5 := DataSignerMd5(val)
				md5Mutex.Unlock()
				crc32Md5 := DataSignerCrc32(md5)
				wg.Wait()

				crc32 := <-tmpCh
				pair := hashPair{crc32 + "~" + crc32Md5, parsedInt, 0}
				out <- pair
			}
		})
	}
	mainWg.Wait()
}

func CalculateCrc32(str string, out chan string) {
	res := DataSignerCrc32(str)
	out <- res
	close(out)
}

func MultiHash(in, out chan any) {
	var mainWg sync.WaitGroup
	for range 7 {
		mainWg.Go(func() {
			for val := range in {
				pair, ok := val.(hashPair)
				if !ok {
					continue
				}
				var wg sync.WaitGroup
				for i := range 6 {
					wg.Go(func() {
						tmpCh := make(chan string, 1)
						str := strconv.Itoa(i) + pair.hash
						CalculateCrc32(str, tmpCh)
						res := <-tmpCh
						out <- hashPair{res, pair.firstIdx, i}
					})
				}
				wg.Wait()
			}
		})
	}
	mainWg.Wait()

}

func CombineResults(in, out chan any) {
	var resMutex sync.Mutex
	hashes := []string{}
	result := make(map[int]([]string))
	for val := range in { //Заполняет пришедшие маленькие хеши в map
		pair, ok := val.(hashPair)
		if !ok {
			continue
		}
		resMutex.Lock()
		if _, exists := result[pair.firstIdx]; !exists {
			result[pair.firstIdx] = make([]string, 6)
		}
		result[pair.firstIdx][pair.secondIdx] = pair.hash
		resMutex.Unlock()
	}

	for key, val := range result { // собираем маленькие хеши в большие, смотря также на их кол-во в ответе
		var tmp strings.Builder
		times := inputArr[key]
		for _, hash := range val {
			tmp.WriteString(hash)
		}
		for range times {
			hashes = append(hashes, tmp.String())
		}

	}

	sort.Strings(hashes)
	out <- strings.Join(hashes, "_")
}

func ExecutePipeline(jobs ...job) error {
	var wg sync.WaitGroup
	arr := make([]chan any, len(jobs)+1)
	for i := range arr {
		arr[i] = make(chan any, 1)
	}

	for i := range len(jobs) {
		wg.Go(func() {
			StartJob(jobs[i], arr[i], arr[i+1])
		})

	}
	wg.Wait()
	return nil
}

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	jobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, _ := dataRaw.(string)
			fmt.Println("data from last foo", data)
		}),
	}
	ExecutePipeline(jobs...)
}
