/*
 Copyright 2017 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package connect

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"

	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/crunchydata/crunchy-proxy/util/log"
)

/* SSL constants. */
const (
	SSL_REQUEST_CODE int32 = 80877103

	/* SSL Responses */
	SSL_ALLOWED     byte = 'S'
	SSL_NOT_ALLOWED byte = 'N'

	/* SSL Modes */
	SSL_MODE_REQUIRE     string = "require"
	SSL_MODE_VERIFY_CA   string = "verify-ca"
	SSL_MODE_VERIFY_FULL string = "verify-full"
	SSL_MODE_DISABLE     string = "disable"
)

/*
 *
 */
func UpgradeServerConnection(client net.Conn) net.Conn {
	creds := config.GetCredentials()

	if creds.SSL.Enable {
		tlsConfig := tls.Config{}

		cert, _ := tls.LoadX509KeyPair(
			creds.SSL.SSLServerCert,
			creds.SSL.SSLServerKey)

		tlsConfig.Certificates = []tls.Certificate{cert}

		client = tls.Server(client, &tlsConfig)
	}

	return client
}

/*
 *
 */
func UpgradeClientConnection(hostPort string, connection net.Conn) net.Conn {
	verifyCA := false
	hostname, _, _ := net.SplitHostPort(hostPort)
	tlsConfig := tls.Config{}
	creds := config.GetCredentials()

	/*
	 * Configure the connection based on the mode specificed in the proxy
	 * configuration. Valid mode options are 'require', 'verify-ca',
	 * 'verify-full' and 'disable'. Any other value will result in a fatal
	 * error.
	 */
	switch creds.SSL.SSLMode {

	case SSL_MODE_REQUIRE:
		tlsConfig.InsecureSkipVerify = true

		/*
		 * According to the documentation provided by
		 * https://www.postgresql.org/docs/current/static/libpq-ssl.html, for
		 * backwards compatibility with earlier version of PostgreSQL, if the
		 * root CA file exists, then the behavior of 'sslmode=require' needs to
		 * be the same as 'sslmode=verify-ca'.
		 */
		verifyCA = (creds.SSL.SSLRootCA != "")
	case SSL_MODE_VERIFY_CA:
		tlsConfig.InsecureSkipVerify = true
		verifyCA = true
	case SSL_MODE_VERIFY_FULL:
		tlsConfig.ServerName = hostname
	case SSL_MODE_DISABLE:
		return connection
	default:
		log.Fatalf("Unsupported sslmode %s\n", creds.SSL.SSLMode)
	}

	/* Add client SSL certificate and key. */
	log.Debug("Loading SSL certificate and key")
	cert, _ := tls.LoadX509KeyPair(creds.SSL.SSLCert, creds.SSL.SSLKey)
	tlsConfig.Certificates = []tls.Certificate{cert}

	/* Add root CA certificate. */
	log.Debug("Loading root CA.")
	tlsConfig.RootCAs = x509.NewCertPool()
	rootCA, _ := ioutil.ReadFile(creds.SSL.SSLRootCA)
	tlsConfig.RootCAs.AppendCertsFromPEM(rootCA)

	/* Upgrade the connection. */
	log.Info("Upgrading to SSL connection.")
	client := tls.Client(connection, &tlsConfig)

	if verifyCA {
		log.Debug("Verify CA is enabled")
		err := verifyCertificateAuthority(client, &tlsConfig)
		if err != nil {
			log.Fatalf("Could not verify certificate authority: %s", err.Error())
		} else {
			log.Info("Successfully verified CA")
		}
	}

	return client
}

/*
 * This function will perform a TLS handshake with the server and to verify the
 * certificates against the CA.
 *
 * client - the TLS client connection.
 * tlsConfig - the configuration associated with the connection.
 */
func verifyCertificateAuthority(client *tls.Conn, tlsConf *tls.Config) error {
	err := client.Handshake()

	if err != nil {
		return err
	}

	/* Get the peer certificates. */
	certs := client.ConnectionState().PeerCertificates

	/* Setup the verification options. */
	options := x509.VerifyOptions{
		DNSName:       client.ConnectionState().ServerName,
		Intermediates: x509.NewCertPool(),
		Roots:         tlsConf.RootCAs,
	}

	for i, certificate := range certs {
		/*
		 * The first certificate in the list is client certificate and not an
		 * intermediate certificate. Therefore it should not be added.
		 */
		if i == 0 {
			continue
		}

		options.Intermediates.AddCert(certificate)
	}

	/* Verify the client certificate.
	 *
	 * The first certificate in the certificate to verify.
	 */
	_, err = certs[0].Verify(options)

	return err
}
