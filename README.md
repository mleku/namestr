# namestr
simple, zero configuration NIP-05 with LetsEncrypt TLS security

## usage

install go on your VPS, clone this repo:

    git clone https://github.com/mleku/namestr.git

build it:

	cd namestr
    go build

make it so it can listen on ports under 1024:

	sudo setcap 'cap_net_bind_service=+ep' 

make the proper location for the binary to go:

	mkdir -p ~/.local/bin

move the built binary there:

	mv namestr ~/.local/bin/

create a file in the following path

    touch /home/<username>/.local/bin/namestrt
	chmod +x /home/<username>/.local/bin/namestrt

put the details of your desired setup with the parameters you use to launch namestr:

	#!/usr/bin/bash

	/home/<username>/.local/bin/namestr -vc mleku.online npub1mlekuhhxqq6p8w9x5fs469cjk3fu9zw7uptujr55zmpuhhj48u3qnwx3q5 a.nos.lol bevo.nostr1.com bitcoiner.social nos.lol nostr.wine relay.nostr.band Bevo.nostr1.com ae.purplerelay.com nostr.plebchain.org christpill.nostr1.com relay.nostr.me relay.devstr.org relay.primal.net

The `-vc` enables logging, which prints the URL requests that came in. it only sends a proper answer to the NIP-05 request path, all others return the text "gfy". (hah, if you want to change it, just edit the file `cmd/root.go` and look for "gfy" and change it)

after that is the domain name. this is intended for a domain name that is your nostr display name, mleku.online in my case.

after that, you put the npub. this is automatically decoded to hex for the json response.

after that, zero or more of just the domain names of the relays that you usually use to post with. the `wss://` prefix is not needed.

	chmod +x /home/<username>/.local/bin/namestrt

edit the service definition to suit the username you set up on your vps.

mine was 'me' and it appears twice in the service file, one after User and the other in the ExecStart.

edit these to fit the username you created on the VPS.

Then, copy the service using sudo to the correct location:

    sudo cp namestr.service /etc/systemd/system/

and then start 'er up:

	sudo systemctl enable --now namestr

enjoy your shiny new domain NIP-05 nostr/domain name.
