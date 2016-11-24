/*
	Copyright 2016 Harald Sitter <sitter@kde.org>

	This program is free software; you can redistribute it and/or
	modify it under the terms of the GNU General Public License as
	published by the Free Software Foundation; either version 3 of
	the License or any later version accepted by the membership of
	KDE e.V. (or its successor approved by the membership of KDE
	e.V.), which shall act as a proxy defined in Section 14 of
	version 3 of the license.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"

	"net/http"
	_ "net/http/pprof"
)

func getFile(c *gin.Context, sftp *sftp.Client, path string) {
	fmt.Println("file")
	file, err := sftp.Open(path)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	buffer := bufio.NewReader(file)
	// For unknown reasons reading from sftp files never EOFs, so
	// we need to manually keep track of how much we can and have read and abort
	// once all bytes are read.
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	toRead := stat.Size()
	c.Stream(func(w io.Writer) bool {
		wrote, err := buffer.WriteTo(w)
		toRead -= wrote
		if err != nil || toRead <= 0 {
			return false
		}
		return true
	})
}

func getDir(c *gin.Context, sftp *sftp.Client, path string) {
	fmt.Println("dir")
	fileInfos, err := sftp.ReadDir(path)
	if err != nil {
		panic(err)
	}
	var buffer bytes.Buffer
	buffer.WriteString("<html>")
	for _, info := range fileInfos {
		url := info.Name()
		buffer.WriteString(fmt.Sprintf("<a href='%s'>%s</a><br/>\n", url, url))
	}
	buffer.WriteString("</html>")
	c.Data(http.StatusOK, "text/html", buffer.Bytes())
}

func newSession() (*ssh.Client, *sftp.Client) {
	// key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh/keys/kde.depot-8192"))
	key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh/id_rsa"))
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "ftpneon",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	client, err := ssh.Dial("tcp", "depot.kde.org:22", config)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}

	return client, sftp
}

func allowed(c *gin.Context) bool {
	path := c.Param("path")
	fmt.Println(path)
	return true // the prefix stuff is somewhat broken for reasons
	// return strings.HasPrefix("/stable", path) || strings.HasPrefix("stable", path) ||
	// strings.HasPrefix("/unstable", path) || strings.HasPrefix("unstable", path)
}

func get(c *gin.Context) {
	if !allowed(c) {
		c.String(http.StatusForbidden, "not an allowed path")
		return
	}
	path := "/home/ftpubuntu/" + c.Param("path")

	ssh, sftp := newSession()
	defer ssh.Close()
	defer sftp.Close()

	fileInfo, err := sftp.Stat(path)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	}

	if fileInfo.IsDir() {
		getDir(c, sftp, path)
	} else {
		getFile(c, sftp, path)
	}
}

func main() {
	router := gin.Default()
	router.GET("*path", get)

	port := os.Getenv("PORT")
	if len(port) <= 0 {
		port = "8080"
	}

	// The bridge technically allows inspection of the entire remote user's home
	// this is a security hazard and so this MUST be locked to localhost!
	// 172.17.0.1 is docker container, which is close enough to localhost.
	router.Run("172.17.0.1:" + port)
}
