//go:build integ
// +build integ

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outboundtrafficpolicy

import (
	"net/http"
	"testing"
)

func TestOutboundTrafficPolicy_AllowAny(t *testing.T) {
	cases := []*TestCase{
		{
			Name:     "HTTP Traffic",
			PortName: "http",
			Expected: Expected{
				Metric:          "istio_requests_total",
				PromQueryFormat: `sum(istio_requests_total{reporter="source",destination_service_name="PassthroughCluster",response_code="200"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/1.1",
			},
		},
		{
			Name:     "HTTP H2 Traffic",
			PortName: "http",
			HTTP2:    true,
			Expected: Expected{
				Metric:          "istio_requests_total",
				PromQueryFormat: `sum(istio_requests_total{reporter="source",destination_service_name="PassthroughCluster",response_code="200"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/2.0",
			},
		},
		{
			Name:     "HTTPS Traffic",
			PortName: "https",
			Expected: Expected{
				Metric:          "istio_tcp_connections_opened_total",
				PromQueryFormat: `sum(istio_tcp_connections_opened_total{reporter="source",destination_service_name="PassthroughCluster"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/1.1",
			},
		},
		{
			Name:     "HTTPS Traffic Conflict",
			PortName: "https-conflict",
			Expected: Expected{
				Metric:          "istio_tcp_connections_opened_total",
				PromQueryFormat: `sum(istio_tcp_connections_opened_total{reporter="source",destination_service_name="PassthroughCluster"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/1.1",
			},
		},
		{
			Name:     "HTTPS H2 Traffic",
			PortName: "https",
			HTTP2:    true,
			Expected: Expected{
				Metric:          "istio_tcp_connections_opened_total",
				PromQueryFormat: `sum(istio_tcp_connections_opened_total{reporter="source",destination_service_name="PassthroughCluster"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/2.0",
			},
		},
		{
			Name:     "HTTPS H2 Traffic Conflict",
			PortName: "https-conflict",
			HTTP2:    true,
			Expected: Expected{
				Metric:          "istio_tcp_connections_opened_total",
				PromQueryFormat: `sum(istio_tcp_connections_opened_total{reporter="source",destination_service_name="PassthroughCluster"})`,
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/2.0",
			},
		},
		{
			Name:     "HTTP Traffic Egress",
			PortName: "http",
			Host:     "some-external-site.com",
			Expected: Expected{
				Metric:          "istio_requests_total",
				PromQueryFormat: `sum(istio_requests_total{reporter="source",destination_service_name="istio-egressgateway",response_code="200"})`, // nolint: lll
				StatusCode:      http.StatusOK,
				Protocol:        "HTTP/1.1",
				RequestHeaders: map[string]string{
					// We inject this header in the VirtualService
					"Handled-By-Egress-Gateway": "true",
				},
			},
		},
		{
			Name:     "HTTP H2 Traffic Egress",
			PortName: "http",
			HTTP2:    true,
			Host:     "some-external-site.com",
			Expected: Expected{
				Metric:          "istio_requests_total",
				PromQueryFormat: `sum(istio_requests_total{reporter="source",destination_service_name="istio-egressgateway",response_code="200"})`, // nolint: lll
				StatusCode:      http.StatusOK,
				// Even though we send h2 to the gateway, the gateway should send h1, as configured by the ServiceEntry
				Protocol: "HTTP/1.1",
				RequestHeaders: map[string]string{
					// We inject this header in the VirtualService
					"Handled-By-Egress-Gateway": "true",
				},
			},
		},
		// TODO add HTTPS through gateway
		{
			Name:     "TCP",
			PortName: "tcp",
			Expected: Expected{
				// TODO(https://github.com/istio/istio/issues/22717) re-enable TCP
				// Metric:          "istio_tcp_connections_closed_total",
				// PromQueryFormat: `sum(istio_tcp_connections_closed_total{reporter="source",destination_service_name="PassthroughCluster",source_workload="client-v1"})`,
				StatusCode: http.StatusOK,
				Protocol:   "TCP",
			},
		},
		{
			Name:     "TCP Conflict",
			PortName: "tcp",
			Expected: Expected{
				// TODO(https://github.com/istio/istio/issues/22717) re-enable TCP
				// Metric:          "istio_tcp_connections_closed_total",
				// PromQueryFormat: `sum(istio_tcp_connections_closed_total{reporter="source",destination_service_name="PassthroughCluster",source_workload="client-v1"})`,
				StatusCode: http.StatusOK,
				Protocol:   "TCP",
			},
		},
	}

	RunExternalRequest(cases, prom, AllowAny, t)
}
