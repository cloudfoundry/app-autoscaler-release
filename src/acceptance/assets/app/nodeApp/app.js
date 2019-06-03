var express = require('express');
var app = express();
var http = require('http');
var request = require('request');
var sleep = require('sleep');

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

app.get('/slow/:time', function (req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    sleep.msleep(delayInMS);
    res.send('dummy application with slow response');
});

app.get('/fast', function (req, res) {
    res.send('dummy application with fast response');
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

    var credentials = JSON.parse(process.env.VCAP_SERVICES).autoscaler[0].credentials;
    var metricsForwarderURL = credentials.custom_metrics.url;
    var mfUsername = credentials.custom_metrics.username;
    var mfPassword = credentials.custom_metrics.password;

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

app.get('/', function (req, res) {
    res.send('dummy application root');
});

app.listen(process.env.PORT || 8080, function () {
    console.log('dummy application started');
});
