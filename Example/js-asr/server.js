var express = require('express');
var app = express();
app.use(express.static(__dirname + '/example'));

var WebSocketServer = require('ws').Server;
var wss = new WebSocketServer({ port: 3002 });

var gWebSocket = null;

wss.on('connection', function connection(ws) {
  gWebSocket = ws;

  var isEOF = function(message) {
    return message[0] == 0x45 && message[1] == 0x4f && message[2] == 0x53;
  }

  var sendMockedResult = function() {
    var mockResult = {
      actors: [
        {
          actor: "{\"gender\": \"male\", \"chinesename\": \"\\u5409\\u59c6\\u00b7\\u5361\\u7279\", \"name\": \"charles carson\", \"realname\": \"jim carter\"}",
          score: "0.5"
        },
        {
          actor: "{\"gender\": \"male\", \"chinesename\": \"\\u5409\\u59c6\\u00b7\\u5361\\u7279\", \"name\": \"charles carson\", \"realname\": \"jim carter\"}",
          score: "0.4"
        }
      ]
    }
    ws.send(JSON.stringify(mockResult));
  }

  ws.on('message', function incoming() {
    console.log(arguments);
    var message = arguments[1]['buffer'];
    var base64Str = message.toString('ascii', 4, message.length);
    var configJson = (new Buffer(base64Str, 'base64')).toString('ascii');
    console.log(configJson);
    console.log(message.length);
    if(isEOF(message)) {
      sendMockedResult();
    }
  });
});

app.get('/ws_send', function(req, res) {
  gWebSocket.send("haha");
  res.json({
    success: true
  });
});

var server = app.listen(8080, function() {
  var host = server.address().address;
  var port = server.address().port;

  console.log("mock server started");
  console.log("Visit: http://localhost:8080/asr.html");
});
