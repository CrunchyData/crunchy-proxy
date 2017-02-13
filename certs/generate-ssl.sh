#!/usr/bin/env bash

SSL_DIR="."

DOMAIN="*.crunchy.lab"

PASSPHRASE="password"

SUBJ="
C=US
ST=VA
L=Arlington
O=Crunchy Data Solutions
CN=$DOMAIN
"

CA_CN=root-ca.crunchy.lab

mkdir -p server client

# ----
# Create self-signed CA
# ----

# Create CA private key.
openssl genrsa -aes256 -out ca.key -passout "pass:$PASSPHRASE" 4096 

# Create self-signed CA certificate.
openssl req -new -x509 -sha256 -days 1825 -key ca.key -out ca.crt \
		-passin "pass:$PASSPHRASE" \
		-subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=root-ca"

# ----
# Create intermediate CAs
# ----

# Create server intermediate private key.
openssl genrsa -aes256 -out server-intermediate.key -passout "pass:$PASSPHRASE" 4096

# Create server intermediate certificate signing request.
openssl req -new -sha256 -days 1825 -key server-intermediate.key \
		-out server-intermediate.csr \
		-subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=server-im-ca" \
		-passin "pass:$PASSPHRASE"

# Create server intermediate certificate by signing with CA certificate.
openssl x509 -extfile /etc/ssl/openssl.cnf -extensions v3_ca -req -days 1825 \
        -CA ca.crt -CAkey ca.key -CAcreateserial \
        -in server-intermediate.csr -out server-intermediate.crt \
		-passin "pass:$PASSPHRASE"

# Create client intermediate private key.
openssl genrsa -aes256 -out client-intermediate.key -passout "pass:$PASSPHRASE" 4096

# Create client intermediate certificate signing request.
openssl req -new -sha256 -days 1825 -key client-intermediate.key \
		-out client-intermediate.csr \
  		-subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=client-im-ca" \
		-passin "pass:$PASSPHRASE"

# Create client intermediate certificate by signing with CA certificate.
openssl x509 -extfile /etc/ssl/openssl.cnf -extensions v3_ca -req -days 1825 \
        -CA ca.crt -CAkey ca.key -CAcreateserial \
        -in client-intermediate.csr -out client-intermediate.crt \
		-passin "pass:$PASSPHRASE"

# ----
# Create server/client certificates
# ----

# Create server certificate signing request.
openssl req -nodes -new -newkey rsa:4096 -sha256 -keyout server.key \
		-out server.csr \
        -subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=$DOMAIN" \
		-passin "pass:$PASSPHRASE"

# Create server certificate by signing with intermediate server CA certificate.
openssl x509 -extfile /etc/ssl/openssl.cnf -extensions usr_cert -req -days 1825 \
        -CA server-intermediate.crt -CAkey server-intermediate.key \
        -CAcreateserial -in server.csr -out server.crt \
		-passin "pass:$PASSPHRASE"

# Create client certificate signing request.
openssl req -nodes -new -newkey rsa:4096 -sha256 -keyout client.key \
		-out client.csr \
        -subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=postgres" \
		-passin "pass:$PASSPHRASE"

# Create client certificate by signing with intermediate client CA certificate.
openssl x509 -extfile /etc/ssl/openssl.cnf -extensions usr_cert -req -days 1825 \
        -CA client-intermediate.crt -CAkey client-intermediate.key \
        -CAcreateserial -in client.csr -out client.crt \
		-passin "pass:$PASSPHRASE"

cp ca.crt server/ca.crt
cp server.key server/server.key
cat server.crt server-intermediate.crt ca.crt > server/server.crt
chmod 600 server/*

cp ca.crt client/ca.crt
cp client.key client/client.key
cat client.crt client-intermediate.crt ca.crt > client/client.crt
chmod 600 client/*

