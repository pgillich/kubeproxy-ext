package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/printers"
	"k8s.io/kubernetes/pkg/printers/internalversion"

	"github.com/pgillich/kubeproxy-ext/configs"
)

type Service struct {
	cfg            configs.Proxy
	log            logr.Logger
	proxy          *httputil.ReverseProxy
	server         configs.HTTPServer
	modifiers      map[string]func(item *unstructured.Unstructured) error
	tableGenerator *printers.HumanReadableGenerator
}

// https://stackoverflow.com/questions/52986853/how-to-debug-httputil-newsinglehostreverseproxy

func New(cfg configs.Proxy, log logr.Logger) (*Service, error) {
	var err error
	var targetURL *url.URL
	if targetURL, err = url.Parse(cfg.TargetURL); err != nil {
		return nil, fmt.Errorf("targeturl: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxyTransport := cfg.ProxyTransport
	if proxyTransport == nil {
		proxyTransport = &DebugTransport{log: log}
	}
	proxy.Transport = proxyTransport

	server := cfg.HTTPServer
	if server == nil {
		server = &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: proxy,
		}
	}
	service := &Service{
		cfg:            cfg,
		log:            log,
		proxy:          proxy,
		server:         server,
		tableGenerator: printers.NewTableGenerator(),
	}
	proxy.ModifyResponse = service.ModifyResponse
	service.modifiers = map[string]func(item *unstructured.Unstructured) error{
		"Pod": service.modifyPod,
	}

	internalversion.AddHandlers(service.tableGenerator)

	return service, nil
}

func (s *Service) Serve() {
	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.log.Error(err, "Proxy")
	}
}

var errBodyNotExtended = errors.New("body not extended")

func (s *Service) ModifyResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("resp body close: %w", err)
	}

	newBody := body
	if body, err = s.extendBody(body); err != nil {
		if !errors.Is(err, errBodyNotExtended) {
			s.log.Error(err, "MarshalJSON PodList")
		}
	} else {
		newBody = body
	}

	respBody := io.NopCloser(bytes.NewReader(newBody))
	resp.Body = respBody
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))

	return nil
}

// curl desktop:8003/api/v1/namespaces/kubernetes-dashboard/pods/dashboard-metrics-scraper-c45b7869d-lpdt2
// curl desktop:8003/api/v1/namespaces/kubernetes-dashboard/pods

func (s *Service) extendBody(body []byte) ([]byte, error) {
	unstrList := &unstructured.UnstructuredList{}
	unstrObj := &unstructured.Unstructured{}

	if err := unstrList.UnmarshalJSON(body); err == nil && len(unstrList.Items) > 0 {
		if modifier, has := s.modifiers[unstrList.Items[0].GetKind()]; has {
			for i := range unstrList.Items {
				item := &unstrList.Items[i]
				if err := modifier(item); err != nil {
					return body, fmt.Errorf("modify %s: %w", item.GetKind(), err)
				}
			}
			if bodyOK, err := unstrList.MarshalJSON(); err != nil {
				return body, fmt.Errorf("marshalljson podlist: %w", err)
			} else {
				return bodyOK, nil
			}
		}
	} else if err := unstrObj.UnmarshalJSON(body); err == nil {
		if modifier, has := s.modifiers[unstrObj.GetKind()]; has {
			if err := modifier(unstrObj); err != nil {
				return body, fmt.Errorf("modify %s: %w", unstrObj.GetKind(), err)
			}

			if bodyOK, err := unstrObj.MarshalJSON(); err != nil {
				return body, fmt.Errorf("marshalljson pod: %w", err)
			} else {
				return bodyOK, nil
			}
		}
	}

	return body, errBodyNotExtended
}

func (s *Service) modifyPod(item *unstructured.Unstructured) error {
	pod := &api.Pod{}
	if err := legacyscheme.Scheme.Convert(item, pod, item.GroupVersionKind()); err != nil {
		return fmt.Errorf("fromunstructured pod: %w", err)
	}

	table, err := s.tableGenerator.GenerateTable(pod, printers.GenerateOptions{NoHeaders: false, Wide: true})
	if err != nil {
		return fmt.Errorf("generatetable: %w", err)
	}
	if len(table.Rows) < 1 {
		return configs.ErrGenerateTableNoRow
	}
	if len(table.Rows) > 1 {
		return configs.ErrGenerateTableMoreRows
	}

	values := map[string]interface{}{}
	conditions := []string{}
	for _, condition := range table.Rows[0].Conditions {
		conditions = append(conditions, fmt.Sprintf("%s, %s", condition.Reason, condition.Message))
	}
	if len(conditions) == 0 {
		values[FormatKubectlColumn("Conditions")] = "<none>"
	} else {
		values[FormatKubectlColumn("Conditions")] = strings.Join(conditions, "; ")
	}
	for c, column := range table.ColumnDefinitions {
		values[FormatKubectlColumn(column.Name)] = table.Rows[0].Cells[c]
	}

	if err := unstructured.SetNestedMap(item.UnstructuredContent(), values, configs.ObjectKeyKubectl); err != nil {
		return fmt.Errorf("setnestedstringmap: %w", err)
	}

	return nil
}

func FormatKubectlColumn(col string) string {
	return strings.ReplaceAll(strings.ToUpper(col), " ", "_")
}

type DebugTransport struct {
	log logr.Logger
}

func (d *DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	d.log.Info(string(b))

	return http.DefaultTransport.RoundTrip(r) // nolint:wrapcheck // OK
}
