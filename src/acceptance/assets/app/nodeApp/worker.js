const { workerData, BroadcastChannel } = require('worker_threads')

async function loadCPU (util, minute) {
  const bc = new BroadcastChannel('stop_channel')
  const startTime = Date.now()
  const endTime = startTime + minute * 60 * 1000

  const busyMilliSeconds = (util / 100) * 1000
  const idleMilliSeconds = ((100 - util) / 100) * 1000

  console.log(`set app cpu utilization to ${util}% for ${minute} minutes, startTime=${new Date(startTime)}, endTime=${new Date(endTime)} busyTime=${busyMilliSeconds}, idleTime=${idleMilliSeconds}`)

  let enableCpuTest = true
  bc.onmessage = (event) => {
    enableCpuTest = false
    console.log('cpu test worker stopped by request')
    bc.close()
  }

  while (enableCpuTest && Date.now() < endTime) { // eslint-disable-line no-unmodified-loop-condition
    const currentCycleStartTime = Date.now()
    while (enableCpuTest && new Date().getTime() - currentCycleStartTime < busyMilliSeconds) { // eslint-disable-line no-unmodified-loop-condition
      ;
    }
    await new Promise((resolve, reject) => {
      setTimeout(() => resolve(), idleMilliSeconds)
    })
  }
  console.log('cpu test worker finished')
}

loadCPU(workerData.util, workerData.minute)
