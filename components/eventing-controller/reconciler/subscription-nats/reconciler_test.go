package subscription_nats

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	bigTimeOut         = 40 * time.Second
	bigPollingInterval = 3 * time.Second
)

var _ = Describe("Subscription Reconciliation Tests", func() {
	var namespaceName string

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	BeforeEach(func() {
		namespaceName = "test"
		// we need to reset the http requests which the mock captured
		beb.Reset()
	})

	AfterEach(func() {
		// detailed request logs
		logf.Log.V(1).Info("beb requests", "number", len(beb.Requests))

		i := 0
		for req, payloadObject := range beb.Requests {
			reqDescription := fmt.Sprintf("method: %q, url: %q, payload object: %+v", req.Method, req.RequestURI, payloadObject)
			fmt.Printf("request[%d]: %s\n", i, reqDescription)
			i++
		}

		//// print all subscriptions in the namespace for debugging purposes
		//if err := printSubscriptions(namespaceName); err != nil {
		//	logf.Log.Error(err, "error while printing subscriptions")
		//}
	})

	When("Creating a Subscription with a valid Sink", func() {
		It("Should create a subscription in Nats and make it ready", func() {
			ctx := context.Background()
			subscriptionName := "sub-create-with-sink"

			// Ensuring subscriber svc
			_ = reconcilertesting.NewSubscriberSvc("test", namespaceName)
			//ensureSubscriberSvcCreated(subscriberSvc, ctx)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilterForNats, reconcilertesting.WithWebhookForNats)
			givenSubscription.Spec.Sink = "invalid"
			ensureSubscriptionCreated(givenSubscription, ctx)

			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
			))

			//By("Deleting the object to not provoke more reconciliation requests")
			//Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			////getSubscription(givenSubscription, ctx).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(Finalizer))

			//By("Emitting a Subscription deleted event")
			////var subscriptionEvents = v1.EventList{}
			//_ = v1.Event{
			//	Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionDeleted),
			//	Message: "",
			//	Type:    v1.EventTypeWarning,
			//}
			//getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionDeletedEvent))
		})
	})
})

func ensureSubscriptionCreated(subscription *eventingv1alpha1.Subscription, ctx context.Context) {

	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("failed to fetch subscription(%s): %v", lookupKey.String(), err)
			return eventingv1alpha1.Subscription{}
		}
		log.Printf("[Subscription] name:%s ns:%s apiRule:%s", subscription.Name, subscription.Namespace, subscription.Status.APIRuleName)
		return *subscription
	}, bigTimeOut, bigPollingInterval)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var beb *reconcilertesting.BebMock

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apigatewayv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	//client, err := client.New()
	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
	syncPeriod := time.Second
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:     scheme.Scheme,
		SyncPeriod: &syncPeriod,
	})
	Expect(err).ToNot(HaveOccurred())
	envConf := env.NatsConfig{
		Url: "",
	}
	err = NewReconciler(
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
