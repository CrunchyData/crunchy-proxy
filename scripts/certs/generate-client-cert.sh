USERNAME=$1
PASSPHRASE=$2

# Create client certificate signing request.
openssl req -nodes -new -newkey rsa:4096 -sha256 -keyout "$USERNAME.key" \
		-out "$USERNAME.csr" \
        -subj "/C=US/ST=VA/L=Arlington/O=Crunchy Data Solutions/CN=$USERNAME" \
		-passin "pass:$PASSPHRASE"

# Create client certificate by signing with intermediate client CA certificate.
openssl x509 -extfile /etc/ssl/openssl.cnf -extensions usr_cert -req -days 1825 \
        -CA client-intermediate.crt -CAkey client-intermediate.key \
        -CAcreateserial -in "$USERNAME.csr" -out "$USERNAME.crt" \
		-passin "pass:$PASSPHRASE"

