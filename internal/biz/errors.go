// Copyright 2023 Nautes Authors
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

package biz

import (
	errors "github.com/go-kratos/kratos/v2/errors"
)

const (
	PROJECT_NOT_FOUND  = "PROJECT_NOT_FOUND"
	GROUP_NOT_FOUND    = "GROUP_NOT_FOUND"
	NODE_NOT_FOUND     = "NODE_NOT_FOUND"
	RESOURCE_NOT_FOUND = "RESOURCE_NOT_FOUND"
	RESOURCE_NOT_MATCH = "RESOURCE_NOT_MATCH"
	NO_AUTHORIZATION   = "NO_AUTHORIZATION"
)

var (
	ErrorProjectNotFound = errors.New(404, PROJECT_NOT_FOUND, "the project path is not found")
	ErrorGroupNotFound   = errors.New(404, GROUP_NOT_FOUND, "the group path is not found")
	ErrorNodetNotFound   = errors.New(404, NODE_NOT_FOUND, "the node is not found")
	ErrorResourceNoFound = errors.New(404, RESOURCE_NOT_FOUND, "the resource is not found")
	ErrorResourceNoMatch = errors.New(500, RESOURCE_NOT_MATCH, "the resource is not match")
	ErrorNoAuth          = errors.New(403, NO_AUTHORIZATION, "no access to the code repository")
)

const _ResourceDoesNotExistOrUnavailable = "During global validation, it was found that %s '%s' does not exist or is unavailable. Please check %s '%s' in directory '%s'."
