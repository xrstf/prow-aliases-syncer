// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package util

import (
	"fmt"
	"reflect"
	"strings"

	"go.xrstf.de/prow-aliases-syncer/pkg/github"
	"go.xrstf.de/prow-aliases-syncer/pkg/prow"
)

func Equal(oldFileContent string, teams []github.Team, strict bool, keep bool, fileHeader string) (bool, string, error) {
	oldData, err := prow.FromString(oldFileContent)
	if err != nil {
		return false, "", fmt.Errorf("invalid aliases file: %w", err)
	}

	newData := BuildNewOwners(oldData, teams, keep)

	encoded, err := newData.ToYAML(fileHeader)
	if err != nil {
		return false, "", fmt.Errorf("failed to encode YAML: %w", err)
	}

	if strict {
		return strings.TrimSpace(oldFileContent) == strings.TrimSpace(encoded), encoded, nil
	}

	oldData.Sort()
	newData.Sort()

	return reflect.DeepEqual(oldData, newData), encoded, nil
}
