var wss = new Object();

MF.log = function(msg) {
    console.log(msg);
    app.ports.websocketState.send(msg);
}

app.ports.connectWebsocket.subscribe(function(data) {
  var url = 'wss://localhost/ws/channels/' + data.channelid + '/messages?authorization=' + data.thingkey
  if (wss[url]) {
    MF.log("Websocket already open. URL: " + url );
    return;
  }

  var ws = new WebSocket(url);
  
  ws.onopen = function (event) {
    MF.log("Websocket opened. URL: " + url);
    wss[url] = ws;          
  }

  ws.onerror = function (event) {
    console.log(event);
  }
  
  ws.onmessage = function(message) {
    app.ports.websocketIn.send(JSON.stringify({data: message.data, timestamp: message.timeStamp}));
  };
  
  ws.onclose = function () {
    MF.log("Websocket closed. URL: " + url);
    delete wss[ws.url];
  };

});

app.ports.websocketOut.subscribe(function(data) {
  var url = 'wss://localhost/ws/channels/' + data.channelid + '/messages?authorization=' + data.thingkey
  if (wss[url]) {
    wss[url].send(data.message);
  } else {
    MF.log("Websocket is not open. URL: " + url);
  }
});

app.ports.disconnectWebsocket.subscribe(function(data) {
  var url = 'wss://localhost/ws/channels/' + data.channelid + '/messages?authorization=' + data.thingkey
  if (wss[url]) {
    wss[url].close();
  } else {
    MF.log("Websocket is not open. URL: " + url);
  }
})

