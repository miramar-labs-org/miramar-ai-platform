package validator

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// portForward opens a client-go SPDY port-forward to podName, mirroring what
// `kubectl port-forward` does internally — this only works against Pods, not
// Services. It returns the local port chosen by the OS and a stop function.
func portForward(restConfig *rest.Config, clientset kubernetes.Interface, namespace, podName string, remotePort int) (int, func(), error) {
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return 0, nil, err
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", remotePort)}, stopCh, readyCh, io.Discard, io.Discard)
	if err != nil {
		return 0, nil, err
	}

	errCh := make(chan error, 1)
	go func() { errCh <- fw.ForwardPorts() }()

	select {
	case <-readyCh:
	case err := <-errCh:
		return 0, nil, fmt.Errorf("port-forward failed: %w", err)
	case <-time.After(30 * time.Second):
		close(stopCh)
		return 0, nil, fmt.Errorf("port-forward did not become ready in time")
	}

	ports, err := fw.GetPorts()
	if err != nil {
		close(stopCh)
		return 0, nil, err
	}

	return int(ports[0].Local), func() { close(stopCh) }, nil
}
