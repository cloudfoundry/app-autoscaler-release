package app

//#include <time.h>
import "C"
import "time"

var startTime = time.Now()
var startTicks = C.clock()

// CpuTotalUsageTime https://stackoverflow.com/questions/11356330/how-to-get-cpu-usage/31030753#31030753
func CpuTotalUsageTime() float64 {
	clockSeconds := float64(C.clock()-startTicks) / float64(C.CLOCKS_PER_SEC)
	realSeconds := time.Since(startTime).Seconds()
	return clockSeconds + realSeconds
}
