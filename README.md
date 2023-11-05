# namestr

simple, zero configuration NIP-05 with LetsEncrypt TLS security

well, not very much configuration. the actual thing itself just takes parameters, you have to add the launch parameters on your VPS in the server launch where you pointed your DNS name.

## features

- [x] provide web service that listens at `<domain name>/.well_known/nostr.json` and returns a correctly formed NIP-05 response that enables clients to associate the `npub` with the domain, as the sole identity (without the `name@` prefix)
- [x] provides a simple single redirection for any other request on the domain that forwards to another web address, such as a social media or other hosted account like github.
- [ ] automatic deployment on a VPS with the use of either password authentication or key authentication over SSH.

## usage

### deploying from your computer to a VPS running linux

in releases (soon TM) can be found binaries that contain all of the tooling and code to deploy a namestr server on a linux based VPS built for linux, windows and macos

invocation is the same for all three and will do all the heavy lifting for you in one, and only require the user to understand basic use of the system's shell/command line interface.

    namestr install [<username>@]<domain> <npub> <redirect> [<relays>...]

### running it manually

i am assuming here that you have basic shell and systems administration skills. the following assumes you have a VPS set up, DNS name pointing to the VPS IP, and you've put an installation of Go in the path and pointed to it using `/etc/profile`, including setting up a `GOBIN` where built binaries will be placed after running `go install`.

install go on your VPS, clone this repo:

    git clone https://github.com/mleku/namestr.git

build it:

	cd namestr
    go build

make it so it can listen on ports under 1024:

	sudo setcap 'cap_net_bind_service=+ep' /path/to/namestr

make the proper location for the binary to go:

	mkdir -p ~/.local/bin

move the built binary there:

	mv namestr ~/.local/bin/

create a file in the following path

    touch /home/<username>/.local/bin/namestrt
	chmod +x /home/<username>/.local/bin/namestrt

put the details of your desired setup with the parameters you use to launch namestr:

	#!/usr/bin/bash

	/home/<username>/.local/bin/namestr -vc mleku.online npub1mlekuhhxqq6p8w9x5fs469cjk3fu9zw7uptujr55zmpuhhj48u3qnwx3q5 https://github.com/mleku a.nos.lol bevo.nostr1.com bitcoiner.social nos.lol nostr.wine relay.nostr.band Bevo.nostr1.com ae.purplerelay.com nostr.plebchain.org christpill.nostr1.com relay.nostr.me relay.devstr.org relay.primal.net

the `-vc` enables logging, and color, printing the URL requests that came in, and the IP address they came from. you can see these appear if you run `journalctl -f` as the same user running `namestr` or `root`.

after that is the domain name, `mleku.online` in the example. this is intended for a domain name that is your nostr display name, mleku.online in my case.

then you put the npub. this is automatically decoded to hex for the json response.

third element is a URL that you want to be given as a redirect, such as your github profile or other social network profile, here it is `https://github.com/mleku`, IMPORTANT: it must be the full url ie https://domain.name/path/to/something otherwise it will redirect in a loop, interpreting the path as relative to the domain. this server is designed for running the name only, not for a full website. anyone running a full website can just put the json in a static file or as a request path.

after that, zero or more of just the domain names of the relays that you usually use to post with. the `wss://` prefix is not needed.

next, make the script executable.

	chmod +x /home/<username>/.local/bin/namestrt

having the script separate from the systemd service means you can alter that script, as the user you log in with, and don't have to use `daemon-reload` or disable/enable the service when you change the parameters.

edit the service definition in `namestr.service` to suit the username you set up on your vps.

mine was 'me' and it appears twice in the service file, one after `User` and the other in the `ExecStart`.

edit these to fit the username you created on the VPS.

Then, copy the service using sudo to the correct location:

    sudo cp namestr.service /etc/systemd/system/

and then start 'er up:

	sudo systemctl enable --now namestr

### enjoy your shiny new domain NIP-05 nostr/domain name.
