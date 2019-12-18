package scenario

import (
	"crypto/tls"
	"fmt"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"
)

// E2E executes complete external solution integration test scenario
type E2EEventMesh struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
}


// AddFlags adds CLI flags to given FlagSet
func (s *E2EEventMesh) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-mesh", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
}

// Steps return scenario steps
func (s *E2EEventMesh) Steps(config *rest.Config) ([]step.Step, error) {
	appOperatorClientset := appOperatorClient.NewForConfigOrDie(config)
	appBrokerClientset := appBrokerClient.NewForConfigOrDie(config)
	kubelessClientset := kubelessClient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	pods := coreClientset.CoreV1().Pods(s.testID)
	eventingClientset := eventingClient.NewForConfigOrDie(config)
	serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(config)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	knativeBrokerClientSet := eventing.NewForConfigOrDie(config)

	connector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(s.skipSSLVerify),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		gatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, lambdaPort)
	state := s.NewState()

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, coreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant,
				s.applicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(s.testID)),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, lambdaPort, kubelessClientset.KubelessV1beta1().Functions(s.testID), pods),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
		),
		testsuite.NewRegisterTestService(s.testID, testService, state),
		testsuite.NewCreateServiceInstance(s.testID,
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID),
			state,
		),
		testsuite.NewCreateServiceBinding(s.testID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID), state),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID), state,
			knativeBrokerClientSet.EventingV1alpha1().Brokers(s.testID)),
		testsuite.NewCreateSubscription(s.testID, s.testID, lambdaEndpoint, eventingClientset.EventingV1alpha1().Subscriptions(s.testID)),
		testsuite.NewSendEventToMesh(s.testID, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}

type e2EEventMeshState struct {
	domain        string
	skipSSLVerify bool
	appName       string

	serviceClassID      string
	serviceInstanceName string
	registryClient      *testkit.RegistryClient
	eventSender         *testkit.EventSender
}

func (s *E2EEventMesh) NewState() *e2EEventMeshState {
	return &e2EEventMeshState{domain: s.domain, skipSSLVerify: s.skipSSLVerify, appName: s.testID}
}

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *e2EEventMeshState) SetServiceClassID(serviceID string) {
	s.serviceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *e2EEventMeshState) GetServiceClassID() string {
	return s.serviceClassID
}

// SetServiceInstanceName allows to set ServiceInstanceName so it can be shared between steps
func (s *e2EEventMeshState) SetServiceInstanceName(serviceID string) {
	s.serviceInstanceName = serviceID
}

// GetServiceInstanceName allows to get ServiceInstanceName so it can be shared between steps
func (s *e2EEventMeshState) GetServiceInstanceName() string {
	return s.serviceInstanceName
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EEventMeshState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.skipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.domain, s.appName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHTTPClient)
	s.eventSender = testkit.NewEventSender(resilientHTTPClient, s.domain)
}

// GetRegistryClient returns connected RegistryClient
func (s *e2EEventMeshState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

// GetEventSender returns connected EventSender
func (s *e2EEventMeshState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}
