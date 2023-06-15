//go:build e2e
// +build e2e

package multiple_hosts_test

import (
	"fmt"
	"testing"

	"github.com/tj/assert"
	"k8s.io/client-go/kubernetes"

	. "github.com/kedacore/http-add-on/tests/helper"
)

const (
	testName = "multiple-hosts-test"
)

var (
	testNamespace        = fmt.Sprintf("%s-ns", testName)
	deploymentName       = fmt.Sprintf("%s-deployment", testName)
	serviceName          = fmt.Sprintf("%s-service", testName)
	httpScaledObjectName = fmt.Sprintf("%s-http-so", testName)
	host1                = fmt.Sprintf("%s-host1", testName)
	host2                = fmt.Sprintf("%s-host2", testName)
	initReplicaCount     = 0
	minReplicaCount      = 1
	maxReplicaCount      = 2
)

type templateData struct {
	TestNamespace        string
	DeploymentName       string
	ServiceName          string
	HTTPScaledObjectName string
	Host                 string
	Host1                string
	Host2                string
	Replicas             int
	MinReplicas          int
	MaxReplicas          int
}

const (
	serviceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{.ServiceName}}
  namespace: {{.TestNamespace}}
  labels:
    app: {{.DeploymentName}}
spec:
  ports:
    - port: 8080
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app: {{.DeploymentName}}
`

	deploymentTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.DeploymentName}}
  namespace: {{.TestNamespace}}
  labels:
    app: {{.DeploymentName}}
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: {{.DeploymentName}}
  template:
    metadata:
      labels:
        app: {{.DeploymentName}}
    spec:
      containers:
        - name: {{.DeploymentName}}
          image: registry.k8s.io/e2e-test-images/agnhost:2.45
          args:
          - netexec
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          readinessProbe:
            httpGet:
              path: /
              port: http
`

	loadJobTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: generate-request
  namespace: {{.TestNamespace}}
spec:
  template:
    spec:
      containers:
      - name: generate-request
        image: alpine
        imagePullPolicy: Always
        command:
		- sh
		- -c
		- 'apk add apache2-utils && ab -H "Host: {{.Host}}" -c 25 -n 1000000 "http://keda-http-add-on-interceptor-proxy.keda:8080/"'
      restartPolicy: Never
  activeDeadlineSeconds: 600
  backoffLimit: 5
`

	httpScaledObjectTemplate = `
kind: HTTPScaledObject
apiVersion: http.keda.sh/v1alpha1
metadata:
  name: {{.HTTPScaledObjectName}}
  namespace: {{.TestNamespace}}
spec:
  hosts:
  - {{.Host1}}
  - {{.Host2}}
  targetPendingRequests: 100
  scaledownPeriod: 10
  scaleTargetRef:
    deployment: {{.DeploymentName}}
    service: {{.ServiceName}}
    port: 8080
  replicas:
    min: {{ .MinReplicas }}
    max: {{ .MaxReplicas }}
`
)

func TestCheck(t *testing.T) {
	// setup
	t.Log("--- setting up ---")
	// Create kubernetes resources
	kc := GetKubernetesClient(t)
	data, templates := getTemplateData()
	CreateNamespace(t, kc, testNamespace)

	testScaleHost(t, kc, data, templates, host1)
	testScaleHost(t, kc, data, templates, host2)

	// cleanup
	DeleteNamespace(t, testNamespace)
}

func testScaleHost(t *testing.T, kc *kubernetes.Clientset, data templateData, templates []Template, host string) {
	t.Log("--- testing scale out ---")

	// create resources
	KubectlApplyMultipleWithTemplate(t, data, templates)
	assert.True(t, WaitForDeploymentReplicaReadyCount(t, kc, deploymentName, testNamespace, minReplicaCount, 6, 10),
		"replica count should be %d after 1 minutes", minReplicaCount)

	data.Host = host
	// scale out
	KubectlApplyWithTemplate(t, data, "loadJobTemplate", loadJobTemplate)
	assert.True(t, WaitForDeploymentReplicaReadyCount(t, kc, deploymentName, testNamespace, maxReplicaCount, 6, 10),
		"replica count should be %d after 1 minutes", maxReplicaCount)

	// scale in
	KubectlDeleteWithTemplate(t, data, "loadJobTemplate", loadJobTemplate)
	assert.True(t, WaitForDeploymentReplicaReadyCount(t, kc, deploymentName, testNamespace, minReplicaCount, 12, 10),
		"replica count should be %d after 2 minutes", minReplicaCount)

	// cleanup
	KubectlDeleteMultipleWithTemplate(t, data, templates)
}

func getTemplateData() (templateData, []Template) {
	return templateData{
			TestNamespace:        testNamespace,
			DeploymentName:       deploymentName,
			ServiceName:          serviceName,
			HTTPScaledObjectName: httpScaledObjectName,
			Host1:                host1,
			Host2:                host2,
			Replicas:             initReplicaCount,
			MinReplicas:          minReplicaCount,
			MaxReplicas:          maxReplicaCount,
		}, []Template{
			{Name: "deploymentTemplate", Config: deploymentTemplate},
			{Name: "serviceNameTemplate", Config: serviceTemplate},
			{Name: "httpScaledObjectTemplate", Config: httpScaledObjectTemplate},
		}
}
