package inbound_test

import (
	"github.com/direktiv/direktiv/direktiv-ee/pkg/plugins/inbound"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/stretchr/testify/assert"
)

var (
	cloudEventEmpty        = ``
	cloudEventNonCompliant = `{ "data": "value" }`
	cloudEventGarbage      = `fsdsfsddsdsfdsdsf`
	cloudEventOk           = `
	{
		"specversion" : "1.0",
		"type" : "com.github.pull.create1",
		"source" : "https://github.com/cloudevents/spec/pull",
		"subject" : "123",
		"time" : "2018-04-05T17:31:00Z",
		"comexampleextension1" : "value",
		"comexampleothervalue" : 5,
		"datacontenttype" : "text/xml",
		"data" : "<much wow=\"xml\"/>"
	}`
)

func TestEventFilterBrokenScript(t *testing.T) {

	p := &inbound.EventFilterInboundPlugin{}

	config := core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		brokenScript
		`,
		},
	}

	p2, _ := p.NewInstance(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventOk))
	r = gateway.InjectContextEndpoint(r, &core.Endpoint{})
	w := httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestEventFilterOkReturn(t *testing.T) {

	p := &inbound.EventFilterInboundPlugin{}

	config := core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		if (event["type"] == "com.github.pull.create1") {
			event["type"] = "changed"
		}
		
		return event
		`,
		},
	}

	p2, _ := p.NewInstance(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventOk))
	w := httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Empty(t, w.Header().Get(inbound.DirektivEventDropHeader))

}

func TestEventFilterOkEmpty(t *testing.T) {

	p := &inbound.EventFilterInboundPlugin{}

	config := core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		if (event["type"] == "com.github.pull.create1") {
			event["type"] = "changed"
		}
		`,
		},
	}

	p2, _ := p.NewInstance(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventOk))
	w := httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Equal(t, w.Header().Get(inbound.DirektivEventDropHeader), "true")

}

func TestEventFilterGarbage(t *testing.T) {

	p := &inbound.EventFilterInboundPlugin{}

	config := core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		return event
		`,
		},
	}

	p2, _ := p.NewInstance(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventGarbage))
	w := httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Equal(t, "true", w.Header().Get(inbound.DirektivEventDropHeader))

	config = core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		return event
		`,
			"allow_non_events": true,
		},
	}

	p2, _ = p.NewInstance(config)
	r, _ = http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventGarbage))
	w = httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Empty(t, w.Header().Get(inbound.DirektivEventDropHeader))

}

func TestEventFilterNonCompliant(t *testing.T) {

	p := &inbound.EventFilterInboundPlugin{}

	config := core.PluginConfig{
		Typ: p.Type(),
		Config: map[string]any{
			"script": `
		if (event["type"] == "com.github.pull.create1") {
			event["type"] = "changed"
		}

		return event
		`,
		},
	}

	p2, _ := p.NewInstance(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", strings.NewReader(cloudEventNonCompliant))
	w := httptest.NewRecorder()
	p2.Execute(w, r)

	assert.Empty(t, w.Header().Get(inbound.DirektivEventDropHeader))
}
