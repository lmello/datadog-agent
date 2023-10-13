// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package checkconfig

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/snmp/internal/profile"
	"io"
	"os"
	"path/filepath"

	"github.com/DataDog/datadog-agent/pkg/util/log"

	"github.com/DataDog/datadog-agent/pkg/networkdevice/profile/profiledefinition"
)

func loadBundleJSONProfiles() (profile.ProfileConfigMap, error) {
	jsonStr, err := getProfilesBundleJSON()
	if err != nil {
		return nil, err
	}

	profiles, err := unmarshallProfilesBundleJSON(jsonStr)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func unmarshallProfilesBundleJSON(jsonStr []byte) (profile.ProfileConfigMap, error) {
	bundle := profiledefinition.ProfileBundle{}
	err := json.Unmarshal(jsonStr, &bundle)
	if err != nil {
		return nil, err
	}

	profiles := make(profile.ProfileConfigMap)
	for _, p := range bundle.Profiles {
		if p.Profile.Name == "" {
			log.Warnf("Profile with missing name: %s", p.Profile.Name)
			continue
		}

		if _, exist := profiles[p.Profile.Name]; exist {
			log.Warnf("duplicate profile found: %s", p.Profile.Name)
			continue
		}
		// TODO: (separate PR) resolve extends with custom + local default profiles (yaml)
		profiles[p.Profile.Name] = profile.ProfileConfig{
			Definition:    p.Profile,
			IsUserProfile: true,
		}
	}
	return profiles, nil
}

func getProfilesBundleJSON() ([]byte, error) {
	gzipFilePath := getProfileBundleFilePath()
	gzipFile, err := os.Open(gzipFilePath)
	if err != nil {
		return nil, err
	}
	defer gzipFile.Close()
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return nil, err
	}
	jsonStr, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}
	return jsonStr, nil
}

func getProfileBundleFilePath() string {
	return getProfileConfdRoot(filepath.Join(userProfilesFolder, profilesJSONGzipFile))
}

func profileBundleFileExist() bool {
	filePath := getProfileBundleFilePath()
	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}
