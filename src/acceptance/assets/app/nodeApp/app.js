var express = require('express');
var app = express();
var http = require('http');
var sleep = require('sleep');
var request = require('request');
process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

app.get('/slow/:time', function(req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    sleep.msleep(delayInMS);
    res.send('dummy application with slow response');
});

app.get('/fast', function(req, res) {
    res.send('dummy application with fast response');
});

app.get('/', function(req, res) {
    res.send('dummy application root');
});

app.listen(process.env.PORT || 8080, function() {
    console.log('dummy application started');
});

app.get('/custom-metrics/:type/:value', function(req, res) {
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
    request(options, function(err, result, body) {
        if (err) {
            console.log(err);
            res.send(err);
        } else {
            res.send("success");
        }
    });
})