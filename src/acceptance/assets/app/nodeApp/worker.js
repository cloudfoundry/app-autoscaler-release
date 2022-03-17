const { workerData } = require("worker_threads");
var enableCpuTest = false;

async function loadCPU(util, minute) {
    var busyTime = 100;
    var idleTime = busyTime * (100 - util) / util;
    var msg = 'set app cpu utilization to ' + util + '% for ' + minute + ' minutes, busyTime=' + busyTime + ', idleTime=' + idleTime
    console.log(msg);
    var startTime = new Date().getTime();
    var endTime = startTime + minute * 60 * 1000;
    enableCpuTest = true;
    while (enableCpuTest && startTime < endTime) {
        startTime = new Date().getTime();
        while (enableCpuTest && new Date().getTime() - startTime < busyTime) {
            ;
        }
        await new Promise(done => setTimeout(done, idleTime));
    }
}

loadCPU(workerData.util, workerData.minute)
console.log('finish cpu test on worker');