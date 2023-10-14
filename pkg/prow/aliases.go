// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package prow

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	OwnersAliasesFilename = "OWNERS_ALIASES"
)

type OwnersAliases struct {
	Aliases map[string][]string
}

func FromString(data string) (*OwnersAliases, error) {
	result := &OwnersAliases{}

	if err := yaml.Unmarshal([]byte(data), result); err != nil {
		return nil, err
	}

	return result, nil
}

func FromFile(filename string) (*OwnersAliases, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromString(string(content))
}

func (oa *OwnersAliases) ToYAML(header string) (string, error) {
	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(oa); err != nil {
		return "", nil
	}

	if header == "" {
		return buf.String(), nil
	}

	return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(header), buf.String()), nil
}

func (os *OwnersAliases) Sort() {
	for team, members := range os.Aliases {
		sort.Strings(members)
		os.Aliases[team] = members
	}
}
