import {parentPort} from 'worker_threads'

let enableCpuTest = false
parentPort.on("message", async (payload, node) => {
    console.log(`cpu: new event ${JSON.stringify(payload)}`)
    if (payload.action && payload.action === "start_load") {
        enableCpuTest = true
        await loadCPU(payload.utilization)
    } else {
        enableCpuTest = false
        console.log(`cpu: got a ${payload.action} ending test`)
        process.exit()
    }
})

async function loadCPU(util) {
    const busyMilliSeconds = Math.min((util / 100) * 1000, 995)
    const idleMilliSeconds = Math.max(((100 - util) / 100) * 1000, 5)

    console.log(`set app cpu utilization to ${util}% busyTime=${busyMilliSeconds}, idleTime=${idleMilliSeconds}`)

    while (enableCpuTest) { // eslint-disable-line no-unmodified-loop-condition
        const currentCycleStartTime = Date.now()
        while (enableCpuTest && new Date().getTime() - currentCycleStartTime < busyMilliSeconds) { // eslint-disable-line no-unmodified-loop-condition
            ;
        }
        await new Promise(r => setTimeout(() => r(), idleMilliSeconds))
    }
    console.log('cpu test worker finished')
}
