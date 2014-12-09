var ws, status;
var conn;
var host, port, service;
var ipaddr;
var sockport = '8080';
var permission = 0;

statvals = {"CONN": 2, "ING": 1, "DIS": 0};
datavals = {"SUCC": "1", "FAIL": "0"}

$(function() {
	$('#connect-button').click(function() {
		status = statvals['DIS'];
		// ipaddr = $('#ipaddr').val();
		// if ((ipaddr == null) || (ipaddr == '')) {
			// alert('you need to provide an ip address');
			// return;
		// }
		// service = ipaddr;
		var sockurl = 'ws://localhost:4000/jscli';
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
					switch (msg) {
						case "play":
							$('#play-res').click(); break;
						case "pause":
							$('#pause-res').click(); break;
						case "stop":
							$('#stop-res').click(); break;
						case "load":
							$('#load-res').click(); break;
						default:
							var parts = msg.split("&")
							if (parts.length == 2 && parts[0] == "perm") {
								setPermission(parts[1]);
								permission = parseInt(parts[1]);
							}
							break;
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

	function setPermission(str) {
		$('#permission').text(str);
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

	$('#invite-button').click(function() {
		if (permission != undefined && permission == 0) {
			var ipaddr = $('#ipaddr').val();
			if ((ipaddr == null) || (ipaddr == '')) {
				alert('you need to provide an ip address');
				return;
			}
			var perm = $('#perm').val();
			if ((perm == null) || (perm == '')) {
				alert('you need to provide a permission level');
				return;
			}
			var msg = "invite=" + ipaddr + '&perm=' + perm;
			console.log("invite message: " + msg);
			sendMessage(msg);
		} else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});


	$('#play-button').click(function() {
		if (permission != undefined && permission < 2) sendMessage('play');
		else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});

	$('#pause-button').click(function() {
		if (permission && permission < 2) sendMessage('pause');
		else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});

	$('#stop-button').click(function() {
		if (permission && permission < 2) sendMessage('stop');
		else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});

	$('#load-button').click(function() {
		if (permission && permission == 0) sendMessage('load')
		else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});


});