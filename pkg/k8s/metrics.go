package k8s

import (
	"fmt"
	"io"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
)

// AdminHTTPPortNameSuffix is the suffix for ports used by admin HTTP servers.
// Ports may be named <container>-admin to avoid conflicts across multiple
// containers.
const AdminHTTPPortNameSuffix string = "-admin"

// GetContainerMetrics returns the metrics exposed by a container on the passed in portName
func GetContainerMetrics(
	k8sAPI *KubernetesAPI,
	pod corev1.Pod,
	container corev1.Container,
	emitLogs bool,
	portName string,
) ([]byte, error) {
	portForward, err := NewContainerMetricsForward(k8sAPI, pod, container, emitLogs, portName)
	if err != nil {
		return nil, err
	}

	defer portForward.Stop()
	if err = portForward.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running port-forward: %s\n", err)
		return nil, err
	}

	metricsURL := portForward.URLFor("/metrics")
	return getResponse(metricsURL)
}

// getResponse makes a http Get request to the passed url and returns the response/error
func getResponse(url string) ([]byte, error) {
	// url has been constructed by k8s.newPortForward and is not passed in by
	// the user.
	//nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
