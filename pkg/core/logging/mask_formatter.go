// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

const (
	maskedResult                 = "********"
	maskJWTAccessTokenPatternStr = "[A-Za-z0-9-_=]+\\.[A-Za-z0-9-_=]+\\.?[A-Za-z0-9-_.+/=]*"
)

var MaskKeyPattern *regexp.Regexp
var MaskJWTAccessTokenPattern *regexp.Regexp

func init() {
	// pattern for JWT access token
	MaskJWTAccessTokenPattern, _ = regexp.Compile(maskJWTAccessTokenPatternStr)
}

func MakeMaskPattern(input string) (*regexp.Regexp, error) {
	p := fmt.Sprintf("\\b%s\\b", input)
	return regexp.Compile(p)
}

type MaskFormatter struct {
	MaskPatterns     []*regexp.Regexp
	DefaultFormatter logrus.Formatter
}

func (f *MaskFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	fields := entry.Data
	for k, v := range entry.Data {
		switch m := v.(type) {
		case string:
			fields[k] = f.mask(m)
		}
	}
	// set masked message back to entry
	entry.Message = f.mask(entry.Message)
	return f.DefaultFormatter.Format(entry)
}

func (f *MaskFormatter) mask(src string) (output string) {
	output = src
	for _, p := range f.MaskPatterns {
		output = p.ReplaceAllString(output, maskedResult)
	}
	return output
}
