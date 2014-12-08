var ws, status;
var conn;
var host, port, service;
var ipaddr;
var sockport = '8080';

statvals = {"CONN": 2, "ING": 1, "DIS": 0};
datavals = {"SUCC": "1", "FAIL": "0"}

$(function() {
	$('#connect-button').click(function() {
		status = statvals['DIS'];
		ipaddr = $('#ipaddr').val();
		if ((ipaddr == null) || (ipaddr == '')) {
			alert('you need to provide an ip address');
			return;
		}

		service = ipaddr;
		// var sockurl = 'ws://'+ipaddr+':3000/ws';
		var sockurl = 'ws://localhost:4000/jscli';
		// var sockurl = 'ws://'+ipaddr+':3000/jscli';
		console.log('trying to connect to: ' + sockurl);
		ws = new WebSocket(sockurl);
		if (ws == null) {
			console.log('socket creation failed');
			return;
		} else {
			console.log('socket creation successful');
			// "http://www.youtube.com/v/4kQnrKvOTNg&enablejsapi=1&playerapiid=ytplayer",
			var params = { allowScriptAccess: "always" };
			swfobject.embedSWF("http://www.youtube.com/apiplayer?enablejsapi=1&playerapiid=ytplayer", "ytplayer", "425", "365", "8", null, null, params);
		}

		//websocket callbacks
		ws.onopen = function(event) {
			console.log('onopen: trying to connect...');
			setStatus('connecting...');
			status = statvals['ING'];
			ws.send(service);
		}
		ws.onerror = function(event) {
			console.log('onerror');
			status = statvals['DIS'];
			ws.close();
			ws = null;
		}
		ws.onmessage = function(event) {
			if (status == statvals['CONN']) {
				console.log('onmessage: already connected. got:')
				var reader = new FileReader();
				reader.addEventListener("loadend", function() {
					var msg = reader.result;
					console.log(msg)
					if (msg == "play") {
						$('#play-res').click();
					}
					if (msg == "pause") {
						$('#pause-res').click();
					}
					if (msg == "stop") {
						$('#stop-res').click();
					}
					if (msg == "load") {
						$('#load-res').click();
					}
				})
				reader.readAsText(event.data);
			} else {
				if (event.data == datavals['SUCC']) {
					console.log('connected successfully. waiting for new messages...')
					status = statvals['CONN'];
					setStatus('connected.');					
				} else {
					status = statvals['DIS'];
					setStatus('disconnected: ' + event.data);
					ws.close();
					ws = null;
				}
			}
		}
		ws.onclose = function(event) {
			console.log('onclose');
			setStatus('disconnected.');
			status = statvals['DIS'];
			ws.close();
			ws = null;
		}
	});

	function setStatus(str) {
		$('#status').text(str);
	}

	function htmlEncode(str) {
		return $('<div/>').text(value).html();
	}

	function sendMessage(msg) {
		if (ws != null) { //&& (status == statvals['CONN'])){
			ws.send(msg);
			// console.log('status = ' + status + '; sent msg.');
		} else {
			if (ws == null) console.log('sendMessage: no ws');
			if (status != statvals['CONN']) console.log('sendMessage: status != connected');
			alert('there was some error');
		}
	}

	$('#play-button').click(function() {
		sendMessage('play');
	});

	$('#pause-button').click(function() {
		sendMessage('pause');
	});

	$('#stop-button').click(function() {
		sendMessage('stop');
	});

	$('#load-button').click(function() {
		sendMessage('load')
	});


});