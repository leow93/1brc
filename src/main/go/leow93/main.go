package main

import (
	"bufio"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

type measurement struct {
	min   float64
	max   float64
	sum   float64
	count int
}

func readFile(filename string) map[string]*measurement {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Open: %v", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Stat: %v", err)
	}

	size := fi.Size()
	if size <= 0 || size != int64(int(size)) {
		log.Fatalf("Invalid file size: %d", size)
	}

	result := make(map[string]*measurement)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		parts := strings.Split(txt, ";")
		if len(parts) != 2 {
			log.Fatalf("expected 2 parts, got %d", len(parts))
		}
		city := parts[0]
		tempStr := parts[1]
		temp, err := strconv.ParseFloat(tempStr, 32)
		if err != nil {
			log.Fatalf("Could not parse float: %s", err.Error())
		}

		m := result[city]
		if m == nil {
			m = &measurement{}
			result[city] = m
		}

		m.count += 1
		m.sum += temp

		if temp > m.max {
			m.max = temp
		}
		if temp < m.min {
			m.min = temp
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return result
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	if len(os.Args) != 2 {
		log.Fatalf("Missing measurements filename")
	}
	measurements := readFile(os.Args[1])

	ids := make([]string, 0, len(measurements))
	for id := range measurements {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Print("{")
	for i, id := range ids {
		v := measurements[id]
		mean := v.sum / float64(v.count)
		if i < len(ids)-1 {
			fmt.Printf("%s=%.1f/%.1f/%.1f, ", id, v.min, mean, v.max)
		} else {
			fmt.Printf("%s=%.1f/%.1f/%.1f", id, v.min, mean, v.max)
		}
	}
	fmt.Print("}")
}
