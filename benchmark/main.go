package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/delicb/ligno"

	"flag"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/cihub/seelog"
	"github.com/inconshreveable/log15"
	gommon "github.com/labstack/gommon/log"
	logxi "github.com/mgutz/logxi/v1"
	"resenje.org/logging"
)

// Measurement struct that holds info about single measurement.
type Measurement struct {
	name      string
	average   time.Duration
	total     time.Duration
	logFunc   func()
	afterFunc func()
}

// String is implementation of Stringer interface for Measurement. Returns some basic measurement info.
func (m *Measurement) String() string {
	return fmt.Sprintf("%-15s average time: %10s, total time: %15s", m.name, m.average, m.total)
}

// Run executes measurement and records results.
func (m *Measurement) Run() {
	times := make([]time.Duration, 0, count)
	totalStart := time.Now()
	for i := 0; i < count; i++ {
		start := time.Now()
		m.logFunc()
		times = append(times, time.Now().Sub(start))
	}
	if m.afterFunc != nil {
		m.afterFunc()
	}
	totalEnd := time.Now()
	m.average = time.Duration(int64(avg(times)))
	m.total = totalEnd.Sub(totalStart)
}

// MeasurementList list of measurements. It implements sort.Interface for
// measurement sorting based on different criteria.
type MeasurementList []*Measurement

func (ml MeasurementList) Len() int {
	return len(ml)
}
func (ml MeasurementList) Less(i, j int) bool {
	if total {
		return ml[i].total < ml[j].total
	}
	return ml[i].average < ml[j].average
}
func (ml MeasurementList) Swap(i, j int) {
	ml[i], ml[j] = ml[j], ml[i]
}

var count int
var total bool
var average bool

func main() {
	start := time.Now()
	flag.IntVar(&count, "count", 1024, "Number of messages log log with each logger.")
	flag.BoolVar(&total, "total", false, "Sort by total time logger needs to process messages.")
	flag.BoolVar(&average, "average", true, "Sort by average time logger needs to process message.")
	flag.Parse()

	results := make(MeasurementList, 0)
	results = append(results, &Measurement{
		name:      "Ligno",
		logFunc:   func() { ligno.Info("Ligno message") },
		afterFunc: func() { ligno.WaitAll() },
	})
	results = append(results, &Measurement{
		name:    "Log15",
		logFunc: func() { log15.Info("Log15 message") },
	})
	results = append(results, &Measurement{
		name:      "resenje-logging",
		logFunc:   func() { logging.Info("logging message") },
		afterFunc: func() { logging.WaitForAllUnprocessedRecords() },
	})
	results = append(results, &Measurement{
		name:    "gommon",
		logFunc: func() { gommon.Info("Gommon message") },
	})
	results = append(results, &Measurement{
		name:    "stdlib",
		logFunc: func() { log.Println("Stdlib message") },
	})
	logger := logxi.New("pkg")
	logger.SetLevel(logxi.LevelTrace)
	results = append(results, &Measurement{
		name:    "logxi",
		logFunc: func() { logger.Info("Logxi message") },
	})
	results = append(results, &Measurement{
		name:      "seelog",
		logFunc:   func() { seelog.Info("Seelog message") },
		afterFunc: func() { seelog.Flush() },
	})
	results = append(results, &Measurement{
		name:    "logrus",
		logFunc: func() { logrus.Info("logrus message") },
	})

	var wg sync.WaitGroup
	wg.Add(len(results))
	for _, r := range results {
		//		r.Run()
		go func(m *Measurement) {
			m.Run()
			wg.Done()
		}(r)
	}
	wg.Wait()
	sort.Sort(results)

	fmt.Printf("Logging %d messages.\n", count)
	for _, r := range results {
		fmt.Println(r)
	}
	fmt.Println("Total execution time:", time.Now().Sub(start))
}

func avg(times []time.Duration) float64 {
	var sum int64
	for i := 0; i < len(times); i++ {
		sum += int64(times[i])
	}
	return float64(sum) / float64(len(times))
}
