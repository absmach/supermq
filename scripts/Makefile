CRT_LOCATION = ../docker/ssl/certs
KEY = default
O = Mainflux
OU = mainflux
EA = info@mainflux.com
CN = localhost

all: clean_certs ca server_crt

ca:
	openssl req -newkey rsa:2048 -x509 -nodes -sha512 \
				-keyout $(CRT_LOCATION)/ca.key -out $(CRT_LOCATION)/ca.crt -subj "/CN=localhost/O=Mainflux/OU=IoT/emailAddress=info@mainflux.com"

server_crt:
	# Create mainflux server key and CSR.
	openssl genrsa -out $(CRT_LOCATION)/mainflux-server.key 4096
	openssl req -new -sha256 -key $(CRT_LOCATION)/mainflux-server.key -out $(CRT_LOCATION)/mainflux-server.csr -subj "/CN=$(CN)/O=$(O)/OU=$(OU)/emailAddress=$(EA)"
	# Sign server CSR.
	openssl x509 -req -in $(CRT_LOCATION)/mainflux-server.csr -CA $(CRT_LOCATION)/ca.crt -CAkey $(CRT_LOCATION)/ca.key -CAcreateserial -out $(CRT_LOCATION)/mainflux-server.crt
	# Remove CSR.
	rm $(CRT_LOCATION)/mainflux-server.csr

client_crt:
	# Create mainflux client key and CSR. This requires Thing key to be passed as a KEY variable.
	openssl genrsa -out $(CRT_LOCATION)/client.key 4096
	openssl req -new -sha256 -key $(CRT_LOCATION)/client.key -out $(CRT_LOCATION)/client.csr -subj "/CN=$(KEY)/O=$(O)/OU=$(OU)/emailAddress=$(EA)"
	# Sign client CSR.
	openssl x509 -req -in $(CRT_LOCATION)/client.csr -CA $(CRT_LOCATION)/ca.crt -CAkey $(CRT_LOCATION)/ca.key -CAcreateserial -out $(CRT_LOCATION)/client.crt
	# Remove CSR.
	rm $(CRT_LOCATION)/client.csr

clean_certs:
	rm -r $(CRT_LOCATION)/*.crt
	rm -r $(CRT_LOCATION)/*.key
	rm -r $(CRT_LOCATION)/*.srl
