// Copyright 2020 PingCAP-QE libs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package crawler

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// newGithubClient new clientv4 by github tokens.
func NewGithubClient(token string) *github.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return github.NewClient(httpClient)
}

// FetchLatestArtifact fetch the latest artifact in the repo.
func FetchLatestArtifactUrl(client *github.Client, owner, name string) *url.URL {
	pageIndex := 1
	listOpt := github.ListOptions{
		Page:    pageIndex,
		PerPage: 1,
	}
	artifacts, _, err := client.Actions.ListArtifacts(context.Background(), owner, name, &listOpt)
	if err != nil {
		log.Fatal(err)
	}
	parsedURL, _, err := client.Actions.DownloadArtifact(context.Background(), owner, name, *artifacts.Artifacts[0].ID, false)
	if err != nil {
		log.Fatal(err)
	}
	return parsedURL
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// DownloadAndUnzipArtifact Download And Unzip Artifact by the url from FetchLatestArtifactUrl.
func DownloadAndUnzipArtifact(url url.URL) [][]byte {
	resp, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Fatal(err)
	}

	bytesList := make([][]byte, len(zipReader.File))
	// Read all the files from zip archive
	for i, zipFile := range zipReader.File {
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			log.Println(err)
			continue
		}
		bytesList[i] = unzippedFileBytes
	}

	return bytesList
}
