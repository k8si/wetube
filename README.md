wetube
======

Current milestone: Basic Go WebSocket infrastructure

# setup
* set WETUBE_ROOT environment variable
		export WETUBE_ROOT=/path/to/wetube
* generate certificates for tls
		chmod +x $WETUBE_ROOT/scripts/makecert.sh
		$WETUBE_ROOT/scripts/makecert.sh [your email address]

# links
* [old/broken tcp chat client&server](http://raycompstuff.blogspot.com/2009/12/simpler-chat-server-and-client-in.html)
* [project iris](https://github.com/project-iris/iris)
* [stackoverflow/p2p connectivity/udp "hole punching"](http://stackoverflow.com/questions/8523330/programming-p2p-application/8524609#8524609)
* [peerjs](http://peerjs.com/)
* [http-server](https://github.com/nodeapps/http-server)
* [websockets](https://developer.mozilla.org/en-US/docs/WebSockets)
* [go websockets api](http://godoc.org/golang.org/x/net/websocket)
* [go websockets tutorial](http://www.ajanicij.info/content/websocket-tutorial-go)