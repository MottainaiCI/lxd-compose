/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package lxd

import (
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd_api "github.com/canonical/lxd/shared/api"
)

func (e *LxdExecutor) GetAclList() ([]string, error) {
	return e.LxdClient.GetNetworkACLNames()
}

func (e *LxdExecutor) IsPresentACL(name string) (bool, error) {
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

func (e *LxdExecutor) CreateACL(acl *specs.LxdCAcl) error {
	if acl.Name == "" {
		return errors.New("Invalid acl with empty name")
	}

	post := lxd_api.NetworkACLsPost{
		NetworkACLPost: lxd_api.NetworkACLPost{
			Name: acl.Name,
		},
		NetworkACLPut: lxd_api.NetworkACLPut{
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

	return e.LxdClient.CreateNetworkACL(post)
}

func (e *LxdExecutor) UpdateACL(acl *specs.LxdCAcl) error {
	if acl.Name == "" {
		return errors.New("Invalid acl with empty name")
	}

	put := lxd_api.NetworkACLPut{
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

	return e.LxdClient.UpdateNetworkACL(acl.Name, put, "")
}

func (e *LxdExecutor) aclRule2Lxd(rule *specs.LxdCAclRule) *lxd_api.NetworkACLRule {
	return &lxd_api.NetworkACLRule{
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
