package inbound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/dop251/goja"
)

const (
	EventFilterInboundName  = "event-filter"
	DirektivEventDropHeader = "Direktiv-Event-Drop"
)

type EventFilterInboundPlugin struct {
	Script         string `mapstructure:"script"`
	AllowNonEvents bool   `mapstructure:"allow_non_events"`
}

func (js *EventFilterInboundPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	x := &EventFilterInboundPlugin{}

	err := gateway.ConvertConfig(config.Config, x)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (js *EventFilterInboundPlugin) Type() string {
	return EventFilterInboundName
}

func (js *EventFilterInboundPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can not set read body for js inbound plugin")

			return nil, nil
		}
		defer r.Body.Close()
	}

	// try to make it an object, if it is not possible proceed with next plugin
	var event map[string]interface{}
	err = json.Unmarshal(b, &event)
	if err != nil && js.AllowNonEvents {
		// proceed with non-event
		r.Body = io.NopCloser(bytes.NewReader(b))
		return w, r
	} else if err != nil {
		w.Header().Set(DirektivEventDropHeader, "true")
		return nil, nil
	}

	vm := goja.New()
	err = vm.Set("event", event)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set event object")

		return nil, nil
	}

	// set log
	err = vm.Set("log", func(txt interface{}) {
		slog.Info("event filter log", slog.Any("log", txt))
	})
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set log function")

		return nil, nil
	}

	script := fmt.Sprintf("function run() { %s; } run()",
		js.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not execute script")

		return nil, nil
	}

	if val != nil && !val.Equals(goja.Undefined()) {
		o := val.ToObject(vm)

		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(event) {
			// nolint checked before
			responseDone := o.Export().(map[string]interface{})
			e, err := json.Marshal(responseDone)
			if err != nil {
				gateway.WriteInternalError(r, w, err, "can not execute script")

				return nil, nil
			}

			r.Body = io.NopCloser(bytes.NewReader(e))

			return w, r
		}
	}

	w.Header().Set(DirektivEventDropHeader, "true")

	// got dropped and we return false
	return nil, nil
}

//nolint:gochecknoinits
func init() {
	gateway.RegisterPlugin(&EventFilterInboundPlugin{})
}
