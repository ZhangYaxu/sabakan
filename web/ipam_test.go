package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func testConfigIPAMGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	config := &sabakan.IPAMConfig{
		MaxNodesInRack:   28,
		NodeIPv4Pool:     "10.69.0.0/20",
		NodeIPv4Offset:   "",
		NodeRangeSize:    6,
		NodeRangeMask:    26,
		NodeIndexOffset:  3,
		NodeIPPerNode:    3,
		BMCIPv4Pool:      "10.72.16.0/20",
		BMCIPv4Offset:    "0.0.1.0",
		BMCRangeSize:     5,
		BMCRangeMask:     20,
		BMCGatewayOffset: 1,
	}

	err := m.IPAM.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/config/ipam", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	var result sabakan.IPAMConfig
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&result, config) {
		t.Errorf("wrong config: %#v", result)
	}
}

func testConfigIPAMPut(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	bad := "{}"
	good := `
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/20",
   "node-ipv4-offset": "0.0.0.0",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-index-offset": 3,
   "node-ip-per-node": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-offset": "0.0.1.0",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20,
   "bmc-ipv4-gateway-offset": 1
}
`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(bad))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest")
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	conf, err := m.IPAM.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	expected := &sabakan.IPAMConfig{
		MaxNodesInRack:   28,
		NodeIPv4Pool:     "10.69.0.0/20",
		NodeIPv4Offset:   "0.0.0.0",
		NodeRangeSize:    6,
		NodeRangeMask:    26,
		NodeIndexOffset:  3,
		NodeIPPerNode:    3,
		BMCIPv4Pool:      "10.72.16.0/20",
		BMCIPv4Offset:    "0.0.1.0",
		BMCRangeSize:     5,
		BMCRangeMask:     20,
		BMCGatewayOffset: 1,
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("mismatch: %#v", conf)
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	machine := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "1234",
	})
	err = m.Machine.Register(context.Background(), []*sabakan.Machine{machine})
	if err != nil {
		t.Fatal(err)
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode == http.StatusOK {
		t.Error("request must not succeed")
	}
}

func TestConfigIPAM(t *testing.T) {
	t.Run("Get", testConfigIPAMGet)
	t.Run("Put", testConfigIPAMPut)
}
