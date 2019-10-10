package kustomize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type InstallTest struct {
	expectedCommand string
	installStep     InstallStep
}

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_UnmarshalInstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-robotshop-input.yaml")
	require.NoError(t, err)

	var action InstallAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 2)
	step := action.Steps[0]

	assert.Equal(t, "porter-robotshop-cart", step.Name)
	assert.Equal(t, "Generate the Kubernetes deployment file the Shopping Cart", step.Description)
	assert.Contains(t, step.Kustomization, "kustomize/robotshop/overlays/local/cart")
	assert.Equal(t, "manifests/", step.Manifests)

	assert.Equal(t, map[string]string{"kustomizeBaseGHToken": "{{ bundle.parameters.gh_token }}"}, step.Set)
}

func TestMixin_Install(t *testing.T) {
	microService := "cart"
	name := "porter-robotshop-" + microService
	kustomization := []string{"kustomize/robotshop/overlays/local/"+microService}
	manifests := "manifests"
	reorder := "legacy"
	setArgs := map[string]string{
		"kustomizeBaseGHToken": "{{ bundle.parameters.gh_token }}",
	}

	expectedCmd := fmt.Sprintf("kustomize build %s -o %s/%s", kustomization, manifests, microService)
	expectedGitCmd := fmt.Sprintf("git config --global url.https://{{ bundle.parameters.gh_token }}:@github.com/.insteadOf https://github.com/")
	installTests := []InstallTest{
		{
			expectedCommand: expectedGitCmd,
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:          Step{Description: "Install Robotshop"},
					Name:          name,
					Kustomization: kustomization,
					Manifests:     manifests,
					Set:			setArgs,
					Reorder:       reorder,
				},
			},
		},
		{
			expectedCommand: expectedCmd,
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step:          Step{Description: "Install Robotshop"},
					Name:          name,
					Kustomization: kustomization,
					Manifests:     manifests,
					Set:			setArgs,
					Reorder:       reorder,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, installTest := range installTests {
		t.Run(installTest.expectedCommand, func(t *testing.T) {
			err := os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

			if err != nil {
				os.Exit(-1)
			}
			action := InstallAction{Steps: []InstallStep{installTest.installStep}}
			b, _ := yaml.Marshal(action)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err = h.Install()

			require.NoError(t, err)
		})
	}
}
