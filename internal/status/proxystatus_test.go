// Copyright Project Contour Authors
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

package status

import (
	"testing"
	"time"

	sesame_api_v1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	"github.com/projectsesame/sesame/internal/k8s"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConditionFor(t *testing.T) {
	simpleValidCondition := sesame_api_v1.DetailedCondition{
		Condition: sesame_api_v1.Condition{
			Type: "Valid",
		},
	}

	pu := ProxyUpdate{
		Fullname: k8s.NamespacedNameFrom("test/test"),
		Conditions: map[ConditionType]*sesame_api_v1.DetailedCondition{
			ValidCondition: &simpleValidCondition,
		},
	}

	got := pu.ConditionFor(ValidCondition)

	assert.Equal(t, simpleValidCondition, *got.DeepCopy())

	emptyProxyUpdate := ProxyUpdate{
		Fullname:   k8s.NamespacedNameFrom("test/test"),
		Conditions: make(map[ConditionType]*sesame_api_v1.DetailedCondition),
	}

	newDc := sesame_api_v1.DetailedCondition{
		Condition: sesame_api_v1.Condition{
			Type:    string(ValidCondition),
			Status:  sesame_api_v1.ConditionTrue,
			Reason:  "Valid",
			Message: "Valid HTTPProxy",
		},
	}
	gotEmpty := emptyProxyUpdate.ConditionFor(ValidCondition)
	assert.Equal(t, newDc, *gotEmpty)

}

func TestStatusMutator(t *testing.T) {
	type testcase struct {
		testProxy         sesame_api_v1.HTTPProxy
		proxyUpdate       ProxyUpdate
		wantConditions    []sesame_api_v1.DetailedCondition
		wantCurrentStatus string
		wantDescription   string
	}

	testTransitionTime := v1.NewTime(time.Now())
	var testGeneration int64 = 7

	run := func(desc string, tc testcase) {
		newProxy := tc.proxyUpdate.Mutate(&tc.testProxy)

		switch o := newProxy.(type) {
		case *sesame_api_v1.HTTPProxy:
			assert.Equal(t, tc.wantConditions, o.Status.Conditions, desc)
			assert.Equal(t, tc.wantCurrentStatus, o.Status.CurrentStatus, desc)
			assert.Equal(t, tc.wantDescription, o.Status.Description, desc)
		default:
			t.Fatal("Got a non-HTTPProxy object.")
		}
	}

	validConditionWarning := testcase{
		testProxy: sesame_api_v1.HTTPProxy{
			ObjectMeta: v1.ObjectMeta{
				Name:       "test",
				Namespace:  "test",
				Generation: testGeneration,
			},
		},
		proxyUpdate: ProxyUpdate{
			Fullname:       k8s.NamespacedNameFrom("test/test"),
			Generation:     testGeneration,
			TransitionTime: testTransitionTime,
			Conditions: map[ConditionType]*sesame_api_v1.DetailedCondition{
				ValidCondition: {
					Condition: sesame_api_v1.Condition{
						Type:    string(ValidCondition),
						Status:  sesame_api_v1.ConditionTrue,
						Reason:  "Valid",
						Message: "Valid HTTPProxy",
					},
					Warnings: []sesame_api_v1.SubCondition{
						{
							Type:    "TLSError",
							Reason:  "TLSConfigError",
							Message: "Syntax Error in TLS Config",
						},
					},
				},
			},
		},
		wantConditions: []sesame_api_v1.DetailedCondition{
			{
				Condition: sesame_api_v1.Condition{
					Type:               string(ValidCondition),
					Status:             sesame_api_v1.ConditionTrue,
					ObservedGeneration: testGeneration,
					LastTransitionTime: testTransitionTime,
					Reason:             "Valid",
					Message:            "Valid HTTPProxy",
				},
				Warnings: []sesame_api_v1.SubCondition{
					{
						Type:    "TLSError",
						Reason:  "TLSConfigError",
						Message: "Syntax Error in TLS Config",
					},
				},
			},
		},
		wantCurrentStatus: string(ProxyStatusValid),
		wantDescription:   "Valid HTTPProxy",
	}
	run("valid with one warning", validConditionWarning)

	inValidConditionError := testcase{
		testProxy: sesame_api_v1.HTTPProxy{
			ObjectMeta: v1.ObjectMeta{
				Name:       "test",
				Namespace:  "test",
				Generation: 6,
			},
		},
		proxyUpdate: ProxyUpdate{
			Fullname:       k8s.NamespacedNameFrom("test/test"),
			Generation:     testGeneration,
			TransitionTime: testTransitionTime,
			Conditions: map[ConditionType]*sesame_api_v1.DetailedCondition{
				ValidCondition: {
					Condition: sesame_api_v1.Condition{
						Type:    string(ValidCondition),
						Status:  sesame_api_v1.ConditionFalse,
						Reason:  "ErrorPresent",
						Message: "At least one error present, see Errors for details",
					},
					Errors: []sesame_api_v1.SubCondition{
						{
							Type:    "TLSError",
							Reason:  "TLSConfigError",
							Message: "Syntax Error in TLS Config",
						},
					},
				},
			},
		},
		wantConditions: []sesame_api_v1.DetailedCondition{
			{
				Condition: sesame_api_v1.Condition{
					Type:               string(ValidCondition),
					Status:             sesame_api_v1.ConditionFalse,
					ObservedGeneration: testGeneration,
					LastTransitionTime: testTransitionTime,
					Reason:             "ErrorPresent",
					Message:            "At least one error present, see Errors for details",
				},
				Errors: []sesame_api_v1.SubCondition{
					{
						Type:    "TLSError",
						Reason:  "TLSConfigError",
						Message: "Syntax Error in TLS Config",
					},
				},
			},
		},
		wantCurrentStatus: string(ProxyStatusInvalid),
		wantDescription:   "At least one error present, see Errors for details",
	}
	run("invalid status, one error", inValidConditionError)

	orphanedCondition := testcase{
		testProxy: sesame_api_v1.HTTPProxy{
			ObjectMeta: v1.ObjectMeta{
				Name:       "test",
				Namespace:  "test",
				Generation: testGeneration,
			},
		},
		proxyUpdate: ProxyUpdate{
			Fullname:       k8s.NamespacedNameFrom("test/test"),
			Generation:     testGeneration,
			TransitionTime: testTransitionTime,
			Conditions: map[ConditionType]*sesame_api_v1.DetailedCondition{
				ValidCondition: {
					Condition: sesame_api_v1.Condition{
						Type:    string(ValidCondition),
						Status:  sesame_api_v1.ConditionFalse,
						Reason:  "Orphaned",
						Message: "this HTTPProxy is not part of a delegation chain from a root HTTPProxy",
					},
					Errors: []sesame_api_v1.SubCondition{
						{
							Type:    "Orphaned",
							Reason:  "Orphaned",
							Message: "this HTTPProxy is not part of a delegation chain from a root HTTPProxy",
						},
					},
				},
			},
		},
		wantConditions: []sesame_api_v1.DetailedCondition{
			{
				Condition: sesame_api_v1.Condition{
					Type:               string(ValidCondition),
					Status:             sesame_api_v1.ConditionFalse,
					ObservedGeneration: testGeneration,
					LastTransitionTime: testTransitionTime,
					Reason:             "Orphaned",
					Message:            "this HTTPProxy is not part of a delegation chain from a root HTTPProxy",
				},
				Errors: []sesame_api_v1.SubCondition{
					{
						Type:    "Orphaned",
						Reason:  "Orphaned",
						Message: "this HTTPProxy is not part of a delegation chain from a root HTTPProxy",
					},
				},
			},
		},
		wantCurrentStatus: string(ProxyStatusOrphaned),
		wantDescription:   "this HTTPProxy is not part of a delegation chain from a root HTTPProxy",
	}

	run("orphaned HTTPProxy", orphanedCondition)

	updateExistingValidCond := testcase{
		testProxy: sesame_api_v1.HTTPProxy{
			ObjectMeta: v1.ObjectMeta{
				Name:       "test",
				Namespace:  "test",
				Generation: testGeneration,
			},
			Status: sesame_api_v1.HTTPProxyStatus{
				Conditions: []sesame_api_v1.DetailedCondition{
					{
						Condition: sesame_api_v1.Condition{
							Type:   string(ValidCondition),
							Status: sesame_api_v1.ConditionTrue,
						},
					},
				},
			},
		},
		proxyUpdate: ProxyUpdate{
			Fullname:       k8s.NamespacedNameFrom("test/test"),
			Generation:     testGeneration,
			TransitionTime: testTransitionTime,
			Conditions: map[ConditionType]*sesame_api_v1.DetailedCondition{
				ValidCondition: {
					Condition: sesame_api_v1.Condition{
						Type:    string(ValidCondition),
						Status:  sesame_api_v1.ConditionTrue,
						Reason:  "Valid",
						Message: "Valid HTTPProxy",
					},
					Warnings: []sesame_api_v1.SubCondition{
						{
							Type:    "TLSError",
							Reason:  "TLSConfigError",
							Message: "Syntax Error in TLS Config",
						},
					},
				},
			},
		},
		wantConditions: []sesame_api_v1.DetailedCondition{
			{
				Condition: sesame_api_v1.Condition{
					Type:               string(ValidCondition),
					Status:             sesame_api_v1.ConditionTrue,
					ObservedGeneration: testGeneration,
					LastTransitionTime: testTransitionTime,
					Reason:             "Valid",
					Message:            "Valid HTTPProxy",
				},
				Warnings: []sesame_api_v1.SubCondition{
					{
						Type:    "TLSError",
						Reason:  "TLSConfigError",
						Message: "Syntax Error in TLS Config",
					},
				},
			},
		},
		wantCurrentStatus: string(ProxyStatusValid),
		wantDescription:   "Valid HTTPProxy",
	}

	run("Test updating existing Valid Condition", updateExistingValidCond)
}
