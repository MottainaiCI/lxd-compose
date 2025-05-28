/*
Copyright © 2020-2025 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/base64"

	"github.com/canonical/lxd/shared"
)

func (c *LxdCCertificate) ReadCertificate() error {
	// Add trust relationship.
	x509Cert, err := shared.ReadCert(c.CertificatePath)
	if err != nil {
		return err
	}

	c.Certificate = base64.StdEncoding.EncodeToString(x509Cert.Raw)
	return nil
}
