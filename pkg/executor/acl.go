/*
Copyright (C) 2020-2022  Daniele Rondina <geaaru@funtoo.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package executor

import (
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd_api "github.com/lxc/lxd/shared/api"
)

func (e *LxdCExecutor) GetAclList() ([]string, error) {
	return e.LxdClient.GetNetworkACLNames()
}

func (e *LxdCExecutor) IsPresentACL(name string) (bool, error) {
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

func (e *LxdCExecutor) CreateACL(acl *specs.LxdCAcl) error {
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
		for idx, _ := range acl.Egress {
			post.NetworkACLPut.Egress = append(
				post.NetworkACLPut.Egress,
				*e.aclRule2Lxd(&acl.Egress[idx]),
			)
		}
	}

	if len(acl.Ingress) > 0 {
		for idx, _ := range acl.Ingress {
			post.NetworkACLPut.Ingress = append(
				post.NetworkACLPut.Ingress,
				*e.aclRule2Lxd(&acl.Ingress[idx]),
			)
		}
	}

	return e.LxdClient.CreateNetworkACL(post)
}

func (e *LxdCExecutor) UpdateACL(acl *specs.LxdCAcl) error {
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
		for idx, _ := range acl.Egress {
			put.Egress = append(
				put.Egress,
				*e.aclRule2Lxd(&acl.Egress[idx]),
			)
		}
	}

	if len(acl.Ingress) > 0 {
		for idx, _ := range acl.Ingress {
			put.Ingress = append(
				put.Ingress,
				*e.aclRule2Lxd(&acl.Ingress[idx]),
			)
		}
	}

	return e.LxdClient.UpdateNetworkACL(acl.Name, put, "")
}

func (e *LxdCExecutor) aclRule2Lxd(rule *specs.LxdCAclRule) *lxd_api.NetworkACLRule {
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
