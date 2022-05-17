var express = require('express');
var app = express();
const fs = require('fs');
const https = require('https');
const axios = require('axios');
var enableCpuTest = false;
const os = require('os');
const cpuCount = os.cpus().length;
const { Worker } = require("worker_threads");

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

app.get('/slow/:time', async function (req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    await new Promise(done => setTimeout(done, delayInMS));
    res.status(200).send('dummy application with slow response');
});

app.get('/fast', function (req, res) {
    res.status(200).send('dummy application with fast response');
});

app.get('/health', function (req, res) {
  res.status(200).json({status: "OK", cpuTestRunning: enableCpuTest});
});

app.get('/', function (req, res) {
    res.status(200).send('dummy application root');
});

app.listen(process.env.PORT || 8080, function () {
    console.log('dummy application started');
});

app.get('/custom-metrics/:type/:value', async function (req, res) {
  try {
    var metricType = req.params.type;
    var metricValue = parseInt(req.params.value, 10);
    var instanceIndex = process.env.CF_INSTANCE_INDEX;
    var appGuid = JSON.parse(process.env.VCAP_APPLICATION).application_id;

    var postData = {
      "instance_index": parseInt(instanceIndex),
      "metrics": [{
        "name": metricType,
        "value": parseInt(metricValue),
        "unit": "test-unit"
      }]
    }
    var credentials = {};
    var metricsForwarderURL = "";
    var mfUsername = "";
    var mfPassword = "";
    // for service offering
    if (process.env.VCAP_SERVICES) {
      var vcapServices = JSON.parse(process.env.VCAP_SERVICES);
      if (vcapServices.autoscaler && vcapServices.autoscaler[0]
          && vcapServices.autoscaler[0].credentials) {
        credentials = vcapServices.autoscaler[0].credentials;
        metricsForwarderURL = credentials.custom_metrics.url;
        mfUsername = credentials.custom_metrics.username;
        mfPassword = credentials.custom_metrics.password;
      }
    }
    //for build-in offering
    if (metricsForwarderURL === "" || mfUsername === "" || mfPassword === "") {
      if (process.env.AUTO_SCALER_CUSTOM_METRIC_ENV) {
        credentials = JSON.parse(process.env.AUTO_SCALER_CUSTOM_METRIC_ENV);
        metricsForwarderURL = credentials.url;
        mfUsername = credentials.username;
        mfPassword = credentials.password;
      } else {
        console.log("Not all credentials!!!!");
        console.log(
            `metricsForwarderURL "${metricsForwarderURL}" || mfUsername === "${mfUsername}" || mfPassword "${mfPassword}`);
        console.log(process.env.VCAP_SERVICES)
        res.status(500).json({error: "No credentials found"})
        return
      }
    }

    var options = {
      url: metricsForwarderURL + '/v1/apps/' + appGuid + '/metrics',
      method: 'POST',
      data: postData,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Basic ' + Buffer.from(
            mfUsername + ":" + mfPassword).toString('base64')
      },
      validateStatus: null
    }
    const result = await axios(options);
    if (result.status !== 200) {
      console.log(`Got non-200 response ${result.status} response ${result.data}`);
      var payload = {
          statusCode: result.status,
          response: result.data,
      }
      res.status(500).json(payload).end();
    } else {
      res.status(200).send("success");
    }
  } catch (err) {
    console.log(err);
    res.status(500).json({exception: JSON.stringify(err)}).end();
  }
})

app.get('/custom-metrics/mtls/:type/:value', async function (req, res) {
    try {
        var metricType = req.params.type;
        var metricValue = parseInt(req.params.value, 10);
        var instanceIndex = process.env.CF_INSTANCE_INDEX;
        var appGuid = JSON.parse(process.env.VCAP_APPLICATION).application_id;

        var postData = {
            "instance_index": parseInt(instanceIndex),
            "metrics": [{
                "name": metricType,
                "value": parseInt(metricValue),
                "unit": "test-unit"
            }]
        }
        var credentials = {};
        var metricsForwarderURL = "";

        // for service offering
        if (process.env.VCAP_SERVICES) {
            var vcapServices = JSON.parse(process.env.VCAP_SERVICES);
            if (vcapServices.autoscaler && vcapServices.autoscaler[0] && vcapServices.autoscaler[0].credentials) {
                credentials = vcapServices.autoscaler[0].credentials;
                metricsForwarderURL = credentials.custom_metrics.mtls_url;
            }
        }

        const httpsAgent = new https.Agent({
            cert: fs.readFileSync(process.env.CF_INSTANCE_CERT),
            key: fs.readFileSync(process.env.CF_INSTANCE_KEY),
        });
        var options = {
            url: metricsForwarderURL + '/v1/apps/' + appGuid + '/metrics',
            method: 'POST',
            data: postData,
            headers: { 'Content-Type': 'application/json' },
            validateStatus: null,
            httpsAgent: httpsAgent
        }
        const result = await axios(options);
        if (result.status !== 200) {
            console.log(`Got non-200 response ${result.status} response ${result.data}`);
            var payload = {
                statusCode: result.status,
                response: result.data,
            }
            res.status(500).json(payload).end();
        } else {
            res.status(200).send("success with mtls").end();
        }

    } catch(err) {
        var payload = { exception: err };
        console.log(payload);
        res.status(500).json(payload);
    }
});

app.get('/cpu/:util/:minute', async function (req, res) {
  var util = parseInt(req.params.util, 10);
  var minute = parseInt(req.params.minute, 10);
  var maxUtil = cpuCount * 100;
  util = Math.max(1, util);
  util = Math.min(maxUtil, util);
  var msg = 'set app cpu utilization to ' + util + '% for ' + minute + ' minutes';
  maxWorkerUtil = 99;
  remainingUtil = util;
  while (remainingUtil > maxWorkerUtil) {
    workerUtil = maxWorkerUtil;
    startWorker(workerUtil, minute);
    remainingUtil = remainingUtil - maxWorkerUtil;
  };
  startWorker(remainingUtil, minute);
  res.status(200).send(msg);
});

function startWorker(util, minute) {
  const worker = new Worker("./worker.js", { workerData: { util: util, minute: minute } });
}

app.get('/cpu/close', async function (req, res) {
    enableCpuTest = false;
    res.status(200).send('close cpu test');
});
