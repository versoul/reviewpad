// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_functions

import (
	"strings"

	"github.com/reviewpad/reviewpad/v2/lang/aladino"
	"github.com/reviewpad/reviewpad/v2/utils"
)

func HasFileExtensions() *aladino.BuiltInFunction {
	return &aladino.BuiltInFunction{
		Type: aladino.BuildFunctionType([]aladino.Type{aladino.BuildArrayOfType(aladino.BuildStringType())}, aladino.BuildBoolType()),
		Code: hasFileExtensionsCode,
	}
}

func hasFileExtensionsCode(e aladino.Env, args []aladino.Value) (aladino.Value, error) {
	argExtensions := args[0].(*aladino.ArrayValue)

	extensionSet := make(map[string]bool, len(argExtensions.Vals))
	for _, argExt := range argExtensions.Vals {
		argStringVal := argExt.(*aladino.StringValue)

		normalizedStr := strings.ToLower(argStringVal.Val)
		extensionSet[normalizedStr] = true
	}

	patch := e.GetPatch()
	for fp := range patch {
		fpExt := utils.FileExt(fp)
		normalizedExt := strings.ToLower(fpExt)

		if _, ok := extensionSet[normalizedExt]; !ok {
			return aladino.BuildFalseValue(), nil
		}
	}

	return aladino.BuildTrueValue(), nil
}
