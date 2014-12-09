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


		/* change this line to:
		 ws://[remote-ip-addr]:4000/jscli 
		 if you're trying to access on remote host */
		var sockurl = 'ws://localhost:4000/jscli';


		console.log('trying to connect to: ' + sockurl);
		ws = new WebSocket(sockurl);
		if (ws == null) {
			console.log('socket creation failed');
			return;
		} else {
			console.log('socket creation successful');
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
							if (ytplayer) ytplayer.playVideo();
							break;
						case "pause":
							if(ytplayer) ytplayer.pauseVideo();
							break;
						case "stop":
							if(ytplayer) ytplayer.stopVideo();
							break;
						default:
							var parts = msg.split("&");
							if (parts.length == 2 && parts[0] == "perm") {
								setPermission(parts[1]);
								permission = parseInt(parts[1]);
								break;
							}
							parts = msg.split("=");
							if (parts.length == 2) {
								console.log("got load message")
								if (ytplayer) ytplayer.loadVideoById(parts[1], 0);
								break;
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

	$('#load-button').click(function() {
		if (permission != undefined && permission == 0) {
			var vidid = $('#vidid').val();
			if ((vidid == null) || (vidid =='')) {
				alert('you need to provide a video id to load');
			}
			var msg = "load=" + vidid
			console.log("load message: " + msg);
			sendMessage(msg)
			if (ytplayer) ytplayer.loadVideoById(vidid, 0);
		} else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});


	$('#play-button').click(function() {
		if (permission != undefined && permission < 2) { 
			sendMessage('play');
			if (ytplayer) ytplayer.playVideo();
		} else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});

	$('#pause-button').click(function() {
		if (permission != undefined && permission < 2) {
			sendMessage('pause');
			if(ytplayer) ytplayer.pauseVideo(); 
		} else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});

	$('#stop-button').click(function() {
		if (permission != undefined && permission < 2) {
			sendMessage('stop');
			if(ytplayer) ytplayer.stopVideo();
		} else {
			alert("you dont have permission to do that");
			console.log("permission = " + permission);
		}
	});




});