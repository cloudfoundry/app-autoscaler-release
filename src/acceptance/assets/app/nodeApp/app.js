var express = require('express');
var app = express();
var http = require('http');
var sleep = require('sleep');

app.get('/slow/:time', function (req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    sleep.msleep(delayInMS);
    res.send('dummy application with slow response');
});

app.get('/fast', function (req, res) {
    res.send('dummy application with fast response');
});

app.get('/', function (req, res) {
    res.send('dummy application root');
});

app.get('/cpuload/:ms', function (req, res) {
  var now = new Date().getTime();
  var result = 0;
  while (new Date().getTime() < now + parseInt(req.params.ms)) {
    result += Math.random() * Math.random();
  }
  res.send('dummy application cpuload ' + req.params.ms + 'ms');
})

app.listen(process.env.PORT || 8080, function () {
  console.log('dummy application started');
});
