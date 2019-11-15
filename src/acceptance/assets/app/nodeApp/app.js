var express = require('express');
var app = express();
var http = require('http');
var request = require('request');
var enableCpuTest = false;

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

app.get('/slow/:time', async function (req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    await new Promise(done => setTimeout(done, delayInMS));
    res.send('dummy application with slow response');
});

app.get('/fast', function (req, res) {
    res.send('dummy application with fast response');
});

app.get('/', function (req, res) {
    res.send('dummy application root');
});

app.listen(process.env.PORT || 8080, function () {
    console.log('dummy application started');
});

app.get('/custom-metrics/:type/:value', function (req, res) {
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
    var credentials = {}
    var metricsForwarderURL = "";
    var mfUsername = "";
    var mfPassword = "";
    // for service offering
    if (process.env.VCAP_SERVICES) {
        var vcapServices = JSON.parse(process.env.VCAP_SERVICES);
        if (vcapServices.autoscaler && vcapServices.autoscaler[0] && vcapServices.autoscaler[0].credentials) {
            credentials = vcapServices.autoscaler[0].credentials;
            metricsForwarderURL = credentials.custom_metrics.url;
            mfUsername = credentials.custom_metrics.username;
            mfPassword = credentials.custom_metrics.password;
        }
    }
    //for build-in offering
    if (metricsForwarderURL === "" || mfUsername === "" || mfPassword === "") {
        credentials = JSON.parse(process.env.AUTO_SCALER_CUSTOM_METRIC_ENV);
        metricsForwarderURL = credentials.url;
        mfUsername = credentials.username;
        mfPassword = credentials.password;
    }

    var options = {
        uri: metricsForwarderURL + '/v1/apps/' + appGuid + '/metrics',
        method: 'POST',
        body: JSON.stringify(postData),
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic ' + Buffer.from(mfUsername + ":" + mfPassword).toString('base64')
        }
    }
    request(options, function (err, result, body) {
        if (err) {
            console.log(err);
            res.send(err);
        } else {
            res.send("success");
        }
    });
})

app.get('/cpu/:util/:minute', async function (req, res) {
    var util = parseInt(req.params.util, 10);
    var minute = parseInt(req.params.minute, 10);
    util = Math.max(1, util);
    util = Math.min(90, util);
    var busyTime = 10;
    var idleTime = busyTime * (100 - util) / util;
    var msg = 'set app cpu utilization to ' + util + '% for ' + minute + ' minutes, busyTime=' + busyTime + ', idleTime=' + idleTime
    console.log(msg);
    res.send(msg);
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
    console.log('finish cpu test');
});

app.get('/cpu/close', async function (req, res) {
    enableCpuTest = false;
    res.send('close cpu test');
});
