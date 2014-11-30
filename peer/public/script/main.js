var ws, status;
var conn;
var host, port, service;
var ipaddr;
var sockport = '8080';


//ip addr: 174.62.219.8

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
		// service = ipaddr + ':8080';
		// service = '174.62.219.8:8000';
		// service = 'localhost:8000';
		if (ws != null) ws.close();
		ws = new WebSocket('ws://localhost:3000/websocket/ws');
		if (ws == null) {
			console.log('socket creation failed');
			return;
		} else {
			console.log('socket creation successful');
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
				console.log('connected. onmessage: ' + event.data);				
			} else {
				console.log('onmessage: event.data = ' + event.data);
				if (event.data == datavals['SUCC']) {
					console.log('connected.')
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
		console.log('trying to send message: ' + msg);
		if ((ws != null)) { //&& (status == statvals['CONN'])){
			ws.send(msg);
			console.log('status = ' + status + '; sent msg anyway.');
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
});