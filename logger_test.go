package ligno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
	//	"flag"
	"strings"
)

func randString() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//func TestMain(m *testing.M) {
//	flag.Parse()
//	os.Exit(m.Run())
//}

func TestGetLoggerDefault(t *testing.T) {
	emptyStringLogger := GetLogger("")
	if emptyStringLogger.Name() != "" || emptyStringLogger.FullName() != "" {
		t.Fatal("Expected root logger, found: ", emptyStringLogger.FullName())
	}
}

func TestGetLoggerCustom(t *testing.T) {
	for _, loggerName := range []string{
		randString(),
		fmt.Sprintf("%s.%s", randString(), randString()),
		fmt.Sprintf("%s.%s.%s", randString(), randString(), randString()),
	} {
		newLogger := GetLogger(loggerName)
		if newLogger.FullName() != loggerName {
			t.Errorf("Expected logger with full name %s, got %s\n", loggerName, newLogger.FullName())
		}
		parts := strings.Split(loggerName, ".")
		last := parts[len(parts)-1]
		if last != newLogger.Name() {
			t.Errorf("Expected logger with name: %s, got %s\n", last, newLogger.Name)
		}
	}
}

func TestCreate(t *testing.T) {
	params := []struct {
		ctx     Ctx
		handler Handler
		level   Level
	}{
		struct {
			ctx     Ctx
			handler Handler
			level   Level
		}{nil, nil, INFO},
		struct {
			ctx     Ctx
			handler Handler
			level   Level
		}{nil, nil, INFO},
	}

	for _, p := range params {
		l := GetLoggerOptions("test.1", LoggerOptions{
			Context: p.ctx,
			Handler: p.handler,
			Level:   p.level,
		})
		l.Info("AAAA", "foo", "bar", "bla")
		l.Wait()
	}
}

func TestCreateWithName(t *testing.T) {
	l := GetLogger("some_name")
	l.Info("some name message")
	WaitAll()
}

func TestBasicLogging(t *testing.T) {
	log.Println("Info")
	Info("foobar")
	WaitAll()
}

func TestCreate1(t *testing.T) {
	//	ch := make(chan Record, 2048)
	l := GetLoggerOptions("test.1",
		LoggerOptions{
			Context: Ctx{"address": "some address"},
			Handler: FilterLevelHandler(WARNING, StreamHandler(os.Stdout, SimpleFormat())),
			Level:   DEBUG,
		},
	)
	fmt.Println("--->", l.FullName())
	//	l := &Logger{
	//		Context: Record{"address": "some address"},
	//		Handlers: []Handler{
	//			&StdoutHandler{
	////				Formatter: defaultFormatter,
	////				Formatter: &JsonFormatter{},
	//				Filters:   []Filter{FilterLevel(WARNING)},
	//			}},
	//		messages: ch,
	//	}

	//	go l.run()
	//	l.log(Data{EVENT: "some message", "foo": "bar"})
	l.Error("some message that is quite a bit longer then other logged messages", "key", "value")
	l.Critical("Shutting down.")
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		time.Sleep(time.Duration(int(rand.Float32()*100)) * time.Nanosecond)
		l.Log(DEBUG, "g1")
		wg.Done()
		//		l.log(Data{EVENT: "g1", LEVEL: DEBUG})
	}()
	go func() {
		time.Sleep(time.Duration(int(rand.Float32()*100)) * time.Nanosecond)
		l.Log(WARNING, "g1", "a b", "aaa")
		wg.Done()
		//		l.log(Data{EVENT: "g2", "a b": "aaa"})
	}()
	go func() {
		time.Sleep(time.Duration(int(rand.Float32()*100)) * time.Nanosecond)
		l.Log(WARNING, "g3")
		wg.Done()
		//		l.log(Data{"$event": "g3", LEVEL: WARNING})
	}()
	go func() {
		time.Sleep(time.Duration(int(rand.Float32()*100)) * time.Nanosecond)
		l.Log(INFO, "g4")
		wg.Done()
		//		l.log(Data{"$event": "g4"})
	}()
	wg.Wait()
	//	time.Sleep(2 * time.Second)
	runtime.Gosched()
	//	ok := l.WaitTimeout(1 * time.Second)
	//	ok := WaitAllTimeout(1 * time.Second)
	//	WaitAll()
	//	l.Stop()
	//	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	//	fmt.Println("--------------------------------")
	//	pprof.Lookup("heap").WriteTo(os.Stdout, 1)
	l.Wait()
	l.StopAndWait()
	ok := WaitAllTimeout(1 * time.Second)
	if !ok {
		t.Fatal("Not all log messages are processed.")
	}
	debug.FreeOSMemory()
}

func TestJsonMarshalLevel(t *testing.T) {
	m, err := json.Marshal(map[string]Level{"a": ERROR})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(m))
	t.Log(string(m))
	var r map[string]Level
	json.Unmarshal([]byte(m), &r)

	fmt.Printf("--- %#v\n", r)

	m1, _ := json.Marshal(r)
	fmt.Println("-->", string(m1))
}

func BenchmarkPrint(b *testing.B) {
	buff := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(buff, "%s Test", time.Now().UTC())
	}
}

func TestNilHandler(t *testing.T) {
	l := GetLoggerOptions("test.2",
		LoggerOptions{
			Handler: StreamHandler(os.Stderr, TerminalFormat()),
			Level:   INFO,
		})
	l.Info("Some message")
	l.Wait()
}

func TestContext(t *testing.T) {
	l1 := GetLoggerOptions("a", LoggerOptions{
		Context:            Ctx{"a": "a"},
		PreventPropagation: true,
	})
	l2 := l1.SubLoggerOptions("b", LoggerOptions{
		Context:            Ctx{"b": "b"},
		Handler:            StreamHandler(os.Stderr, TerminalFormat()),
		PreventPropagation: true,
	})
	l2.Info("L2 event", "foo", "bar")
	l1.Wait()
}
