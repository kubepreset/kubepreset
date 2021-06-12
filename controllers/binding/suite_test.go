/*
Copyright 2020 The KubePreset Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package binding_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	bindingv1beta1 "github.com/kubepreset/kubepreset/apis/binding/v1beta1"
	bindingcontrollers "github.com/kubepreset/kubepreset/controllers/binding"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const timeout = time.Minute * 2

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var k8sManager manager.Manager

//logLevel hold the current log level
var logLevel zapcore.Level

func initializeLogLevel() {
	logLvl := os.Getenv("LOG_LEVEL")
	logLvl = strings.ToUpper(logLvl)
	switch {
	case logLvl == "TRACE":
		logLevel = -2
	case logLvl == "DEBUG":
		logLevel = -1
	default:
		logLevel = 0
	}
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	initializeLogLevel()

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	level := uzap.NewAtomicLevelAt(logLevel)
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.Level(&level)))

	By("bootstrapping test environment")
	useExistingCluster := true
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:  []string{filepath.Join("..", "..", "config", "crd", "bases")},
		UseExistingCluster: &useExistingCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = apixv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = bindingv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&bindingcontrollers.ServiceBindingReconciler{
		Client: k8sManager.GetClient(),
		Log:    ctrl.Log.WithName("bindingcontrollers.servicebinding").WithName("ServiceBinding"),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	done := make(chan interface{})
	go func() {
		k8sClient = k8sManager.GetClient()
		Expect(k8sClient).ToNot(BeNil())
		close(done) //signifies the code is done
	}()
	Eventually(done, timeout).Should(BeClosed())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
