package subscription_nats

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/nats-io/nats.go"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	cache.Cache
	natsClient *handlers.Nats
	Log      logr.Logger
	recorder record.EventRecorder
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

func NewReconciler(client client.Client, cache cache.Cache, log logr.Logger, recorder record.EventRecorder,
	cfg env.NatsConfig) *Reconciler {
	natsClient := &handlers.Nats{
		Log: log,
	}
	natsClient.Initialize(cfg)
	return &Reconciler{
		Client:   client,
		Cache:    cache,
		natsClient: natsClient,
		Log:      log,
		recorder: recorder,
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	r.Log.Info("received subscription reconciliation request", "namespace", req.Namespace, "name",
		req.Name)

	actualSubscription := &eventingv1alpha1.Subscription{}
	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, actualSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()

	//Bind fields to logger
	log := r.Log.WithValues("kind", desiredSubscription.GetObjectKind().GroupVersionKind().Kind,
		"name", desiredSubscription.GetName(),
		"namespace", desiredSubscription.GetNamespace(),
		"version", desiredSubscription.GetGeneration(),
	)

	// examine DeletionTimestamp to determine if object is under deletion
	if desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			log.Info("Adding finalizer to subscription object")
			desiredSubscription.ObjectMeta.Finalizers = append(desiredSubscription.ObjectMeta.Finalizers, Finalizer)
			if err := r.Update(context.Background(), desiredSubscription); err != nil {
				return ctrl.Result{}, err
			}
			result.Requeue = true
		}

		if result.Requeue {
			return result, nil
		}

		// TODO refactor the common code between NATS reconciler and BEB reconciler
		//// sync the initial Subscription status
		//if err := r.syncInitialStatus(desiredSubscription, &result, ctx); err != nil {
		//	return ctrl.Result{}, errors.Wrap(err, "failed to sync status")
		//}
		//if result.Requeue {
		//	return result, nil
		//}

		//TODO Create subscription
		r.natsClient.Client.Subscribe("foo", func(msg *nats.Msg) {
			panic("Hey!")
		})

	} else {
		// The object is being deleted
		if containsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(desiredSubscription); err != nil {
				log.Info("fail to delete the external dependency of the subscription object")
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			log.Info("Removing finalizer from subscription object")
			desiredSubscription.ObjectMeta.Finalizers = removeString(desiredSubscription.ObjectMeta.Finalizers,
				Finalizer)
			if err := r.Update(context.Background(), desiredSubscription); err != nil {
				log.Info("Failed to remove finalizer from subscription object")
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// mark if the subscription status was changed
	//statusChanged := false
	//if statusChangedForBeb, err := r.syncBEBSubscription(desiredSubscription, &result, ctx, log, apiRule); err != nil {
	//	log.Error(err, "error while syncing BEB subscription")
	//	return ctrl.Result{}, err
	//} else {
	//	statusChanged = statusChanged || statusChangedForBeb
	//}
	return result, nil
}

func (r *Reconciler) deleteExternalResources(subscription *eventingv1alpha1.Subscription) error {
	//
	// delete any external resources associated with the subscription
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple types for same object.
	r.Log.Info("Deleting External Resources!!")
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// isInDeletion checks if the Subscription shall be deleted
func isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}

// syncInitialStatus determines the desires initial status and updates it accordingly (if conditions changed)
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context) error {
	currentStatus := subscription.Status
	expectedStatus := eventingv1alpha1.SubscriptionStatus{}
	expectedStatus.InitializeConditions()
	currentReadyStatusFromConditions := currentStatus.IsReady()

	var updateReadyStatus bool
	if currentReadyStatusFromConditions && !currentStatus.Ready {
		currentStatus.Ready = true
		updateReadyStatus = true
	} else if !currentReadyStatusFromConditions && currentStatus.Ready {
		currentStatus.Ready = false
		updateReadyStatus = true
	}
	// case: conditions are already initialized
	if len(currentStatus.Conditions) >= len(expectedStatus.Conditions) && !updateReadyStatus {
		return nil
	}
	if len(currentStatus.Conditions) == 0 {
		subscription.Status = expectedStatus
	} else {
		subscription.Status.Ready = currentStatus.Ready
	}
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}
	result.Requeue = true
	return nil
}

func (r Reconciler) deleteNATSSubscription(subscription *eventingv1alpha1.Subscription, logger logr.Logger,
	ctx context.Context) error {
	//TODO
	r.natsClient.DeleteSubscription(nil)
	panic("Implement me")
	return nil
}

// syncBEBSubscription delegates the subscription synchronization to the backend client. It returns true if the subscription status was changed.
func (r *Reconciler) syncNATSSubscription(subscription *eventingv1alpha1.Subscription, result *ctrl.Result,
	ctx context.Context, logger logr.Logger) (bool, error) {
	logger.Info("Syncing subscription with NATS")

	// if object is marked for deletion, we need to delete the BEB subscription
	if isInDeletion(subscription) {
		return false, r.deleteNATSSubscription(subscription, logger, ctx)
	}

	var statusChanged bool
	//var err error
	//if statusChanged, err = r.natsClient.SyncBebSubscription(subscription, apiRule); err != nil {
	//	logger.Error(err, "Update BEB subscription failed")
	//	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse)
	//	if err := r.updateCondition(subscription, condition, ctx); err != nil {
	//		return statusChanged, err
	//	}
	//	return false, err
	//}
	//
	//if !subscription.Status.IsConditionSubscribed() {
	//	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue)
	//	if err := r.updateCondition(subscription, condition, ctx); err != nil {
	//		return statusChanged, err
	//	}
	//	statusChanged = true
	//}
	//
	//statusChangedAtCheck, retry, errTimeout := r.checkStatusActive(subscription)
	//statusChanged = statusChanged || statusChangedAtCheck
	//if errTimeout != nil {
	//	logger.Error(errTimeout, "timeout at retry")
	//	result.Requeue = false
	//	return statusChanged, errTimeout
	//}
	//if retry {
	//	logger.Info("Wait for subscription to be active", "name:", subscription.Name, "status:", subscription.Status.EmsSubscriptionStatus.SubscriptionStatus)
	//	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, corev1.ConditionFalse)
	//	if err := r.updateCondition(subscription, condition, ctx); err != nil {
	//		return statusChanged, err
	//	}
	//	result.RequeueAfter = time.Second * 1
	//} else if statusChanged {
	//	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, corev1.ConditionTrue)
	//	if err := r.updateCondition(subscription, condition, ctx); err != nil {
	//		return statusChanged, err
	//	}
	//}
	//// OK
	return statusChanged, nil
}