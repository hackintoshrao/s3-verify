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

var makeBucketCmd = cli.Command{
	Name:   "makebucket",
	Usage:  "Run the makebucket test",
	Action: mainMakeBucket,
	Flags:  append(makeBucketFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
	s3verify {{.Name}} - {{.Usage}}

USAGE: 
	s3verify {{.Name}} [COMMAND...] [FLAGS]

FLAGS:
	{{range .Flags}}{{.}}
	{{end}}

EXAMPLES:
		1. Test on the Minio server. Note that play.minio.io is a public test server. You are free to use these secret and access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify makebucket
	2. Test on the Amazon S3 server using flags. Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify makebucket --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com

	`,
}

var (
	makeBucketFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Help of make bucket",
		},
	}
)

// Messages printed during the running of the MakeBucket tests.
// When a new test for MakeBucket is added make sure to add its message here.
var (
	makeBucketMessages = []string{"MakeBucket (No Header):"}
)

// Declare all tests run for the MakeBucket API.
// When a new test for MakeBucket is added make sure its added here.
var (
	makeBucketTests = []APItest{mainMakeBucketNoHeader}
)

// Entry point for the make bucket test.
func mainMakeBucket(ctx *cli.Context) {
	// TODO: Differentiate different errors: s3verify vs Minio vs test failure.
	// Generate a new config.
	config := newServerConfig(ctx)
	for i, test := range makeBucketTests {
		message := fmt.Sprintf("[%d/%d] "+makeBucketMessages[i], i+1, len(makeBucketTests))
		if err := test(*config, message); err != nil {
			console.Fatalln(err)
		}
		// Print final success message.
		console.Eraseline()
		// Pad accordingly
		padding := messageWidth - len([]rune(message))
		console.PrintC(message + strings.Repeat(" ", padding) + "[OK]\n")
	}
}
