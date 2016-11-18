package appmonitor

import (
	"runtime"
	"testing"
	"time"

	"github.com/wgliang/metrics"
)

func BenchmarkRuntimeMemStats(b *testing.B) {
	r := metrics.NewCollectry()
	collectorRuntimeMemStats(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		captureRuntimeMemStatsOnce(r)
	}
}

func TestRuntimeMemStats(t *testing.T) {
	r := metrics.NewCollectry()
	collectorRuntimeMemStats(r)
	captureRuntimeMemStatsOnce(r)
	zero := runtimeMetrics.MemStats.PauseNs.Count() // Get a "zero" since GC may have run before these tests.
	runtime.GC()
	captureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 1 != count-zero {
		t.Fatal(count - zero)
	}
	runtime.GC()
	runtime.GC()
	captureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 3 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 256; i++ {
		runtime.GC()
	}
	captureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 259 != count-zero {
		t.Fatal(count - zero)
	}
	for i := 0; i < 257; i++ {
		runtime.GC()
	}
	captureRuntimeMemStatsOnce(r)
	if count := runtimeMetrics.MemStats.PauseNs.Count(); 515 != count-zero { // We lost one because there were too many GCs between captures.
		t.Fatal(count - zero)
	}
}

func TestRuntimeMemStatsBlocking(t *testing.T) {
	if g := runtime.GOMAXPROCS(0); g < 2 {
		t.Skipf("skipping TestRuntimeMemStatsBlocking with GOMAXPROCS=%d\n", g)
	}
	ch := make(chan int)
	go testRuntimeMemStatsBlocking(ch)
	var memStats runtime.MemStats
	t0 := time.Now()
	runtime.ReadMemStats(&memStats)
	t1 := time.Now()
	t.Log("i++ during runtime.ReadMemStats:", <-ch)
	go testRuntimeMemStatsBlocking(ch)
	d := t1.Sub(t0)
	t.Log(d)
	time.Sleep(d)
	t.Log("i++ during time.Sleep:", <-ch)
}

func testRuntimeMemStatsBlocking(ch chan int) {
	i := 0
	for {
		select {
		case ch <- i:
			return
		default:
			i++
		}
	}
}
