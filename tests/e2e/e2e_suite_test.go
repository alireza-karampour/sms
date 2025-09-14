package e2e_test

import (
	"testing"

	"github.com/alireza-karampour/sms/tests/helpers"
)

func TestE2E(t *testing.T) {
	helpers.SetupGinkgoSuite(t, "E2E Test Suite")
}