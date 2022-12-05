const express = require('express')
const app = express()
const fs = require('fs')
const https = require('https')
const axios = require('axios')
const os = require('os')
const cpuCount = os.cpus().length
const { Worker } = require('worker_threads')

let metricsForwarderURL = ''
let mfUsername = ''
let mfPassword = ''
const serviceName = process.env.SERVICE_NAME
let cpuWorkers = []
let memWorker = null
let timer = null

function gc () {
  try {
    if (global.gc) {
      global.gc()
    } else {
      console.log('There is no gc exposed to the worker thread.')
    }
  } catch (e) {
    console.log("Tried to garbage collect Failed please start with the option '--expose-gc'")
  }
}

function getCredentials () {
  // NOTE: the way we check for credentials existence might be further improved.

  let credentials = {}
  console.log('Getting credentials...')

  // for service offering (broker)
  if (process.env.VCAP_SERVICES) {
    console.log(` - found vcap looking for ${serviceName}`)
    const vcapServices = JSON.parse(process.env.VCAP_SERVICES)
    const service = vcapServices[serviceName]
    if (service && service[0]) {
      console.log(` - found service ${serviceName}`)
      if (service[0].credentials) {
        console.log(' - found credentials')
        credentials = service[0].credentials
        metricsForwarderURL = credentials.custom_metrics.url
        mfUsername = credentials.custom_metrics.username
        mfPassword = credentials.custom_metrics.password
      } else {
        const err = 'ERROR: VCAP_SERVICES env variable does not contain expected credentials. '
        console.error(err)
        throw err
      }
    }
  }

  // for build-in offering
  if (metricsForwarderURL === '' || mfUsername === '' || mfPassword === '') {
    console.log(' - looking for creds in env (built in case)')
    if (process.env.AUTO_SCALER_CUSTOM_METRIC_ENV) {
      console.log(' - found credentials in AUTO_SCALER_CUSTOM_METRIC_ENV')
      credentials = JSON.parse(process.env.AUTO_SCALER_CUSTOM_METRIC_ENV)
      metricsForwarderURL = credentials.url
      mfUsername = credentials.username
      mfPassword = credentials.password
    } else {
      console.error('ERROR: AUTO_SCALER_CUSTOM_METRIC_ENV env variable does not contain expected credentials. ')
      console.log(
                `metricsForwarderURL "${metricsForwarderURL}" || mfUsername === "${mfUsername}" || mfPassword (length) "${mfPassword.length}"`
      )
      const err = `ERROR: process.env.AUTO_SCALER_CUSTOM_METRIC_ENV: ${process.env.AUTO_SCALER_CUSTOM_METRIC_ENV}`
      console.error(err)
      throw err
    }
  }
}

async function getMtlsAgent () {
  return new https.Agent({
    cert: await fs.promises.readFile(process.env.CF_INSTANCE_CERT),
    key: await fs.promises.readFile(process.env.CF_INSTANCE_KEY)
  })
}

app.get('/slow/:time', async function (req, res) {
  const delayInMS = Math.min(parseInt(req.params.time, 10), 10000) // Define maximum to avoid attack vector
  await new Promise((resolve) => setTimeout(() => resolve(), delayInMS))
  res.status(200).send('dummy application with slow response')
})

app.get('/fast', function (req, res) {
  res.status(200).send('dummy application with fast response')
})

app.get('/health', function (req, res) {
  res.status(200).json({ status: 'OK' })
})

app.get('/', function (req, res) {
  res.status(200).send('dummy application root')
})

app.listen(process.env.PORT || 8080, function () {
  console.log('dummy application started')
})

app.get('/custom-metrics/:type/:value', async function (req, res) {
  try {
    getCredentials()

    const metricType = req.params.type
    const metricValue = parseInt(req.params.value, 10)
    const instanceIndex = process.env.CF_INSTANCE_INDEX
    const appGuid = JSON.parse(process.env.VCAP_APPLICATION).application_id

    const postData = {
      instance_index: parseInt(instanceIndex),
      metrics: [
        {
          name: metricType,
          value: parseInt(metricValue),
          unit: 'test-unit'
        }
      ]
    }

    const options = {
      url: metricsForwarderURL + '/v1/apps/' + appGuid + '/metrics',
      method: 'POST',
      data: postData,
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Basic ' + Buffer.from(mfUsername + ':' + mfPassword).toString('base64')
      },
      validateStatus: null
    }
    const result = await axios(options)
    if (result.status !== 200) {
      console.log(`Got non-200 response ${result.status} response '${JSON.stringify(result.data)}'`)
      const payload = {
        statusCode: result.status,
        response: result.data
      }
      res.status(500).json(payload).end()
    } else {
      res.status(200).send('success')
    }
  } catch (err) {
    console.log(err)
    res
      .status(500)
      .json({ exception: JSON.stringify(err) })
      .end()
  }
})

app.get('/custom-metrics/mtls/:type/:value', async function (req, res) {
  try {
    getCredentials()

    const metricType = req.params.type
    const metricValue = parseInt(req.params.value, 10)
    const instanceIndex = process.env.CF_INSTANCE_INDEX
    const appGuid = JSON.parse(process.env.VCAP_APPLICATION).application_id

    const postData = {
      instance_index: parseInt(instanceIndex),
      metrics: [
        {
          name: metricType,
          value: metricValue,
          unit: 'test-unit'
        }
      ]
    }
    const options = {
      url: metricsForwarderURL + '/v1/apps/' + appGuid + '/metrics',
      method: 'POST',
      data: postData,
      headers: { 'Content-Type': 'application/json' },
      validateStatus: null,
      httpsAgent: await getMtlsAgent()
    }
    const result = await axios(options)
    if (result.status !== 200) {
      console.log(`Got non-200 response ${result.status} response '${JSON.stringify(result.data)}'`)
      const payload = {
        statusCode: result.status,
        response: result.data
      }
      res.status(500).json(payload).end()
    } else {
      res.status(200).send('success with mtls').end()
    }
  } catch (err) {
    const payload = { exception: err }
    console.log(payload)
    res.status(500).json(payload)
  }
})

app.get('/memory/close', async function (req, res) {
  await stopMemWorker('close')
  res.status(200).json({ status: 'close memory test' })
})

app.get('/memory/stats', async function (req, res) {
  res.status(200).json({ process: process.memoryUsage() })
})

function startMemWorker (src) {
  stopMemWorker(src)
  memWorker = new Worker('./mem_load.mjs')
}
app.get('/memory/:megabytes/:minute', async function (req, res) {
  startMemWorker('/memory/:megabytes/:minute')
  let megabytes = parseInt(req.params.megabytes, 10)
  const minute = parseInt(req.params.minute, 10)
  const memoryMax = parseInt(process.env.MEMORY_MAX)
  const ramUsableInPercent = 0.85 // If exceeded, OS may terminate the process.
  const totalMemMb = memoryMax * ramUsableInPercent
  if (megabytes > totalMemMb) {
    res.status(400).json({
      result: 'asked for more heap than is available',
      memory_usage: process.memoryUsage(),
      requested_heap_usage: megabytes / 1000,
      application_limit: process.env.MEMORY_MAX,
      max_memory_allowed: totalMemMb
    })
    return
  }
  megabytes = Math.max(1, megabytes)
  memWorker.postMessage({ action: 'chew', totalMemoryUsage: megabytes, source: '/memory/:megabytes/:minute' })
  timer = setTimeout(async () => stopMemWorker('timer'), minute * 60 * 1000)
  res.status(200).json({ result: 'success', msg: `using worker to allocate ${megabytes}MB of heap for ${minute} minutes` })
})

function stopMemWorker (src) {
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
  if (memWorker) {
    memWorker.postMessage({ action: 'stop', source: src })
  }
  memWorker = null
  gc()
}

function stopCpuWorkers (src) {
  cpuWorkers.forEach(worker => worker.postMessage({ action: 'stop', source: src }))
  cpuWorkers = []
}

function startCpuWorker (utilization) {
  const worker = new Worker('./cpu_load.mjs')
  worker.postMessage({
    action: 'start_load',
    utilization,
    source: `/cpu/:util/:minute worker[${cpuWorkers.length}]`
  })
  cpuWorkers.push(worker)
}

app.get('/cpu/:util/:minute', async function (req, res) {
  stopCpuWorkers('start')
  let util = parseInt(req.params.util, 10)
  const minute = parseInt(req.params.minute, 10)
  const maxUtil = cpuCount * 100
  util = Math.max(1, util)
  if (util > maxUtil) {
    res.status(400).json({
      result: 'asked for more cpu util than is available',
      cpus: cpuCount,
      max_util: maxUtil
    })
    return
  }
  const msg = 'set app cpu utilization to ' + util + '% for ' + minute + ' minutes'
  const maxWorkerUtil = 100
  let remainingUtil = util
  while (remainingUtil > maxWorkerUtil) {
    startCpuWorker(maxWorkerUtil)
    remainingUtil = remainingUtil - maxWorkerUtil
  }
  startCpuWorker(remainingUtil)
  setTimeout(() => stopCpuWorkers('timer'), minute * 60 * 1000)
  res.status(200).send(msg)
})

app.get('/cpu/close', async function (req, res) {
  stopCpuWorkers()
  res.status(200).json({ status: 'close cpu test' })
})

gc()
