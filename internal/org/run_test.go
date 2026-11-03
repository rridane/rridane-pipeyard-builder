package org

import "testing"

// The credentials an image stage pushes with come from the environment, keyed
// by the registry it names.
func TestRegistriesReadTheEnvironment(t *testing.T) {
	t.Setenv(registryHostEnv, "ghcr.io")
	t.Setenv(registryUserEnv, "rridane")
	t.Setenv(registryTokenEnv, "token")

	credential, found := registries()["ghcr.io"]
	if !found || credential.Username != "rridane" || credential.Password != "token" {
		t.Fatalf("got %+v, %v", credential, found)
	}
}

// Without a registry named, the machine's own docker config applies.
func TestRegistriesAreLeftToTheMachineByDefault(t *testing.T) {
	t.Setenv(registryHostEnv, "")

	if got := registries(); got != nil {
		t.Fatalf("got %v", got)
	}
}
