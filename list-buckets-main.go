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
	"fmt"
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

// Messages printed during the running of the listBuckets test.
// When a new test for ListBuckets is added make sure its message is added here.
var (
	listBucketsMessages = []string{"ListBuckets (No Params):"}
)

// Declare all tests run for the ListBuckets API.
// When a new test for ListBuckets is added make sure its added here.
var (
	listBucketsTests = []APItest{mainListBucketsExist}
)

// mainListBuckets - Entry point for the listbuckets command and List Buckets test.
func mainListBuckets(ctx *cli.Context) {
	// TODO: Differentiate different errors: s3verify vs Minio vs test failure.
	setGlobalsFromContext(ctx)
	// Generate a new config.
	config := newServerConfig(ctx)
	s3Client, err := NewS3Client(config.Endpoint, config.Access, config.Secret)
	if err != nil {
		console.Fatalln(err)
	}
	for i, test := range listBucketsTests {
		message := fmt.Sprintf("[%d/%d] "+listBucketsMessages[i], i+1, len(listBucketsTests))
		if err := test(*config, *s3Client, message); err != nil {
			console.Fatalln(err)
		}
		// Print final success message.
		console.Eraseline()
		// Pad the message accordingly
		padding := messageWidth - len([]rune(message))
		console.PrintC(message + strings.Repeat(" ", padding) + "[OK]\n")

	}
}
