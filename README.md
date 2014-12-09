WETUBE
======

# Getting Started

1. Clone the repo: `git clone https://github.com/k8si/wetube.git && cd wetube`
2. Set the `WETUBE_ROOT` environment variable: `export WETUBE_ROOT=$PWD`
3. Set your `GOPATH`: `export GOPATH=$PWD/peer`
4. Build all of the things: `./build.sh [your-email] [your-computers-hostname]` (the arguments to `build.sh` are used to generate RSA keys -- the script will ask you for a password, which you should leave blank for security reasons i.e. I am not secure about whether or not anything will work if you provide a password)
5. Start all of the things: `./start.sh [your-public-ip-address]`
6. Open a browser window to `http://localhost:8080` and start tubin'. It might be more fun if you repeat steps 1-5 process on a bunch of internet-connected computers that are on different subnets. See *Kicking the Tires*.

**NOTE** `start.sh` will start the GUI, the http server, and the Wetube client as background processes (it prints the PID for each). If you for some reason should want to stop any of these components, you'll have to `kill [pid]`. Alternatively, you could just run the relevant commands in separate windows (which may be easier anyhow):

			./wetube --ip=[your-public-ip-addr]
			./gui
			DEBUG=app $WETUBE_ROOT/peer/app/bin/www

After you're done with everything, or if you want to start over, you can run `./clean.sh` to clean up all the files generated by `build.sh`.

# System Requirements

* *Nix OS (this code was tested on OSX Yosemite and Ubuntu 14.04.1 LTS)
* node and npm
* go
* openssl

# Kicking the Tires

### Initializing the Network

In order to initialize the wetube network, the user has to explicitly specify that the first node has permission 0, like so:

			./wetube --ip=[your ip address] --permission=0

Then, the user can submit addresses to invite to the network via the browser interface.

### Interactive Mode

If you're running wetube peer on a remote host and for some reason you don't feel like controlling it via the browser, you can start your node in "interactive mode" instead:

			./wetube --ip=[your ip address] --interactive=true

Interactive mode allows you to send messages and list currently connected peers/directors via stdin. The valid commands are:

			msg [message]  #send a message to connected peers e.g. "msg play", "msg pause"
			list  #lists all currently connected peers
			dirs  #lists all currently connected directors

# Probable Problems

* Wetube peers will receive incoming traffic from the outside world on ports 8080 (HTTP) and 3000 (TCP). Ports may have to be forwarded; iptables may have to change.

* If you're running wetube on a remote host e.g. EC2, you can't visit "localhost:8080" and get the web server -- you have to visit "[EC2-IP]:8080". Since the websockets URL is hardcoded into Javascript, this messes up the GUI and causes errors. In order to solve this problem, just change line 19 of `peer/app/public/javascripts/main.js' from:

					var sockurl = 'ws://localhost:4000/jscli';

to

					var sockurl = 'ws://[remote-host-ip-address]:4000/jscli';


* If you're running on EC2, sometimes ExpressJS *mysteriously* doesn't work -- see [this](http://iws.io/hosting-a-nodejs-express-application-on-amazon-web-services-ec2/), you have to do the parts about `nvm` for some reason.
* During `./build.sh`, the edlab machines throw this error:

			go building...
			/nfs/elsrv4/users2/grad/ksilvers/cs630/wetube
			command-line-arguments
			# command-line-arguments
			./director.go:5: import /nfs/elsrv4/users2/grad/ksilvers/cs630/wetube/peer/pkg/linux_386/helper.a: not a package file
			/nfs/elsrv4/users2/grad/ksilvers/cs630/wetube
			command-line-arguments
			# command-line-arguments
			./handlegui.go:5: import /nfs/elsrv4/users2/grad/ksilvers/cs630/wetube/peer/pkg/linux_386/golang.org/x/net/websocket.a: not a package file
			/nfs/elsrv4/users2/grad/ksilvers/cs630/wetube

but not on my local machine or EC2. I'm not sure what the solution is to this yet (TODO).

# References

* https://code.google.com/p/whispering-gophers/source/browse/master/main.go
* https://gist.github.com/spikebike/2232102
