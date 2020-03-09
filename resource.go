package cke

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Annotations for CKE-managed resources.
const (
	AnnotationResourceImage    = "cke.cybozu.com/image"
	AnnotationResourceRevision = "cke.cybozu.com/revision"
)

// Kind represents Kubernetes resource kind
type Kind string

// Supported resource kinds
const (
	KindNamespace           = "Namespace"
	KindServiceAccount      = "ServiceAccount"
	KindPodSecurityPolicy   = "PodSecurityPolicy"
	KindNetworkPolicy       = "NetworkPolicy"
	KindClusterRole         = "ClusterRole"
	KindRole                = "Role"
	KindClusterRoleBinding  = "ClusterRoleBinding"
	KindRoleBinding         = "RoleBinding"
	KindConfigMap           = "ConfigMap"
	KindDeployment          = "Deployment"
	KindDaemonSet           = "DaemonSet"
	KindCronJob             = "CronJob"
	KindService             = "Service"
	KindPodDisruptionBudget = "PodDisruptionBudget"
)

// IsSupported returns true if k is supported by CKE.
func (k Kind) IsSupported() bool {
	switch k {
	case KindNamespace, KindServiceAccount,
		KindPodSecurityPolicy, KindNetworkPolicy,
		KindClusterRole, KindRole, KindClusterRoleBinding, KindRoleBinding,
		KindConfigMap, KindDeployment, KindDaemonSet, KindCronJob, KindService, KindPodDisruptionBudget:
		return true
	}
	return false
}

// Order returns the precedence of resource creation order as an integer.
func (k Kind) Order() int {
	switch k {
	case KindNamespace:
		return 1
	case KindServiceAccount:
		return 2
	case KindPodSecurityPolicy:
		return 3
	case KindNetworkPolicy:
		return 4
	case KindClusterRole:
		return 5
	case KindRole:
		return 6
	case KindClusterRoleBinding:
		return 7
	case KindRoleBinding:
		return 8
	case KindConfigMap:
		return 9
	case KindDeployment:
		return 10
	case KindDaemonSet:
		return 11
	case KindCronJob:
		return 12
	case KindService:
		return 13
	case KindPodDisruptionBudget:
		return 14
	}
	panic("unknown kind: " + string(k))
}

var resourceDecoder runtime.Decoder
var resourceEncoder runtime.Encoder

func init() {
	gvs := schema.GroupVersions{
		schema.GroupVersion{Group: corev1.SchemeGroupVersion.Group, Version: corev1.SchemeGroupVersion.Version},
		schema.GroupVersion{Group: policyv1beta1.SchemeGroupVersion.Group, Version: policyv1beta1.SchemeGroupVersion.Version},
		schema.GroupVersion{Group: networkingv1.SchemeGroupVersion.Group, Version: networkingv1.SchemeGroupVersion.Version},
		schema.GroupVersion{Group: rbacv1.SchemeGroupVersion.Group, Version: rbacv1.SchemeGroupVersion.Version},
		schema.GroupVersion{Group: appsv1.SchemeGroupVersion.Group, Version: appsv1.SchemeGroupVersion.Version},
		schema.GroupVersion{Group: batchv1beta1.SchemeGroupVersion.Group, Version: batchv1beta1.SchemeGroupVersion.Version},
	}
	resourceDecoder = scheme.Codecs.DecoderToVersion(scheme.Codecs.UniversalDeserializer(), gvs)
	resourceEncoder = json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, json.SerializerOptions{})
}

func encodeToJSON(obj runtime.Object) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := resourceEncoder.Encode(obj, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ApplyResource creates or patches Kubernetes object.
func ApplyResource(clientset *kubernetes.Clientset, data []byte, rev int64, forceConflicts bool) error {
	obj, gvk, err := resourceDecoder.Decode(data, nil, nil)
	if err != nil {
		return err
	}

	switch o := obj.(type) {
	case *corev1.Namespace:
		c := clientset.CoreV1().RESTClient()
		return applyNamespace(o, rev, c, applyParams{isNamespaced: false, forceConflicts: forceConflicts})
	case *corev1.ServiceAccount:
		c := clientset.CoreV1().RESTClient()
		return applyServiceAccount(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *corev1.ConfigMap:
		c := clientset.CoreV1().RESTClient()
		return applyConfigMap(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *corev1.Service:
		c := clientset.CoreV1().RESTClient()
		return applyService(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *policyv1beta1.PodSecurityPolicy:
		c := clientset.PolicyV1beta1().RESTClient()
		return applyPodSecurityPolicy(o, rev, c, applyParams{isNamespaced: false, forceConflicts: forceConflicts})
	case *networkingv1.NetworkPolicy:
		c := clientset.NetworkingV1().RESTClient()
		return applyNetworkPolicy(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *rbacv1.Role:
		c := clientset.RbacV1().RESTClient()
		return applyRole(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *rbacv1.RoleBinding:
		c := clientset.RbacV1().RESTClient()
		return applyRoleBinding(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *rbacv1.ClusterRole:
		c := clientset.RbacV1().RESTClient()
		return applyClusterRole(o, rev, c, applyParams{isNamespaced: false, forceConflicts: forceConflicts})
	case *rbacv1.ClusterRoleBinding:
		c := clientset.RbacV1().RESTClient()
		return applyClusterRoleBinding(o, rev, c, applyParams{isNamespaced: false, forceConflicts: forceConflicts})
	case *appsv1.Deployment:
		c := clientset.AppsV1().RESTClient()
		return applyDeployment(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *appsv1.DaemonSet:
		c := clientset.AppsV1().RESTClient()
		return applyDaemonSet(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *batchv1beta1.CronJob:
		c := clientset.BatchV1beta1().RESTClient()
		return applyCronJob(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	case *policyv1beta1.PodDisruptionBudget:
		c := clientset.PolicyV1beta1().RESTClient()
		return applyPodDisruptionBudget(o, rev, c, applyParams{isNamespaced: true, forceConflicts: forceConflicts})
	}
	return fmt.Errorf("unsupported type: %s", gvk.String())
}

// ParseResource parses YAML string.
func ParseResource(data []byte) (key string, jsonData []byte, err error) {
	obj, gvk, err := resourceDecoder.Decode(data, nil, nil)
	if err != nil {
		return "", nil, err
	}

	switch o := obj.(type) {
	case *corev1.Namespace:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Name, data, err
	case *corev1.ServiceAccount:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *corev1.ConfigMap:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *corev1.Service:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *policyv1beta1.PodSecurityPolicy:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Name, data, err
	case *networkingv1.NetworkPolicy:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *rbacv1.Role:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *rbacv1.RoleBinding:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *rbacv1.ClusterRole:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Name, data, err
	case *rbacv1.ClusterRoleBinding:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Name, data, err
	case *appsv1.Deployment:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *appsv1.DaemonSet:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *batchv1beta1.CronJob:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	case *policyv1beta1.PodDisruptionBudget:
		data, err := encodeToJSON(o)
		return o.Kind + "/" + o.Namespace + "/" + o.Name, data, err
	}

	return "", nil, fmt.Errorf("unsupported type: %s", gvk.String())
}

// ResourceDefinition represents a CKE-managed kubernetes resource.
type ResourceDefinition struct {
	Key        string
	Kind       Kind
	Namespace  string
	Name       string
	Revision   int64
	Image      string
	Definition []byte
}

// String implements fmt.Stringer.
func (d ResourceDefinition) String() string {
	return fmt.Sprintf("%s@%d", d.Key, d.Revision)
}

// NeedUpdate returns true if annotations of the current resource
// indicates need for update.
func (d ResourceDefinition) NeedUpdate(rs *ResourceStatus) bool {
	if rs == nil {
		return true
	}
	curRev, ok := rs.Annotations[AnnotationResourceRevision]
	if !ok {
		return true
	}
	if curRev != strconv.FormatInt(d.Revision, 10) {
		return true
	}

	if d.Image == "" {
		return false
	}

	curImage, ok := rs.Annotations[AnnotationResourceImage]
	if !ok {
		return true
	}
	return curImage != d.Image
}

// SortResources sort resources as defined order of creation.
func SortResources(res []ResourceDefinition) {
	less := func(i, j int) bool {
		a := res[i]
		b := res[j]
		if a.Kind != b.Kind {
			return a.Kind.Order() < b.Kind.Order()
		}
		switch i := strings.Compare(a.Namespace, b.Namespace); i {
		case -1:
			return true
		case 1:
			return false
		}
		switch i := strings.Compare(a.Name, b.Name); i {
		case -1:
			return true
		case 1:
			return false
		}
		// equal
		return false
	}

	sort.Slice(res, less)
}
