package keptnwebhookcontroller

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/keptn/lifecycle-toolkit/klc-cert-manager/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testDomain         = ServiceName + "." + namespace + ".svc"
	expectedSecretName = DeploymentName + secretPostfix
	strategyWebhook    = "webhook"
	testBytes          = 123
)

func TestReconcileCertificate_Create(t *testing.T) {
	clt := prepareFakeClient(false, false)
	controller, request := prepareController(t, clt)

	res, err := controller.Reconcile(context.TODO(), request)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, SuccessDuration, res.RequeueAfter)

	secret := &corev1.Secret{}
	err = clt.Get(context.TODO(), client.ObjectKey{Name: expectedSecretName, Namespace: namespace}, secret)
	require.NoError(t, err)

	assert.NotNil(t, secret.Data)
	assert.NotEmpty(t, secret.Data)
	assert.Contains(t, secret.Data, RootKey)
	assert.Contains(t, secret.Data, RootCert)
	assert.Contains(t, secret.Data, RootCertOld)
	assert.Contains(t, secret.Data, ServerKey)
	assert.Contains(t, secret.Data, ServerCert)
	assert.NotNil(t, secret.Data[RootCert])
	assert.NotEmpty(t, secret.Data[RootCert])
	assert.Empty(t, secret.Data[RootCertOld])

	verifyCertificates(t, secret, clt, false)
}

func TestReconcileCertificate_Update(t *testing.T) {
	clt := prepareFakeClient(true, false)
	controller, request := prepareController(t, clt)

	res, err := controller.Reconcile(context.TODO(), request)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, SuccessDuration, res.RequeueAfter)

	secret := &corev1.Secret{}
	err = clt.Get(context.TODO(), client.ObjectKey{Name: expectedSecretName, Namespace: namespace}, secret)
	require.NoError(t, err)

	assert.NotNil(t, secret.Data)
	assert.NotEmpty(t, secret.Data)
	assert.Contains(t, secret.Data, RootKey)
	assert.Contains(t, secret.Data, RootCert)
	assert.Contains(t, secret.Data, RootCertOld)
	assert.Contains(t, secret.Data, ServerKey)
	assert.Contains(t, secret.Data, ServerCert)
	assert.NotNil(t, secret.Data[RootCert])
	assert.NotEmpty(t, secret.Data[RootCert])
	assert.Equal(t, []byte{testBytes}, secret.Data[RootCertOld])

	verifyCertificates(t, secret, clt, true)
}

func TestReconcileCertificate_ExistingSecretWithValidCertificate(t *testing.T) {
	clt := prepareFakeClient(true, true)
	controller, request := prepareController(t, clt)

	res, err := controller.Reconcile(context.TODO(), request)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, SuccessDuration, res.RequeueAfter)

	secret := &corev1.Secret{}
	err = clt.Get(context.TODO(), client.ObjectKey{Name: expectedSecretName, Namespace: namespace}, secret)
	require.NoError(t, err)

	verifyCertificates(t, secret, clt, false)
}

func TestReconcile(t *testing.T) {

	crd1 := &apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "crd1",
			Labels: map[string]string{
				"crdGroup": "lifecycle.keptn.sh",
			},
		},
		Spec: apiv1.CustomResourceDefinitionSpec{
			Group: crdGroup,
			Conversion: &apiv1.CustomResourceConversion{
				Strategy: strategyWebhook,
				Webhook: &apiv1.WebhookConversion{
					ClientConfig: &apiv1.WebhookClientConfig{},
				},
			},
		},
		Status: apiv1.CustomResourceDefinitionStatus{},
	}

	crd2 := &apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "crd2",
		},
		Spec: apiv1.CustomResourceDefinitionSpec{
			Group: "Someonelese",
			Conversion: &apiv1.CustomResourceConversion{
				Strategy: strategyWebhook,
				Webhook: &apiv1.WebhookConversion{
					ClientConfig: &apiv1.WebhookClientConfig{
						CABundle: []byte("my unmodified bundle"),
					},
				},
			},
		},
	}
	crd3 := &apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "crd3",
			Labels: map[string]string{
				"crdGroup": "lifecycle.keptn.sh",
			},
		},
		Spec: apiv1.CustomResourceDefinitionSpec{
			Group: crdGroup,
		},
	}

	t.Run(`reconcile successfully with mutatingwebhookconfiguration`, func(t *testing.T) {
		fakeClient := fake.NewClient(crd1, crd2, crd3, &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: Webhookconfig,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					ClientConfig: admissionregistrationv1.WebhookClientConfig{},
				},
				{
					ClientConfig: admissionregistrationv1.WebhookClientConfig{},
				},
			},
		})

		controller, request := prepareController(t, fakeClient)
		result, err := controller.Reconcile(context.TODO(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
	t.Run(`update crd successfully with up-to-date secret`, func(t *testing.T) {
		fakeClient := fake.NewClient(crd1, crd2, crd3)
		cs := newCertificateSecret()
		_ = cs.setSecretFromReader(context.TODO(), fakeClient, namespace, testr.New(t))
		_ = cs.validateCertificates(namespace)
		_ = cs.createOrUpdateIfNecessary(context.TODO(), fakeClient)

		controller, request := prepareController(t, fakeClient)
		result, err := controller.Reconcile(context.TODO(), request)
		require.NoError(t, err)
		assert.NotNil(t, result)

		expectedBundle, err := cs.loadCombinedBundle()
		require.NoError(t, err)
		actualCrd := &apiv1.CustomResourceDefinition{}

		// crd1 should get a new secret
		err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: crd1.Name}, actualCrd)
		require.NoError(t, err)
		assert.Equal(t, expectedBundle, actualCrd.Spec.Conversion.Webhook.ClientConfig.CABundle)

		// crd2 is not a keptn resource and should not be touched
		err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: crd2.Name}, actualCrd)
		require.NoError(t, err)
		assert.Equal(t, crd2.Spec.Conversion.Webhook.ClientConfig.CABundle, actualCrd.Spec.Conversion.Webhook.ClientConfig.CABundle)

		// crd 3 should not have a webhook conversion
		err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: crd3.Name}, actualCrd)
		require.NoError(t, err)
		assert.Empty(t, actualCrd.Spec.Conversion.Webhook)
	})
	// Generation must not be skipped because webhook startup routine listens for the secret
	// See cmd/operator/manager.go and cmd/operator/watcher.go
	t.Run(`do not skip certificates generation if no configuration exists`, func(t *testing.T) {
		fakeClient := fake.NewClient()
		controller, request := prepareController(t, fakeClient)
		result, err := controller.Reconcile(context.TODO(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		secret := &corev1.Secret{}
		err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: expectedSecretName, Namespace: namespace}, secret)
		assert.NoError(t, err)
	})
}

func prepareFakeClient(withSecret bool, generateValidSecret bool) client.Client {
	objs := []client.Object{
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: Webhookconfig,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					ClientConfig: admissionregistrationv1.WebhookClientConfig{},
				},
				{
					ClientConfig: admissionregistrationv1.WebhookClientConfig{},
				},
			},
		},

		&apiv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mycrd1",
			},
			Spec: apiv1.CustomResourceDefinitionSpec{
				Group: crdGroup,
				Conversion: &apiv1.CustomResourceConversion{
					Strategy: strategyWebhook,
					Webhook: &apiv1.WebhookConversion{
						ClientConfig: &apiv1.WebhookClientConfig{},
					},
				},
			},
		},
		&apiv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mycrd2",
			},
			Spec: apiv1.CustomResourceDefinitionSpec{
				Group: "whatever",
				Conversion: &apiv1.CustomResourceConversion{
					Strategy: strategyWebhook,
					Webhook: &apiv1.WebhookConversion{
						ClientConfig: &apiv1.WebhookClientConfig{
							CABundle: []byte("myunmodifiedbundle"),
						},
					},
				},
			},
		},
		&apiv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mycrd3",
			},
			Spec: apiv1.CustomResourceDefinitionSpec{
				Group: crdGroup,
			},
		},
	}
	if withSecret {
		certData := createInvalidTestCertData(nil)
		if generateValidSecret {
			certData = createValidTestCertData(nil)
		}

		objs = append(objs,
			createTestSecret(nil, certData),
		)
	}

	faker := fake.NewClient(objs...)
	return faker
}

func createInvalidTestCertData(_ *testing.T) map[string][]byte {
	return map[string][]byte{
		RootKey:    {testBytes},
		RootCert:   {testBytes},
		ServerKey:  {testBytes},
		ServerCert: {testBytes},
	}
}

func createValidTestCertData(_ *testing.T) map[string][]byte {
	cert := Certs{
		Domain: testDomain,
		Now:    time.Now(),
	}
	_ = cert.ValidateCerts()
	return cert.Data
}

func createTestSecret(_ *testing.T, certData map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      expectedSecretName,
		},
		Data: certData,
	}
}

func prepareController(t *testing.T, clt client.Client) (*KeptnWebhookCertificateReconciler, reconcile.Request) {
	rec := &KeptnWebhookCertificateReconciler{
		ctx:       context.TODO(),
		Client:    clt,
		ApiReader: clt,
		Log:       testr.New(t),
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      DeploymentName,
			Namespace: namespace,
		},
	}

	return rec, request
}

func testWebhookClientConfig(
	t *testing.T, webhookClientConfig *admissionregistrationv1.WebhookClientConfig,
	secretData map[string][]byte, isUpdate bool) {
	require.NotNil(t, webhookClientConfig)
	require.NotEmpty(t, webhookClientConfig.CABundle)

	expectedCert := secretData[RootCert]
	if isUpdate {
		expectedCert = append(expectedCert, []byte{123}...)
	}
	assert.Equal(t, expectedCert, webhookClientConfig.CABundle)
}

func verifyCertificates(t *testing.T, secret *corev1.Secret, clt client.Client, isUpdate bool) {
	cert := Certs{
		Domain:  getDomain(namespace),
		Data:    secret.Data,
		SrcData: secret.Data,
		Now:     time.Now(),
	}

	// validateRootCerts and validateServerCerts return false if the certificates are valid
	assert.False(t, cert.validateRootCerts(time.Now()))
	assert.False(t, cert.validateServerCerts(time.Now()))

	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{}
	err := clt.Get(context.TODO(), client.ObjectKey{
		Name: Webhookconfig,
	}, mutatingWebhookConfig)
	require.NoError(t, err)
	assert.Len(t, mutatingWebhookConfig.Webhooks, 2)
	testWebhookClientConfig(t, &mutatingWebhookConfig.Webhooks[0].ClientConfig, secret.Data, isUpdate)

}
