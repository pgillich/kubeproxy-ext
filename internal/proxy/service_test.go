package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pgillich/kubeproxy-ext/configs"
	"github.com/pgillich/kubeproxy-ext/internal/logger"
)

type ServiceTestSuite struct {
	suite.Suite

	service *Service
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

type TestServer struct {
	server *httptest.Server
}

func (ts *TestServer) ListenAndServe() error {
	return nil
}

func (ts *TestServer) Shutdown(ctx context.Context) error {
	ts.server.Close()

	return http.ErrServerClosed
}

type TestTransport struct {
	fileTransport http.RoundTripper
}

func (t *TestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.fileTransport.RoundTrip(r) // nolint:wrapcheck // OK
}

func (s *ServiceTestSuite) SetupTest() {
	var err error
	testServer := &TestServer{
		httptest.NewServer(nil),
	}
	s.service, err = New(configs.Proxy{
		TargetURL:  "http://127.0.0.1:8001",
		ListenAddr: testServer.server.Listener.Addr().String(),

		HTTPServer:     testServer,
		ProxyTransport: &TestTransport{http.NewFileTransport(http.Dir("../../test/"))},
	}, logger.New().Logger)
	require.NoError(s.T(), err, "SetupTest")

	testServer.server.Config.Handler = s.service.proxy
}

func (s *ServiceTestSuite) TearDownTest() {
	s.service.server.Shutdown(context.Background())
}

type HTTPClient struct {
	http.Client
}

func (c *HTTPClient) Get(ctx context.Context, reqURL fmt.Stringer) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (s *ServiceTestSuite) TestService_Proxy_Pod() {
	tests := s.getPodTests()
	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			ctx := context.Background()
			client := HTTPClient{}
			resp, err := client.Get(ctx, &(url.URL{Scheme: "http", Host: s.service.cfg.ListenAddr, Path: tc.bodyFile}))
			require.NoError(s.T(), err, "Get")
			require.NotNil(s.T(), resp, "Get")

			s.checkPod(resp, tc)
		})
	}
}

func (s *ServiceTestSuite) TestService_ModifyResponse_Pod() {
	tests := s.getPodTests()
	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			body, err := os.ReadFile("../../test" + tc.bodyFile)
			require.NoError(s.T(), err, "bodyFile")
			resp := &http.Response{
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				Header:        make(http.Header),
				Close:         true,
				Body:          io.NopCloser(bytes.NewReader(body)),
				ContentLength: int64(len(body)),
			}
			err = s.service.ModifyResponse(resp)
			if tc.wantErr == nil {
				require.NoError(s.T(), err, "ModifyResponse")
			} else {
				require.ErrorAs(s.T(), tc.wantErr, err, "ModifyResponse")
			}

			s.checkPod(resp, tc)
		})
	}
}

type podTest struct {
	name        string
	bodyFile    string
	wantErr     error
	wantKubectl map[string]interface{}
}

func (s *ServiceTestSuite) checkPod(resp *http.Response, tc podTest) {
	s.T().Helper()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err, "ReadAll resp.Body")
	err = resp.Body.Close()
	require.NoError(s.T(), err, "ReadAll resp.Body")

	unstrObj := &unstructured.Unstructured{}
	err = unstrObj.UnmarshalJSON(respBody)
	require.NoError(s.T(), err, "UnmarshalJSON resp.Body")
	require.Equal(s.T(), "Pod", unstrObj.GetKind())

	kubectlValues, has := unstrObj.UnstructuredContent()[configs.ObjectKeyKubectl]
	s.True(has, "ObjectKeyKubectl")
	kubectlMap, is := kubectlValues.(map[string]interface{})
	s.True(is, "ObjectKeyKubectl")
	if tc.wantKubectl["AGE"] != "" {
		s.NotEmpty(kubectlMap["AGE"], "AGE")
	}
	tc.wantKubectl["AGE"] = kubectlMap["AGE"]
	s.EqualValues(tc.wantKubectl, kubectlMap, "kubectlColumns")
}

func (*ServiceTestSuite) getPodTests() []podTest {
	tests := []podTest{
		{
			name:     "Completed",
			bodyFile: "/pod-status/Completed.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "137d",
				"IP":              "10.92.92.119",
				"NAME":            "secret-generator--1-jvbz8",
				"NODE":            "hu2-vmp9",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Completed",
				"CONDITIONS":      "Succeeded, The pod has completed successfully.",
			},
		},
		{
			name:     "ContainerCreating",
			bodyFile: "/pod-status/ContainerCreating.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "121d",
				"IP":              "<none>",
				"NAME":            "mysql-564d57cc47-qmlwp",
				"NODE":            "hu2-vmp9",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "ContainerCreating",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "CrashLoopBackOff",
			bodyFile: "/pod-status/CrashLoopBackOff.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "130d",
				"IP":              "10.92.92.14",
				"NAME":            "longhorn-driver-deployer-69985cff47-zrr68",
				"NODE":            "hu2-vmp9",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(3778),
				"STATUS":          "CrashLoopBackOff",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "ErrImagePull",
			bodyFile: "/pod-status/ErrImagePull.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "113d",
				"IP":              "10.92.118.250",
				"NAME":            "tester",
				"NODE":            "o-ci-01",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "ErrImagePull",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Error",
			bodyFile: "/pod-status/Error.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "248d",
				"IP":              "10.90.2.3",
				"NAME":            "nodelocaldns-krt85",
				"NODE":            "hu2-vmp3",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(3),
				"STATUS":          "Error",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "ImageInspectError",
			bodyFile: "/pod-status/ImageInspectError.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "114d",
				"IP":              "10.92.81.12",
				"NAME":            "ksniff-8ljgd",
				"NODE":            "o-k8s-vps1",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "ImageInspectError",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "ImagePullBackOff",
			bodyFile: "/pod-status/ImagePullBackOff.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "113d",
				"IP":              "10.92.118.250",
				"NAME":            "tester",
				"NODE":            "o-ci-01",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "ImagePullBackOff",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Init_CrashLoopBackOff",
			bodyFile: "/pod-status/Init_CrashLoopBackOff.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "130d",
				"IP":              "10.92.92.184",
				"NAME":            "kratos-6c76994b97-gstzr",
				"NODE":            "hu2-vmp9",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(5345),
				"STATUS":          "Init:CrashLoopBackOff",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Init_ImagePullBackOff",
			bodyFile: "/pod-status/Init_ImagePullBackOff.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "114d",
				"IP":              "10.244.2.6",
				"NAME":            "my-nginx-68cd7d56f4-8h2m5",
				"NODE":            "demo-worker",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Init:ErrImagePull",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Init_Running",
			bodyFile: "/pod-status/Init_Running.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "114d",
				"IP":              "10.244.2.7",
				"NAME":            "my-nginx-7764d469c9-7wsmh",
				"NODE":            "demo-worker",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Init:0/1",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Init_Terminating",
			bodyFile: "/pod-status/Init_Terminating.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "114d",
				"IP":              "10.244.2.7",
				"NAME":            "my-nginx-7764d469c9-7wsmh",
				"NODE":            "demo-worker",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Terminating",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "PodInitializing",
			bodyFile: "/pod-status/PodInitializing.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "114d",
				"IP":              "<none>",
				"NAME":            "my-nginx-5997694d7b-vfvff",
				"NODE":            "demo-worker",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "0/1",
				"READY":           "0/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Init:0/1",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Running2",
			bodyFile: "/pod-status/Running2.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "239d",
				"IP":              "10.90.2.9",
				"NAME":            "node-exporter-s6tbv",
				"NODE":            "hu2-vmp9",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "2/2",
				"RESTARTS":        int64(6),
				"STATUS":          "Running",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Running3",
			bodyFile: "/pod-status/Running3.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "130d",
				"IP":              "10.92.86.166",
				"NAME":            "blackbox-exporter-6798fb5bb4-rb6qc",
				"NODE":            "hu2-vmp6",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "3/3",
				"RESTARTS":        int64(0),
				"STATUS":          "Running",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Running",
			bodyFile: "/pod-status/Running.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "215d",
				"IP":              "10.92.86.76",
				"NAME":            "coredns-8474476ff8-lfwcf",
				"NODE":            "hu2-vmp6",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "1/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Running",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Terminating2",
			bodyFile: "/pod-status/Terminating2.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "215d",
				"IP":              "10.92.124.5",
				"NAME":            "percona-server-mongodb-operator-7d76d4844d-2wjwv",
				"NODE":            "hu2-vmp3",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "2/2",
				"RESTARTS":        int64(1),
				"STATUS":          "Terminating",
				"CONDITIONS":      "<none>",
			},
		},
		{
			name:     "Terminating",
			bodyFile: "/pod-status/Terminating.json",
			wantErr:  nil,
			wantKubectl: map[string]interface{}{
				"AGE":             "145d",
				"IP":              "10.92.124.218",
				"NAME":            "coredns-8474476ff8-m8xzl",
				"NODE":            "hu2-vmp3",
				"NOMINATED_NODE":  "<none>",
				"READINESS_GATES": "<none>",
				"READY":           "1/1",
				"RESTARTS":        int64(0),
				"STATUS":          "Terminating",
				"CONDITIONS":      "<none>",
			},
		},
	}

	return tests
}

type podListTest struct {
	name        string
	bodyFile    string
	wantErr     error
	wantKubectl map[string]string
}

func (s *ServiceTestSuite) TestService_Proxy_PodList() {
	tests := s.getPodListTests()
	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			ctx := context.Background()
			client := HTTPClient{}
			resp, err := client.Get(ctx, &(url.URL{Scheme: "http", Host: s.service.cfg.ListenAddr, Path: tc.bodyFile}))
			s.NoError(err, "Get")
			s.NotNil(resp, "Get")

			s.checkPodList(resp, tc)
		})
	}
}

func (s *ServiceTestSuite) checkPodList(resp *http.Response, tc podListTest) {
	s.T().Helper()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err, "ReadAll resp.Body")
	err = resp.Body.Close()
	require.NoError(s.T(), err, "ReadAll resp.Body")

	unstrList := &unstructured.UnstructuredList{}
	err = unstrList.UnmarshalJSON(respBody)
	require.NoError(s.T(), err, "UnmarshalJSON resp.Body")
	require.Contains(s.T(), []string{"PodList", "List"}, unstrList.GetKind())

	result := map[string]string{}
	for _, item := range unstrList.Items {
		require.Equal(s.T(), "Pod", item.GetKind())

		kubectlValues, has := item.UnstructuredContent()[configs.ObjectKeyKubectl]
		s.True(has, "ObjectKeyKubectl")
		kubectlMap, is := kubectlValues.(map[string]interface{})
		s.True(is, "ObjectKeyKubectl")
		kubectlKeys := make([]string, 0, len(kubectlMap))
		for k := range kubectlMap {
			if k != "AGE" {
				kubectlKeys = append(kubectlKeys, k)
			}
		}
		sort.Strings(kubectlKeys)
		values := strings.Builder{}
		for _, k := range kubectlKeys {
			if values.Len() > 0 {
				values.WriteString("; ")
			}
			values.WriteString(fmt.Sprintf("%v", kubectlMap[k]))
		}
		result[item.GetName()] = values.String()
	}

	s.EqualValues(tc.wantKubectl, result, "kubectlRows")
}

func (*ServiceTestSuite) getPodListTests() []podListTest {
	tests := []podListTest{

		{
			name:        "No Pod",
			bodyFile:    "/podlist-status/no-pod.json",
			wantErr:     nil,
			wantKubectl: map[string]string{},
		},

		{
			name:     "Longhorn",
			bodyFile: "/podlist-status/longhorn-system.json",
			wantErr:  nil,
			wantKubectl: map[string]string{
				"csi-attacher-5f46994f7-28t4p":              "<none>; 10.72.81.95; csi-attacher-5f46994f7-28t4p; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"csi-attacher-5f46994f7-5pc5s":              "<none>; 10.72.106.230; csi-attacher-5f46994f7-5pc5s; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"csi-attacher-5f46994f7-cnnr9":              "<none>; 10.72.78.126; csi-attacher-5f46994f7-cnnr9; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"csi-provisioner-6ccbfbf86f-qhjwm":          "<none>; 10.72.81.92; csi-provisioner-6ccbfbf86f-qhjwm; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"csi-provisioner-6ccbfbf86f-rn889":          "<none>; 10.72.78.120; csi-provisioner-6ccbfbf86f-rn889; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"csi-provisioner-6ccbfbf86f-zjx9h":          "<none>; 10.72.106.228; csi-provisioner-6ccbfbf86f-zjx9h; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"csi-resizer-6dd8bd4c97-2k62f":              "<none>; 10.72.106.236; csi-resizer-6dd8bd4c97-2k62f; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"csi-resizer-6dd8bd4c97-l2rwd":              "<none>; 10.72.81.99; csi-resizer-6dd8bd4c97-l2rwd; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"csi-resizer-6dd8bd4c97-qdx52":              "<none>; 10.72.78.115; csi-resizer-6dd8bd4c97-qdx52; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"csi-snapshotter-86f65d8bc-56nd7":           "<none>; 10.72.81.90; csi-snapshotter-86f65d8bc-56nd7; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"csi-snapshotter-86f65d8bc-qn6p8":           "<none>; 10.72.106.227; csi-snapshotter-86f65d8bc-qn6p8; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"csi-snapshotter-86f65d8bc-wrn5k":           "<none>; 10.72.78.127; csi-snapshotter-86f65d8bc-wrn5k; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"engine-image-ei-fa2dfbf0-2cql6":            "<none>; 10.72.106.218; engine-image-ei-fa2dfbf0-2cql6; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"engine-image-ei-fa2dfbf0-8z9kj":            "<none>; 10.72.78.129; engine-image-ei-fa2dfbf0-8z9kj; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"engine-image-ei-fa2dfbf0-lvkvx":            "<none>; 10.72.81.96; engine-image-ei-fa2dfbf0-lvkvx; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"instance-manager-e-7c715a5a":               "<none>; 10.72.78.121; instance-manager-e-7c715a5a; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"instance-manager-e-e691695a":               "<none>; 10.72.106.225; instance-manager-e-e691695a; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"instance-manager-e-ee426bee":               "<none>; 10.72.81.97; instance-manager-e-ee426bee; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"instance-manager-r-05f75ac7":               "<none>; 10.72.81.100; instance-manager-r-05f75ac7; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"instance-manager-r-74c1e529":               "<none>; 10.72.106.229; instance-manager-r-74c1e529; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"instance-manager-r-e6aec4c8":               "<none>; 10.72.78.125; instance-manager-r-e6aec4c8; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"longhorn-csi-plugin-hzlbr":                 "<none>; 10.72.106.220; longhorn-csi-plugin-hzlbr; o-k8s-vps3; <none>; <none>; 2/2; 4; Running",
				"longhorn-csi-plugin-npxkc":                 "<none>; 10.72.78.111; longhorn-csi-plugin-npxkc; o-k8s-vps2; <none>; <none>; 2/2; 4; Running",
				"longhorn-csi-plugin-nw2lv":                 "<none>; 10.72.81.79; longhorn-csi-plugin-nw2lv; o-k8s-vps1; <none>; <none>; 2/2; 2; Running",
				"longhorn-driver-deployer-784546d78d-nz9qv": "<none>; 10.72.78.124; longhorn-driver-deployer-784546d78d-nz9qv; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"longhorn-manager-5cl94":                    "<none>; 10.72.81.89; longhorn-manager-5cl94; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"longhorn-manager-9jgkt":                    "<none>; 10.72.78.107; longhorn-manager-9jgkt; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"longhorn-manager-z4cr5":                    "<none>; 10.72.106.221; longhorn-manager-z4cr5; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"longhorn-ui-9fdb94f9-w84gf":                "<none>; 10.72.78.131; longhorn-ui-9fdb94f9-w84gf; o-k8s-vps2; <none>; <none>; 1/1; 1; Running",
			},
		},

		{
			name:     "Mongo",
			bodyFile: "/podlist-status/mongo.json",
			wantErr:  nil,
			wantKubectl: map[string]string{
				"mongodb-exporter-prometheus-mongodb-exporter-6dcdd8c8fc-zqffr": "<none>; 10.72.106.217; mongodb-exporter-prometheus-mongodb-exporter-6dcdd8c8fc-zqffr; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"mongodb-exporter-prometheus-mongodb-exporter-test-connection":  "Failed, The pod failed.; 10.72.106.17; mongodb-exporter-prometheus-mongodb-exporter-test-connection; o-k8s-vps3; <none>; <none>; 0/1; 0; Error",
				"percona-server-mongodb-operator-fcc5c8d6-sqb8m":                "<none>; 10.72.78.102; percona-server-mongodb-operator-fcc5c8d6-sqb8m; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"vcc-rs0-0": "<none>; 10.72.106.241; vcc-rs0-0; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"vcc-rs0-1": "<none>; 10.72.81.105; vcc-rs0-1; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
			},
		},

		{
			name:     "RabbitMQ",
			bodyFile: "/podlist-status/rabbitmq-system.json",
			wantErr:  nil,
			wantKubectl: map[string]string{
				"messaging-topology-operator-f9c69d45b-gxmj6": "<none>; <none>; messaging-topology-operator-f9c69d45b-gxmj6; o-k8s-vps2; <none>; <none>; 0/1; 0; ContainerCreating",
				"rabbitmq-cluster-operator-7cbf865f89-qwwwj":  "<none>; 10.72.78.116; rabbitmq-cluster-operator-7cbf865f89-qwwwj; o-k8s-vps2; <none>; <none>; 1/1; 1; Running",
				"rabbitmq-server-0":                           "<none>; 10.72.81.104; rabbitmq-server-0; o-k8s-vps1; <none>; <none>; 1/1; 0; Running",
				"rabbitmq-server-1":                           "<none>; 10.72.78.136; rabbitmq-server-1; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
			},
		},

		{
			name:     "Redis",
			bodyFile: "/podlist-status/redis.json",
			wantErr:  nil,
			wantKubectl: map[string]string{
				"redisoperator-56d6888cc-ks84t": "<none>; 10.72.78.128; redisoperator-56d6888cc-ks84t; o-k8s-vps2; <none>; <none>; 0/1; 1964; CrashLoopBackOff",
				"rfr-vcc-0":                     "<none>; 10.72.78.141; rfr-vcc-0; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"rfr-vcc-1":                     "<none>; 10.72.106.239; rfr-vcc-1; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
				"rfs-vcc-5cc6bf796c-g9mnr":      "<none>; 10.72.78.123; rfs-vcc-5cc6bf796c-g9mnr; o-k8s-vps2; <none>; <none>; 1/1; 0; Running",
				"rfs-vcc-5cc6bf796c-mmrkr":      "<none>; 10.72.106.224; rfs-vcc-5cc6bf796c-mmrkr; o-k8s-vps3; <none>; <none>; 1/1; 0; Running",
			},
		},
	}

	return tests
}

type svcListTest struct {
	name     string
	bodyFile string
	wantErr  error
}

func (s *ServiceTestSuite) TestService_Proxy_SvcList() {
	tests := s.getSvcListTests()
	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			ctx := context.Background()
			client := HTTPClient{}
			resp, err := client.Get(ctx, &(url.URL{Scheme: "http", Host: s.service.cfg.ListenAddr, Path: tc.bodyFile}))
			s.NoError(err, "Get")
			s.NotNil(resp, "Get")

			s.checkSvcList(resp, tc)
		})
	}
}

func (s *ServiceTestSuite) checkSvcList(resp *http.Response, _ svcListTest) {
	s.T().Helper()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err, "ReadAll resp.Body")
	err = resp.Body.Close()
	require.NoError(s.T(), err, "ReadAll resp.Body")

	unstrList := &unstructured.UnstructuredList{}
	err = unstrList.UnmarshalJSON(respBody)
	require.NoError(s.T(), err, "UnmarshalJSON resp.Body")
	require.Contains(s.T(), []string{"ServiceList", "List"}, unstrList.GetKind())

	for _, item := range unstrList.Items {
		require.Equal(s.T(), "Service", item.GetKind())
	}
}

func (*ServiceTestSuite) getSvcListTests() []svcListTest {
	tests := []svcListTest{
		{
			name:     "Service",
			bodyFile: "/svclist-status/monitoring.json",
			wantErr:  nil,
		},
	}

	return tests
}
