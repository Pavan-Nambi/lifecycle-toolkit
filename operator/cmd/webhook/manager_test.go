package webhook

//func TestCreateOptions(t *testing.T) {
//	t.Run("implements interface", func(t *testing.T) {
//		var provider cmdManager.Provider = NewWebhookManagerProvider("certs-dir", "key-file", "cert-file")
//		_, _ = provider.CreateManager("namespace", scheme.Scheme, &rest.Config{})
//
//		providerImpl := provider.(WebhookProvider)
//		assert.Equal(t, "certs-dir", providerImpl.certificateDirectory)
//		assert.Equal(t, "key-file", providerImpl.keyFileName)
//		assert.Equal(t, "cert-file", providerImpl.certificateFileName)
//	})
//	t.Run("creates options", func(t *testing.T) {
//		provider := WebhookProvider{}
//		options := provider.createOptions(scheme.Scheme, "test-namespace")
//
//		assert.NotNil(t, options)
//		assert.Equal(t, "test-namespace", options.Namespace)
//		assert.Equal(t, scheme.Scheme, options.Scheme)
//		assert.Equal(t, metricsBindAddress, options.MetricsBindAddress)
//		assert.Equal(t, port, options.Port)
//	})
//	t.Run("configures webhooks server", func(t *testing.T) {
//		provider := NewWebhookManagerProvider("certs-dir", "key-file", "cert-file")
//		expectedWebhookServer := &webhook.Server{}
//
//		mgr := &cmdManager.MockManager{}
//		mgr.On("GetWebhookServer").Return(expectedWebhookServer)
//
//		mgrWithWebhookServer := provider.SetupWebhookServer(mgr)
//
//		assert.Equal(t, "certs-dir", expectedWebhookServer.CertDir)
//		assert.Equal(t, "key-file", expectedWebhookServer.KeyName)
//		assert.Equal(t, "cert-file", expectedWebhookServer.CertName)
//
//		mgrWebhookServer := mgrWithWebhookServer.GetWebhookServer()
//		assert.Equal(t, "certs-dir", mgrWebhookServer.CertDir)
//		assert.Equal(t, "key-file", mgrWebhookServer.KeyName)
//		assert.Equal(t, "cert-file", mgrWebhookServer.CertName)
//	})
//}
