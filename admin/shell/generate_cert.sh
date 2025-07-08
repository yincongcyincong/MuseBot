#!/bin/bash

set -e

mkdir -p certs
cd certs

echo "🔧 Generating Root CA..."
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt \
  -subj "/C=US/ST=CA/L=SanFrancisco/O=MyOrg/OU=CA/CN=MyRootCA"

# Create OpenSSL config for SAN
cat > server-openssl.cnf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions     = v3_req
prompt = no

[req_distinguished_name]
C  = US
ST = CA
L  = SanFrancisco
O  = MyOrg
OU = Server
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1  = 127.0.0.1
EOF

echo "🔧 Generating Server Certificate with SAN (localhost + 127.0.0.1)..."
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server-openssl.cnf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 -extensions v3_req -extfile server-openssl.cnf

echo "🔧 Generating Client Certificate..."
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/C=US/ST=CA/L=SanFrancisco/O=MyOrg/OU=Client/CN=my-client"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365 -sha256

echo "✅ All certificates and keys have been generated in the certs/ folder:"
ls -l

echo "📄 Files:
- ca.crt       ← Root CA certificate (trusted by both client and server)
- server.crt   ← Server certificate (with SAN for 127.0.0.1 & localhost)
- server.key   ← Server private key
- client.crt   ← Client certificate
- client.key   ← Client private key
"