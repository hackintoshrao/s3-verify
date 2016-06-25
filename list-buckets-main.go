/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"strings"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

// Flags supported by the listbuckets command.
var (
	listBucketsFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Help of list buckets",
		},
	}
)

// listBucketsCmd can be used to run the listbuckets compatibility test.
var listBucketsCmd = cli.Command{
	Name:   "listbuckets",
	Usage:  "Run the list buckets test",
	Action: mainListBuckets,
	Flags:  append(listBucketsFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
	s3verify {{.Name}} - {{.Usage}}

USAGE:
	s3verify {{.Name}} [FLAGS]

FLAGS:
	{{range .Flags}}{{.}}
	{{end}}

EXAMPLES:
	1. Test on the Minio server. Note that play.minio.io is a public test server.
	You are free to use these secret and access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify listbuckets
	2. Test on the Amazon S3 server using flags. Note that passing access and secret keys as flags should be avoided on a multi-user serverfor security reasons.
		$ s3verify listbuckets --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com
	`,
}

var lbMessage = "[1/7] ListBuckets:"

// mainListBuckets - Entry point for the listbuckets command and List Buckets test.
func mainListBuckets(ctx *cli.Context) {
	// TODO: Differentiate different errors: s3verify vs Minio vs test failure.
	// Generate a new config.
	config := newServerConfig(ctx)
	// Spin the scanBar
	scanBar(lbMessage)
	// Create a pseudo body for a http.Response
	expectedBody, err := ListBucketsInit(*config)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(*config, expectedBody); errC != nil {
			console.Fatalln(errC)
		}
		console.Fatalln(err)
	}
	// Spin the scanBar
	scanBar(lbMessage)
	// Generate new List Buckets request.
	req, err := NewListBucketsReq(*config)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(*config, expectedBody); errC != nil {
			console.Fatalln(errC)
		}
		console.Fatalln(err)
	}
	// Spin the scanBar
	scanBar(lbMessage)

	// Generate the server response.
	res, err := ExecRequest(req)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(*config, expectedBody); errC != nil {
			console.Fatalln(errC)
		}
		console.Fatalln(err)
	}
	// Spin the scanBar
	scanBar(lbMessage)

	// Check for S3 Compatibility
	if err := ListBucketsVerify(res, expectedBody); err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(*config, expectedBody); errC != nil {
			console.Fatalln(errC)
		}
		console.Fatalln(err)
	}
	// Spin the scanBar
	scanBar(lbMessage)

	// Delete all Minio created test buckets.
	if err := ListBucketsCleanUp(*config, expectedBody); err != nil {
		console.Fatalln(err)
	}
	// Print final success message.
	console.Eraseline()
	// Pad the message accordingly
	padding := messageWidth - len([]rune(lbMessage))
	console.PrintC(lbMessage + strings.Repeat(" ", padding) + "[OK]\n")

}
