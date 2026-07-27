package doctor

import (
	"context"
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestCheckAPIConnectivity_Reachable(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	result := checkAPIConnectivity(clientset)

	if result.Status != Pass {
		t.Fatalf("Status = %v, want Pass; detail: %s", result.Status, result.Detail)
	}
}

func TestCheckAPIConnectivity_Unreachable(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Fatal("could not assert fake discovery client")
	}
	fakeDiscovery.PrependReactor("get", "version", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection refused")
	})

	result := checkAPIConnectivity(clientset)

	if result.Status != Fail {
		t.Fatalf("Status = %v, want Fail", result.Status)
	}
	if result.Remediation == "" {
		t.Error("expected remediation text on Fail")
	}
}

func TestCheckSchedulableGPU(t *testing.T) {
	tests := []struct {
		name       string
		nodes      []corev1.Node
		wantStatus Status
	}{
		{
			name:       "no nodes",
			nodes:      nil,
			wantStatus: Fail,
		},
		{
			name: "node with no GPU resource",
			nodes: []corev1.Node{
				nodeWithAllocatable("cpu-node", nil),
			},
			wantStatus: Fail,
		},
		{
			name: "node with zero GPUs",
			nodes: []corev1.Node{
				nodeWithAllocatable("gpu-node", map[corev1.ResourceName]resource.Quantity{
					"nvidia.com/gpu": resource.MustParse("0"),
				}),
			},
			wantStatus: Fail,
		},
		{
			name: "node with schedulable GPU",
			nodes: []corev1.Node{
				nodeWithAllocatable("gpu-node", map[corev1.ResourceName]resource.Quantity{
					"nvidia.com/gpu": resource.MustParse("1"),
				}),
			},
			wantStatus: Pass,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			for i := range tt.nodes {
				if _, err := clientset.CoreV1().Nodes().Create(context.Background(), &tt.nodes[i], metav1.CreateOptions{}); err != nil {
					t.Fatalf("seeding node: %v", err)
				}
			}

			result := checkSchedulableGPU(context.Background(), clientset)

			if result.Status != tt.wantStatus {
				t.Fatalf("Status = %v, want %v; detail: %s", result.Status, tt.wantStatus, result.Detail)
			}
		})
	}
}

func TestCheckNamespace(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "miramar-ai-platform"},
		})

		result, missing := checkNamespace(context.Background(), clientset, "miramar-ai-platform")

		if result.Status != Pass {
			t.Fatalf("Status = %v, want Pass", result.Status)
		}
		if missing {
			t.Error("missing = true, want false when the namespace exists")
		}
	})

	t.Run("missing", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()

		result, missing := checkNamespace(context.Background(), clientset, "miramar-ai-platform")

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn", result.Status)
		}
		if !missing {
			t.Error("missing = false, want true when the namespace does not exist")
		}
	})

	t.Run("get error is Fail (deploy-path authorization, not discovery)", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("get", "namespaces", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("forbidden")
		})

		result, missing := checkNamespace(context.Background(), clientset, "miramar-ai-platform")

		if result.Status != Fail {
			t.Fatalf("Status = %v, want Fail — namespace access is a deploy-path prerequisite, unlike storage/ingress discovery", result.Status)
		}
		if missing {
			t.Error("missing = true, want false on an ambiguous Get error")
		}
	})
}

func TestCheckStorageClasses(t *testing.T) {
	t.Run("none found", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()

		result := checkStorageClasses(context.Background(), clientset)

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn", result.Status)
		}
	})

	t.Run("default class found", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "standard",
				Annotations: map[string]string{"storageclass.kubernetes.io/is-default-class": "true"},
			},
		})

		result := checkStorageClasses(context.Background(), clientset)

		if result.Status != Pass {
			t.Fatalf("Status = %v, want Pass; detail: %s", result.Status, result.Detail)
		}
	})

	t.Run("list error is Warn, not Fail", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("list", "storageclasses", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("forbidden")
		})

		result := checkStorageClasses(context.Background(), clientset)

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn on list error (discovery checks never Fail)", result.Status)
		}
		if result.Remediation == "" {
			t.Error("expected remediation text on Warn")
		}
	})
}

func TestCheckIngressController(t *testing.T) {
	t.Run("none found", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()

		result := checkIngressController(context.Background(), clientset)

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn", result.Status)
		}
	})

	t.Run("class found", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
		})

		result := checkIngressController(context.Background(), clientset)

		if result.Status != Pass {
			t.Fatalf("Status = %v, want Pass; detail: %s", result.Status, result.Detail)
		}
	})

	t.Run("list error is Warn, not Fail", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("list", "ingressclasses", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("forbidden")
		})

		result := checkIngressController(context.Background(), clientset)

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn on list error (discovery checks never Fail)", result.Status)
		}
		if result.Remediation == "" {
			t.Error("expected remediation text on Warn")
		}
	})
}

func TestCheckObservabilityStack(t *testing.T) {
	t.Run("nothing detected", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()

		result := checkObservabilityStack(clientset)

		if result.Status != Warn {
			t.Fatalf("Status = %v, want Warn", result.Status)
		}
	})

	t.Run("prometheus operator detected", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		clientset.Resources = []*metav1.APIResourceList{
			{GroupVersion: "monitoring.coreos.com/v1"},
		}

		result := checkObservabilityStack(clientset)

		if result.Status != Pass {
			t.Fatalf("Status = %v, want Pass; detail: %s", result.Status, result.Detail)
		}
		if !strings.Contains(result.Detail, "Prometheus Operator") {
			t.Errorf("Detail = %q, want it to mention Prometheus Operator", result.Detail)
		}
	})
}

func TestCheckHelmReleaseStorage_SkipsWhenNamespaceMissing(t *testing.T) {
	// flags is nil and never dereferenced: the namespaceMissing=true path
	// must return before touching the Helm SDK at all.
	result := checkHelmReleaseStorage(nil, "miramar-ai-platform", true)

	if result.Status != Warn {
		t.Fatalf("Status = %v, want Warn — a missing namespace must not turn into a Helm-storage Fail", result.Status)
	}
	if result.Remediation == "" {
		t.Error("expected remediation text on Warn")
	}
}

// TestFreshCluster_NamespaceAndHelmStorageNeverFail locks in the doctor ->
// deploy handoff the README's golden path depends on: on a fresh cluster,
// where the target namespace doesn't exist yet, neither the namespace check
// nor the Helm release-storage check may FAIL, even though a namespace-scoped
// RBAC setup could make a real Helm Secret-list call against that namespace
// error out before `deploy` ever gets a chance to create it.
func TestFreshCluster_NamespaceAndHelmStorageNeverFail(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	nsResult, missing := checkNamespace(context.Background(), clientset, "miramar-ai-platform")
	if nsResult.Status == Fail {
		t.Fatalf("Target namespace Status = Fail, want Warn on a fresh cluster")
	}
	if !missing {
		t.Fatal("missing = false, want true on a fresh cluster")
	}

	helmResult := checkHelmReleaseStorage(nil, "miramar-ai-platform", missing)
	if helmResult.Status == Fail {
		t.Fatalf("Helm release storage access Status = Fail, want Warn when the namespace doesn't exist yet")
	}
}

func TestReport_HasFailures(t *testing.T) {
	r := &Report{Results: []CheckResult{{Status: Pass}, {Status: Warn}}}
	if r.HasFailures() {
		t.Error("HasFailures() = true, want false for Pass/Warn only")
	}

	r.Results = append(r.Results, CheckResult{Status: Fail})
	if !r.HasFailures() {
		t.Error("HasFailures() = false, want true once a Fail is present")
	}
}

func nodeWithAllocatable(name string, allocatable map[corev1.ResourceName]resource.Quantity) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Allocatable: allocatable,
		},
	}
}
