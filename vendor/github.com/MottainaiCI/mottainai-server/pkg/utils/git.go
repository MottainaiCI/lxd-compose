/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package utils

import (
	"errors"
	"os"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

//TODO: Git* Can go in a separate object
func GitClone(url, dir string) (*git.Repository, error) {
	//os.RemoveAll(dir)
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		//Progress: os.Stdout,
	})
	if err != nil {
		os.RemoveAll(dir)
		return nil, errors.New("Failed cloning repo: " + url + " " + dir + " " + err.Error())
	}
	return r, nil
}

func GitCheckoutCommit(r *git.Repository, commit string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return err
	}
	return nil
}

func GitCheckoutPullRequest(repo *git.Repository, remote, pullrequest string) error {
	if remote == "" {
		remote = "origin"
	}

	if err := GitFetch(repo, remote, []string{"refs/pull/" + pullrequest + "/head:CI_test"}); err != nil {
		return err
	}
	if err := GitCheckoutCommit(repo, "CI_test"); err != nil {
		return err
	}
	return nil
}

func GitCheckoutMergeRequest(repo *git.Repository, remote, mergeRequest string) error {

	if remote == "" {
		remote = "origin"
	}

	fetchOpts := []string{
		"refs/merge-requests/" + mergeRequest + "/head:CI_test",
	}

	if err := GitFetch(repo, remote, fetchOpts); err != nil {
		return err
	}

	if err := GitCheckoutCommit(repo, "CI_test"); err != nil {
		return err
	}
	return nil
}

func GitFetch(r *git.Repository, remote string, args []string) error {
	var refs []config.RefSpec
	for _, ref := range args {
		refs = append(refs, config.RefSpec(ref))
	}
	err := r.Fetch(&git.FetchOptions{
		RemoteName: remote,
		RefSpecs:   refs,
	})
	if err != nil {
		return err
	}
	return nil
}
