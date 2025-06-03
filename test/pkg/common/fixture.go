// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package common

import "path/filepath"

var (
	basePath                        = filepath.Join("test", "pkg", "common", "fixtures")
	FakeCertFixture                 = filepath.Join(basePath, "fake_cert.yaml")
	LogAnalyticsConfigFixture       = filepath.Join(basePath, "log_analytics_config.yaml")
	LvmStorageClassFixture          = filepath.Join(basePath, "lvm_storageclass.yaml")
	LvmStatefulSetFixture           = filepath.Join(basePath, "lvm_statefulset.yaml")
	LvmAnnotationStatefulSetFixture = filepath.Join(basePath, "lvm_annotation_statefulset.yaml")
)
