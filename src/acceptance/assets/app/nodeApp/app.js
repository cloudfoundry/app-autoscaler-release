const express = require('express')
const app = express()
const fs = require('fs')
const https = require('https')
const axios = require('axios')
const os = require('os')
const cpuCount = os.cpus().length
const { BroadcastChannel, Worker } = require('worker_threads')

let metricsForwarderURL = ''
let mfUsername = ''
let mfPassword = ''
const serviceName = process.env.SERVICE_NAME

function getCredentials () {
  let credentials = {}

  // for service offering
  console.log('Getting credentials...')
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
        const err = 'ERROR: no credentials in bound service env variable'
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
      console.error('ERROR: not all credentials were provided.')
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
  const delayInMS = parseInt(req.params.time, 10)
  await new Promise((resolve, reject) => {
    setTimeout(() => resolve(), delayInMS)
  })
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
          value: parseInt(metricValue),
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

app.get('/cpu/:util/:minute', async function (req, res) {
  let util = parseInt(req.params.util, 10)
  const minute = parseInt(req.params.minute, 10)
  const maxUtil = cpuCount * 100
  util = Math.max(1, util)
  util = Math.min(maxUtil, util)
  const msg =
        'set app cpu utilization to ' + util + '% for ' + minute + ' minutes'
  const maxWorkerUtil = 99
  let remainingUtil = util
  while (remainingUtil > maxWorkerUtil) {
    startWorker(maxWorkerUtil, minute)
    remainingUtil = remainingUtil - maxWorkerUtil
  }
  startWorker(remainingUtil, minute)
  res.status(200).send(msg)
})

function startWorker (util, minute) {
  new Worker('./worker.js', { workerData: { util, minute } }) // eslint-disable-line no-new
}

app.get('/cpu/close', async function (req, res) {
  const bc = new BroadcastChannel('stop_channel')
  bc.postMessage('stop')
  res.status(200).send('close cpu test')
})
