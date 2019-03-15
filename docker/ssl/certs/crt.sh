# Create ca.
openssl req -newkey rsa:2048 -x509 -nodes -sha512 \
			-keyout ca.key -out ca.crt -subj "/CN=localhost/O=Mainflux/OU=IoT/emailAddress=info@mainflux.com"


# Create mainflux server key and CSR.
openssl genrsa -out mainflux-server.key 4096
openssl req -new -sha256 -key mainflux-server.key -out mainflux-server.csr -subj "/CN=localhost/O=Mainflux/OU=mainflux_server/emailAddress=info@mainflux.com"

# Sign server CSR.
openssl x509 -req -in mainflux-server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out mainflux-server.crt


# Create client key and CSR.
openssl genrsa -out client.key 4096
openssl req -new -sha256 -key client.key -out client.csr -subj "/CN=CLIENT_KEY/O=Mainflux/OU=client/emailAddress=info@mainflux.com"

# Sign client CSR.
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
