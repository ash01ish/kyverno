package autogen

import (
	"testing"

	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ivpol = &policiesv1alpha1.ImageValidatingPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: policiesv1alpha1.ImageValidatingPolicySpec{
			MatchConstraints: &admissionregistrationv1.MatchResources{
				ResourceRules: []admissionregistrationv1.NamedRuleWithOperations{
					{
						RuleWithOperations: admissionregistrationv1.RuleWithOperations{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
								admissionregistrationv1.Update,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{""},
								APIVersions: []string{"v1"},
								Resources:   []string{"pods"},
							},
						},
					},
				},
			},
			ImageRules: []policiesv1alpha1.ImageRule{
				{
					Glob: "ghcr.io/*",
				},
			},
			Images: []policiesv1alpha1.Image{
				{
					Name:       "containers",
					Expression: "object.spec.containers.map(e, e.image)",
				},
			},
			Attestors: []policiesv1alpha1.Attestor{
				{
					Name: "notary",
					Notary: &policiesv1alpha1.Notary{
						Certs: `-----BEGIN CERTIFICATE----------END CERTIFICATE-----`,
					},
				},
			},
			Attestations: []policiesv1alpha1.Attestation{
				{
					Name: "sbom",
					Referrer: &policiesv1alpha1.Referrer{
						Type: "sbom/cyclone-dx",
					},
				},
			},
			Validations: []admissionregistrationv1.Validation{
				{
					Expression: "images.bar.map(image, verifyImageSignatures(image, [attestors.notary])).all(e, e > 0)",
					Message:    "failed to verify image with notary cert",
				},
			},
			AutogenConfiguration: &policiesv1alpha1.ImageValidatingPolicyAutogenConfiguration{
				PodControllers: &policiesv1alpha1.PodControllersGenerationConfiguration{
					Controllers: []string{
						"cronjobs",
					},
				},
			},
		},
	}
)

func Test_AutogenImageVerify(t *testing.T) {
	cronRule := []admissionregistrationv1.NamedRuleWithOperations{
		{
			RuleWithOperations: admissionregistrationv1.RuleWithOperations{
				Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create,
					admissionregistrationv1.Update,
				},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"batch"},
					APIVersions: []string{"v1"},
					Resources:   []string{"cronjobs"},
				},
			},
		},
	}

	podctrl := []admissionregistrationv1.NamedRuleWithOperations{
		{
			RuleWithOperations: admissionregistrationv1.RuleWithOperations{
				Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create,
					admissionregistrationv1.Update,
				},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"apps"},
					APIVersions: []string{"v1"},
					Resources:   []string{"deployments", "statefulsets"},
				},
			},
		},
	}

	cronimg := []policiesv1alpha1.Image{
		{
			Name:       "containers",
			Expression: "object.spec.jobTemplate.spec.template.spec.containers.map(e, e.image)",
		},
	}

	podctrlimg := []policiesv1alpha1.Image{
		{
			Name:       "containers",
			Expression: "object.spec.template.spec.containers.map(e, e.image)",
		},
	}

	autogenerated, err := GetAutogenRulesImageVerify(ivpol)
	assert.NoError(t, err)
	assert.Equal(t, len(autogenerated), 1)
	assert.Equal(t, autogenerated[0].Name, "autogen-cronjobs-test")
	assert.Equal(t, autogenerated[0].Spec.MatchConstraints.ResourceRules, cronRule)
	assert.Equal(t, len(autogenerated[0].Spec.Images), 1)
	assert.Equal(t, autogenerated[0].Spec.Images, cronimg)

	pol := ivpol.DeepCopy()
	pol.Spec.AutogenConfiguration.PodControllers.Controllers = []string{
		"cronjobs",
		"deployments",
		"statefulsets",
	}
	autogenerated, err = GetAutogenRulesImageVerify(pol)
	assert.NoError(t, err)
	assert.Equal(t, len(autogenerated), 2)
	assert.Equal(t, autogenerated[0].Spec.MatchConstraints.ResourceRules, cronRule)
	assert.Equal(t, autogenerated[1].Name, "autogen-test")
	assert.Equal(t, autogenerated[1].Spec.MatchConstraints.ResourceRules, podctrl)
	assert.Equal(t, len(autogenerated[1].Spec.Images), 1)
	assert.Equal(t, autogenerated[1].Spec.Images, podctrlimg)
}
