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

import "time"

type GroupOptions struct {
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	Description string `json:"description,omitempty"`
	ParentId    int32  `json:"parent_id,omitempty"`
}

type GitGroupOptions struct {
	Github *GroupOptions
	Gitlab *GroupOptions
}

type GitlabCodeRepoOptions struct {
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	Description string `json:"description,omitempty"`
	NamespaceID int32  `json:"namespace_id,omitempty"`
}

type GitCodeRepoOptions struct {
	Gitlab *GitlabCodeRepoOptions
	Github interface{}
}

type CloneRepositoryParam struct {
	URL   string
	User  string
	Email string
}

type SecretData struct {
	ID   int    `json:"id"`
	Data string `json:"key"`
}

type ProjectDeployKey struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Key       string     `json:"key"`
	CreatedAt *time.Time `json:"created_at"`
	CanPush   bool       `json:"can_push"`
}

type SecretOptions struct {
	SecretPath   string
	SecretEngine string
	SecretKey    string
}

type DeployKeySecretData struct {
	ID          int
	Fingerprint string
}

type ListOptions struct {
	Page    int `url:"page,omitempty" json:"page,omitempty"`
	PerPage int `url:"per_page,omitempty" json:"per_page,omitempty"`
}
