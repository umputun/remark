/*
 * Copyright 2019 Umputun. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package admin

import (
	"encoding/json"

	"github.com/umputun/remark/backend/app/remote"
)

// Remote implements remote engine and delegates all Calls to remote http server
type Remote struct {
	remote.Client
}

// Key returns the key, same for all sites
func (r *Remote) Key() (key string, err error) {
	resp, err := r.Call("admin.key")
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(*resp.Result, &key)
	return key, err
}

// Admins returns list of admin's ids for given site
func (r *Remote) Admins(siteID string) (ids []string, err error) {
	resp, err := r.Call("admin.admins", siteID)
	if err != nil {
		return []string{}, err
	}

	if err = json.Unmarshal(*resp.Result, &ids); err != nil {
		return []string{}, err
	}
	return ids, nil
}

// Email gets email address for given site
func (r *Remote) Email(siteID string) (email string, err error) {
	resp, err := r.Call("admin.email", siteID)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(*resp.Result, &email); err != nil {
		return "", err
	}
	return email, nil
}
