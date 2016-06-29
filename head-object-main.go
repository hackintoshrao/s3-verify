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

//
var headObjectCmd = cli.Command{
	Name:   "headobject",
	Usage:  "Run the headobject test.",
	Action: mainHeadObject,
	Flags:  append(headObjectFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
	s3verify {{.Name}} - {{.Usage}}

USAGE: 
	s3verify {{.Name}} [COMMAND...] [FLAGS]

FLAGS:
	{{range .Flags}}{{.}}
	{{end}}

EXAMPLES:
	1. Test on the Minio server. Note that play.minio.io is a public test server. You are free to use these secret and access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify headobject
	2. Test on the Amazon S3 server using flags. Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify headobject --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com
	`,
}

// Flags available to the headobject command
var headObjectFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "Help of HEAD object",
	},
}

// Messages to be printed when the HeadObject API is being tested.
// When a new test is added for the HeadObject API make sure to  add its message here.
var (
	headObjectMessages = []string{
		"HeadObject (No Header):",
		//	"HeadObject (Range):",
		//	"HeadObject (If-Match):",
		//	"HeadObject (If-None-Match)",
		//	"HeadObject (If-Modified-Since)",
		//	"HeadObject (If-Unmodified-Since)",
	}
)

// Tests to be run on the HeadObject API.
// When a new test is added for the HeadObject API make sure to add it here.
var (
	headObjectTests = []APItest{
		mainHeadObjectNoHeader,
		// mainHeadObjectRange,
		// mainHeadObjectIfMatch,
		// mainHeadObjectIfNoneMatch,
		// mainHeadObjectIfModifiedSince,
		// mainHeadObjectIfUnModifiedSince,
	}
)

// mainHeadObject - Entry point for the headobject command and test.
func mainHeadObject(ctx *cli.Context) {
	config := newServerConfig(ctx)
	for i, test := range headObjectTests {
		message := fmt.Sprintf("[%d/%d] "+headObjectMessages[i], i+1, len(headObjectMessages))
		if err := test(*config, message); err != nil {
			console.Fatalln(err)
		}
		// Erase the old progress bar.
		console.Eraseline()
		padding := messageWidth - len([]rune(message))
		// Update test as complete.
		console.PrintC(message + strings.Repeat(" ", padding) + "[OK]\n")
	}
}
