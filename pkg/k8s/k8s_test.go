package k8s

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/admissionregistration/v1beta1"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/api/core/v1"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	oamapi "github.com/oam-dev/kubevela-core-api/apis/core.oam.dev"
)

const (
	testWebhookName = "c7c95710-d8c3-4cc3-a2a8-8d2b46909c76"
	testSecretName  = "15906410-af2a-4f9b-8a2d-c08ffdd5e129"
	testNamespace   = "7cad5f92-c0d5-4bc9-87a3-6f44d5a5619d"
)

var (
	fail   = admissionv1beta1.Fail
	ignore = admissionv1beta1.Ignore
)

func genSecretData() (ca, cert, key []byte) {
	ca = make([]byte, 4)
	cert = make([]byte, 4)
	key = make([]byte, 4)
	rand.Read(cert)
	rand.Read(key)
	return
}

func newTestSimpleK8s() *k8s {
	Scheme = k8sruntime.NewScheme()
	_ = crdv1.AddToScheme(Scheme)
	_ = oamapi.AddToScheme(Scheme)
	return &k8s{
		clientset: fake.NewSimpleClientset(),
		client:    ctlfake.NewClientBuilder().WithScheme(Scheme).Build(),
	}
}

func TestGetCaFromCertificate(t *testing.T) {
	k := newTestSimpleK8s()

	ca, cert, key := genSecretData()

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: testSecretName,
		},
		Data: map[string][]byte{"ca": ca, "cert": cert, "key": key},
	}

	k.clientset.CoreV1().Secrets(testNamespace).Create(context.Background(), secret, metav1.CreateOptions{})

	retrievedCa := k.GetCaFromSecret(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCa, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestSaveCertsToSecret(t *testing.T) {
	k := newTestSimpleK8s()

	ca, cert, key := genSecretData()

	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", ca, cert, key)

	secret, _ := k.clientset.CoreV1().Secrets(testNamespace).Get(context.Background(), testSecretName, metav1.GetOptions{})

	if !bytes.Equal(secret.Data["cert"], cert) {
		t.Error("'cert' saved data does not match retrieved")
	}

	if !bytes.Equal(secret.Data["key"], key) {
		t.Error("'key' saved data does not match retrieved")
	}
}

func TestSaveThenLoadSecret(t *testing.T) {
	k := newTestSimpleK8s()
	ca, cert, key := genSecretData()
	k.SaveCertsToSecret(testSecretName, testNamespace, "cert", "key", ca, cert, key)
	retrievedCert := k.GetCaFromSecret(testSecretName, testNamespace)
	if !bytes.Equal(retrievedCert, ca) {
		t.Error("Was not able to retrieve CA information that was saved")
	}
}

func TestPatchWebhookConfigurations(t *testing.T) {
	k := newTestSimpleK8s()

	ca, _, _ := genSecretData()

	k.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Create(context.Background(), &v1beta1.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []v1beta1.MutatingWebhook{{Name: "m1"}, {Name: "m2"}}}, metav1.CreateOptions{})

	k.clientset.
		AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Create(context.Background(), &v1beta1.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: testWebhookName,
			},
			Webhooks: []v1beta1.ValidatingWebhook{{Name: "v1"}, {Name: "v2"}}}, metav1.CreateOptions{})

	// create crd step for fake client query
	var crds []string
	crds = append(crds, "applications.core.oam.dev")
	PolicyType := admissionv1beta1.FailurePolicyType("ignore")
	patchNamespace := "test-vela-ns"

	var crdObjectByte []byte
	crdObjectByte, err := ioutil.ReadFile("./test.crd.applications.yaml")
	if err != nil {
		log.WithField("err", err).Fatal("failed to read crd yaml file")
	}

	var crdObject crdv1.CustomResourceDefinition
	err = yaml.Unmarshal(crdObjectByte, &crdObject)
	if err != nil {
		log.WithField("err", err).Fatal("failed to unmarshal yaml file")
	}

	if err := k.client.Create(context.Background(), &crdObject); err != nil {
		log.WithField("err", err).Fatal("failed to generate CRD")
	}

	k.PatchWebhookConfigurations(testWebhookName, ca, &PolicyType, true, true, patchNamespace, crds)

	//  crd check step
	var crd = "applications.core.oam.dev"
	err = k.client.Get(context.TODO(), client.ObjectKey{Name: crd}, &crdObject)
	if err != nil {
		log.WithField("err", err).Fatal("failed to get CRD")
	}

	if crdObject.Spec.Conversion.Webhook.ClientConfig.Service.Namespace != patchNamespace {
		t.Error("patch namespace does not match")
	}
	if crdObject.Annotations["cert-manager.io/inject-ca-from"] != patchNamespace+"/kubevela-vela-core-root-cert" {
		t.Error("patch annotations does not match")
	}

	whmut, err := k.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Get(context.Background(), testWebhookName, metav1.GetOptions{})

	if err != nil {
		t.Error(err)
	}

	whval, err := k.clientset.
		AdmissionregistrationV1beta1().
		MutatingWebhookConfigurations().
		Get(context.Background(), testWebhookName, metav1.GetOptions{})

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(whmut.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from first mutating webhook configuration does not match")
	}
	if !bytes.Equal(whmut.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from second mutating webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[0].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from first validating webhook configuration does not match")
	}
	if !bytes.Equal(whval.Webhooks[1].ClientConfig.CABundle, ca) {
		t.Error("Ca retrieved from second validating webhook configuration does not match")
	}
	if whmut.Webhooks[0].FailurePolicy == nil {
		t.Errorf("Expected first mutating webhook failure policy to be set to %s", fail)
	}
	if whmut.Webhooks[1].FailurePolicy == nil {
		t.Errorf("Expected second mutating webhook failure policy to be set to %s", fail)
	}
	if whval.Webhooks[0].FailurePolicy == nil {
		t.Errorf("Expected first validating webhook failure policy to be set to %s", fail)
	}
	if whval.Webhooks[1].FailurePolicy == nil {
		t.Errorf("Expected second validating webhook failure policy to be set to %s", fail)
	}

}
