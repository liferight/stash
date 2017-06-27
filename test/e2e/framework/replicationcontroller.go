package framework

import (
	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (f *Framework) ReplicationController() apiv1.ReplicationController {
	podTemplate := f.PodTemplate()
	return apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("stash"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": "stash-e2e",
			},
		},
		Spec: apiv1.ReplicationControllerSpec{
			Replicas: types.Int32P(1),
			Template: &podTemplate,
		},
	}
}

func (f *Framework) CreateReplicationController(obj apiv1.ReplicationController) error {
	_, err := f.kubeClient.CoreV1().ReplicationControllers(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) DeleteReplicationController(meta metav1.ObjectMeta) error {
	return f.kubeClient.CoreV1().ReplicationControllers(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyReplicationController(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *apiv1.ReplicationController {
		obj, err := f.kubeClient.CoreV1().ReplicationControllers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}