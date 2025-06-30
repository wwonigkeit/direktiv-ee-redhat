package license_test

import (
	"encoding/json"
	"fmt"
	"github.com/direktiv/direktiv/direktiv-ee/pkg/license"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLicenseGenerateAndVerify(t *testing.T) {
	key, pk, err := license.GenerateKeys()
	fmt.Println(string(pk))
	fmt.Println(string(key))
	require.NoError(t, err)

	// Valid license case
	lic := license.License{
		To:   "LocalTester",
		Type: "dev",
		Features: []string{
			"feature1", "feature2",
		},
		ExpiresAt: time.Now().Add(3 * 30 * 24 * time.Hour).Format(time.RFC3339),
	}

	sig, err := license.Sign(lic, key)
	require.NoError(t, err)

	lic.Signature = sig

	v, err := json.Marshal(lic)
	require.NoError(t, err)
	fmt.Printf("%s\n", v)

	err = license.Verify(lic, pk)
	require.NoError(t, err)

	// Tampered license case
	lic.Features[0] = "me"

	err = license.Verify(lic, pk)
	require.Error(t, err)

	// Expired license case
	expiredLic := license.License{
		To: "Expired Pty Ltd",
		Features: []string{
			"feature1", "feature2",
		},
		ExpiresAt: time.Now().Add(-10 * time.Hour).Format(time.RFC3339), // past date
	}

	sig, err = license.Sign(expiredLic, key)
	require.NoError(t, err)

	expiredLic.Signature = sig
	err = license.Verify(expiredLic, pk)
	require.Error(t, err)
	require.Contains(t, err.Error(), "license expired")
}
