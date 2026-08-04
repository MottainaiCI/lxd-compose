/*
Copyright © 2020-2025 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	incus_api "github.com/lxc/incus/v7/shared/api"
)

func (e *IncusExecutor) GetCertificates() ([]*specs.LxdCCertificate, error) {
	ans := []*specs.LxdCCertificate{}

	certs, err := e.Client.GetCertificates()
	if err != nil {
		return ans, err
	}

	for idx := range certs {
		ans = append(ans, &specs.LxdCCertificate{
			Name:        certs[idx].Name,
			Type:        certs[idx].Type,
			Restricted:  certs[idx].Restricted,
			Projects:    certs[idx].Projects,
			Certificate: certs[idx].Certificate,
			Fingerprint: certs[idx].Fingerprint,
		})
	}

	return ans, nil
}

func (e *IncusExecutor) DeleteCertificate(fingerprint string) error {
	return e.Client.DeleteCertificate(fingerprint)
}

func (e *IncusExecutor) CreateCertificate(cert *specs.LxdCCertificate) error {

	if cert.Certificate == "" {
		if cert.CertificatePath == "" {
			return fmt.Errorf("Certificate %s without path and inline cert!",
				cert.Name)
		}
		err := cert.ReadCertificate()
		if err != nil {
			return err
		}
	}

	post := incus_api.CertificatesPost{
		CertificatePut: incus_api.CertificatePut{
			Name:        cert.Name,
			Type:        cert.Type,
			Restricted:  cert.Restricted,
			Projects:    cert.Projects,
			Certificate: cert.Certificate,
		},
	}

	return e.Client.CreateCertificate(post)
}

func (e *IncusExecutor) IsPresentCertificate(certName string) (bool, error) {
	ans := false
	list, err := e.GetCertificates()

	if err != nil {
		return false, err
	}

	for _, c := range list {
		if c.Name == certName {
			ans = true
			break
		}
	}

	return ans, nil
}
