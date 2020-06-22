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
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
)

func NoEmptySlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func RecurringTimer(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func SerializeToString(key interface{}) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func DecodeString(str string) (*bytes.Buffer, error) {
	//var Task *task.Task
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(by), err
}

func DeserializeFromString(str string, t interface{}) error {
	//var Task *task.Task
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	d := gob.NewDecoder(bytes.NewBuffer(by))
	if err := d.Decode(&t); err != nil {
		return err
	}
	return nil
}

func Hostname() string {
	if name, err := os.Hostname(); err == nil {
		n, err := StrictStrip(name)
		if err != nil {
			return "unknown"
		}
		return n
	}
	return "unknown"
}

func GenID() string {

	id := sha256.New()

	net, _ := ghw.Network()

	for _, nic := range net.NICs {
		io.WriteString(id, nic.Name)
		io.WriteString(id, nic.MacAddress)
	}

	// topology, _ := ghw.Topology()
	// for _, node := range topology.Nodes {
	// 	io.WriteString(id, node.String())
	//
	// 	for _, core := range node.Cores {
	// 		io.WriteString(id, core.String())
	// 	}
	// }
	sha := fmt.Sprintf("XX%x", id.Sum(nil))

	return sha
}

func ArrayContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// MD5Bytes encodes string to MD5 bytes.
func MD5Bytes(str string) []byte {
	m := md5.New()
	m.Write([]byte(str))
	return m.Sum(nil)
}

// MD5 encodes string to MD5 hex value.
func MD5(str string) string {
	return hex.EncodeToString(MD5Bytes(str))
}

// SHA1 encodes string to SHA1 hex value.
func SHA1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Cmd(cmdName string, args []string) (string, string, error) {

	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return out.String(), stderr.String(), err
	}
	return out.String(), stderr.String(), nil
}

// isValidUrl tests a string to determine if it is a url or not.
func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

// ShortSHA1 truncates SHA1 string length to at most 10.
func ShortSHA1(sha1 string) string {
	if len(sha1) > 10 {
		return sha1[:10]
	}
	return sha1
}

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomString returns generated random string in given length of characters.
// It also returns possible error during generation.
func RandomString(n int) (string, error) {
	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))

	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer), nil
}

func randomInt(max *big.Int) (int, error) {
	rand, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return int(rand.Int64()), nil
}

func Strip(s string) (string, error) {

	// Make a Regex to say we only want
	reg, err := regexp.Compile("[^a-zA-Z0-9[-][:][.]]+")
	if err != nil {
		return "", err
	}
	processedString := reg.ReplaceAllString(s, "")
	return processedString, nil
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func StrictStrip(s string) (string, error) {
	s = strings.TrimSuffix(strings.ToLower(s), "/")
	// Make a Regex to say we only want
	reg, err := regexp.Compile("[^a-z0-9-/]+")
	if err != nil {
		return "", err
	}
	processedString := reg.ReplaceAllString(s, "")
	return processedString, nil
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
