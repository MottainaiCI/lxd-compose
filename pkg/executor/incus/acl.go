/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	incus_api "github.com/lxc/incus/v7/shared/api"
)

func (e *IncusExecutor) GetAclList() ([]string, error) {
	return e.Client.GetNetworkACLNames()
}

func (e *IncusExecutor) IsPresentACL(name string) (bool, error) {
	ans := false
	list, err := e.GetAclList()

	if err != nil {
		return false, err
	}

	for _, n := range list {
		if n == name {
			ans = true
			break
		}
	}

	return ans, nil
}

func (e *IncusExecutor) CreateACL(acl *specs.LxdCAcl) error {
	if acl.Name == "" {
		return errors.New("Invalid acl with empty name")
	}

	post := incus_api.NetworkACLsPost{
		NetworkACLPost: incus_api.NetworkACLPost{
			Name: acl.Name,
		},
		NetworkACLPut: incus_api.NetworkACLPut{
			Description: acl.Description,
			Config:      acl.Config,
		},
	}

	if post.NetworkACLPut.Config == nil {
		post.NetworkACLPut.Config = make(map[string]string, 0)
	}

	if post.NetworkACLPut.Description == "" {
		post.NetworkACLPut.Description = fmt.Sprintf(
			"ACL %s created by lxd-compose", acl.Name,
		)
	}

	if len(acl.Egress) > 0 {
		for idx := range acl.Egress {
			post.NetworkACLPut.Egress = append(
				post.NetworkACLPut.Egress,
				*e.aclRule2Lxd(&acl.Egress[idx]),
			)
		}
	}

	if len(acl.Ingress) > 0 {
		for idx := range acl.Ingress {
			post.NetworkACLPut.Ingress = append(
				post.NetworkACLPut.Ingress,
				*e.aclRule2Lxd(&acl.Ingress[idx]),
			)
		}
	}

	return e.Client.CreateNetworkACL(post)
}

func (e *IncusExecutor) UpdateACL(acl *specs.LxdCAcl) error {
	if acl.Name == "" {
		return errors.New("Invalid acl with empty name")
	}

	put := incus_api.NetworkACLPut{
		Description: acl.Description,
		Config:      acl.Config,
	}

	if put.Config == nil {
		put.Config = make(map[string]string, 0)
	}

	if put.Description == "" {
		put.Description = fmt.Sprintf(
			"ACL %s created by lxd-compose", acl.Name,
		)
	}

	if len(acl.Egress) > 0 {
		for idx := range acl.Egress {
			put.Egress = append(
				put.Egress,
				*e.aclRule2Lxd(&acl.Egress[idx]),
			)
		}
	}

	if len(acl.Ingress) > 0 {
		for idx := range acl.Ingress {
			put.Ingress = append(
				put.Ingress,
				*e.aclRule2Lxd(&acl.Ingress[idx]),
			)
		}
	}

	return e.Client.UpdateNetworkACL(acl.Name, put, "")
}

func (e *IncusExecutor) aclRule2Lxd(rule *specs.LxdCAclRule) *incus_api.NetworkACLRule {
	return &incus_api.NetworkACLRule{
		Action:          rule.Action,
		Source:          rule.Source,
		Destination:     rule.Destination,
		Protocol:        rule.Protocol,
		SourcePort:      rule.SourcePort,
		DestinationPort: rule.DestinationPort,
		ICMPType:        rule.ICMPType,
		ICMPCode:        rule.ICMPCode,
		Description:     rule.Description,
		State:           rule.State,
	}
}
