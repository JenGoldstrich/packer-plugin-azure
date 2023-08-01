// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"testing"
)

func TestConfigRetrieverLeavesTenantIDWhenNotEmpty(t *testing.T) {
	c := Config{CloudEnvironmentName: "AzurePublicCloud"}
	userSpecifiedTid := "not-empty"
	c.TenantID = userSpecifiedTid

	getSubscriptionFromIMDS = func() (string, error) { return "unittest", nil }
	if err := c.FillParameters(); err != nil {
		t.Errorf("Unexpected error when calling c.FillParameters: %v", err)
	}

	if expected := userSpecifiedTid; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}
}

func TestConfigRetrieverReturnsErrorWhenTenantIDEmptyAndRetrievalFails(t *testing.T) {
	c := Config{CloudEnvironmentName: "AzurePublicCloud"}
	if expected := ""; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}

	errorString := "sorry, I failed"
	getSubscriptionFromIMDS = func() (string, error) { return "unittest", nil }
	if err := c.FillParameters(); err != nil && err.Error() != errorString {
		t.Errorf("Unexpected error when calling c.FillParameters: %v", err)
	}
}
