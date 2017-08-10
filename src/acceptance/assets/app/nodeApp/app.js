var express = require('express');
var app = express();
var http = require('http');

function causeDelay(howMuch) {
    var current = new Date().getTime();
    var till = current + howMuch;
    while (current < till) {
        current = new Date().getTime();
    }
}

app.get('/slow/:time', function (req, res) {
    var delayInMS = parseInt(req.params.time, 10);
    causeDelay(delayInMS);
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
