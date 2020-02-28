package testing

import (
	"fmt"
	"testing"

	istioauthenticationalpha1 "istio.io/api/authentication/v1alpha1"
	istiosecuritybeta1 "istio.io/api/security/v1beta1"
	istiov1beta1 "istio.io/api/type/v1beta1"
	istiov1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	istiosecurityv1alpha1 "istio.io/client-go/pkg/apis/security/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	FakeChannelName      = "fake-chan"
	FakeSubscriptionName = "fake-sub"
	// brokerTargetSelectorName used for targeting the default-broker svc while creating an istio policy
	brokerTargetSelectorName = "default-broker"
	// filterTargetSelectorName used for targeting the default-broker-filter svc while creating an istio policy
	filterTargetSelectorName = "default-broker-filter"

	istioMtlsPermissiveMode = 1
)

// redefine here to avoid cyclic dependency
const (
	integrationNamespace                      = "kyma-integration"
	applicationNameLabelKey                   = "application-name"
	brokerNamespaceLabelKey                   = "broker-namespace"
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"
	knSubscriptionNamePrefix                  = "brokersub"
)

func NewAppSubscription(appNs, appName string, opts ...SubscriptionOption) *messagingv1alpha1.Subscription {
	sub := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", knSubscriptionNamePrefix),
			Namespace:    integrationNamespace,
			Labels: map[string]string{
				brokerNamespaceLabelKey: appNs,
				applicationNameLabelKey: appName,
			},
		},
	}

	for _, opt := range opts {
		opt(sub)
	}

	return sub
}

// SubscriptionOption is a functional option for Subscription objects.
type SubscriptionOption func(*messagingv1alpha1.Subscription)

// WithSpec sets the spec of a Subscription.
func WithSpec(t *testing.T, subscriberURI string) SubscriptionOption {
	url, err := apis.ParseURL(subscriberURI)
	if err != nil {
		t.Fatalf("error while parsing url: %v, error: %v", subscriberURI, err)
	}
	return func(s *messagingv1alpha1.Subscription) {
		s.Spec = messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name: FakeChannelName,
			},
			Subscriber: &duckv1.Destination{
				URI: url,
			},
		}
	}
}

// WithNameSuffix generates the name of a Subscription using its GenerateName prefix.
func WithNameSuffix(nameSuffix string) SubscriptionOption {
	return func(s *messagingv1alpha1.Subscription) {
		if s.GenerateName != "" && s.Name == "" {
			s.Name = s.GenerateName + nameSuffix
		}
	}
}

func NewAppNamespace(name string, brokerInjection bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if brokerInjection {
		ns.Labels = map[string]string{
			knativeEventingInjectionLabelKey: knativeEventingInjectionLabelValueEnabled,
		}
	}
	return ns
}

func NewDefaultBroker(ns string) *eventingv1alpha1.Broker {
	return &eventingv1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: ns,
		},
	}
}

func NewIstioPolicy(ns, policyName string) *istiov1alpha1.Policy {
	labels := make(map[string]string)
	labels["eventing.knative.dev/broker"] = "default"

	brokerTargetSelector := &istioauthenticationalpha1.TargetSelector{
		Name: brokerTargetSelectorName,
	}
	filterTargetSelector := &istioauthenticationalpha1.TargetSelector{
		Name: filterTargetSelectorName,
	}
	mtls := &istioauthenticationalpha1.MutualTls{
		Mode: istioMtlsPermissiveMode,
	}
	peerAuthenticationMethod := &istioauthenticationalpha1.PeerAuthenticationMethod{
		Params: &istioauthenticationalpha1.PeerAuthenticationMethod_Mtls{
			Mtls: mtls,
		},
	}

	return &istiov1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: istioauthenticationalpha1.Policy{
			Targets: []*istioauthenticationalpha1.TargetSelector{brokerTargetSelector, filterTargetSelector},
			Peers:   []*istioauthenticationalpha1.PeerAuthenticationMethod{peerAuthenticationMethod},
		},
	}
}

func NewBrokerIngressIstioAuthorization(namespace string) *istiosecurityv1alpha1.AuthorizationPolicy {
	policyNamespace := namespace
	brokerRole := "ingress"
	policyName := "broker-ingress"
	natssDispatcherPrincipal := fmt.Sprintf("cluster.local/ns/%s/sa/natss-ch-dispatcher", "knative-eventing")
	principals := []string{natssDispatcherPrincipal}
	allowedMethods := []string{"POST"}
	allowedPaths := []string{"/"}

	matchLabels := map[string]string{
		"eventing.knative.dev/brokerRole": brokerRole,
	}

	labels := map[string]string{
		"app": policyName,
	}

	rulesFrom := []*istiosecuritybeta1.Rule_From{{
		Source: &istiosecuritybeta1.Source{
			Principals: principals,
		},
	}}
	rulesTo := []*istiosecuritybeta1.Rule_To{{
		Operation: &istiosecuritybeta1.Operation{
			Methods: allowedMethods,
			Paths:   allowedPaths,
		},
	}}

	rules := []*istiosecuritybeta1.Rule{{
		From: rulesFrom,
		To:   rulesTo,
	}}

	istioAuthorizationPolicy := &istiosecurityv1alpha1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      policyName,
			Namespace: policyNamespace,
			Labels:    labels,
		},
		Spec: istiosecuritybeta1.AuthorizationPolicy{
			Selector: &istiov1beta1.WorkloadSelector{
				MatchLabels: matchLabels,
			},
			Rules: rules,
		},
	}

	return istioAuthorizationPolicy
}

func NewBrokerFilterIstioAuthorization(namespace string) *istiosecurityv1alpha1.AuthorizationPolicy {
	policyNamespace := namespace
	brokerRole := "filter"
	policyName := "broker-filter"
	natssDispatcherPrincipal := fmt.Sprintf("cluster.local/ns/%s/sa/natss-ch-dispatcher", "knative-eventing")
	principals := []string{fmt.Sprintf("cluster.local/ns/%s/sa/eventing-broker-ingress", namespace),
		natssDispatcherPrincipal}
	allowedPaths := []string{fmt.Sprintf("/triggers/%s/*", namespace)}
	allowedMethods := []string{"POST"}

	matchLabels := map[string]string{
		"eventing.knative.dev/brokerRole": brokerRole,
	}

	labels := map[string]string{
		"app": policyName,
	}

	rulesFrom := []*istiosecuritybeta1.Rule_From{{
		Source: &istiosecuritybeta1.Source{
			Principals: principals,
		},
	}}
	rulesTo := []*istiosecuritybeta1.Rule_To{{
		Operation: &istiosecuritybeta1.Operation{
			Methods: allowedMethods,
			Paths:   allowedPaths,
		},
	}}

	rules := []*istiosecuritybeta1.Rule{{
		From: rulesFrom,
		To:   rulesTo,
	}}

	istioAuthorizationPolicy := &istiosecurityv1alpha1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      policyName,
			Namespace: policyNamespace,
			Labels:    labels,
		},
		Spec: istiosecuritybeta1.AuthorizationPolicy{
			Selector: &istiov1beta1.WorkloadSelector{
				MatchLabels: matchLabels,
			},
			Rules: rules,
		},
	}

	return istioAuthorizationPolicy
}

func NewAppChannel(appName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FakeChannelName,
			Namespace: integrationNamespace,
			Labels: map[string]string{
				applicationNameLabelKey: appName,
			},
		},
	}
}
