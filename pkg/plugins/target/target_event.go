package target

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

const (
	TargetEventPluginName = "target-event"
)

//nolint:revive
type TargetEventPlugin struct {
	Namespaces []string `mapstructure:"namespaces" yaml:"namespaces"`
}

func (te *TargetEventPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	x := &TargetEventPlugin{}

	err := gateway.ConvertConfig(config.Config, x)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (te *TargetEventPlugin) sendToNamespace(ctx context.Context, namespace string, header http.Header, payload []byte) {
	// TODO: does this need to log errors somewhere? I think it's probably okay to fail silently.

	// event url
	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/events/broadcast",
		os.Getenv("DIREKTIV_API_PORT"), namespace)

	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header = header

	// req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("Content-Type", "application/json")

	// add api key if required
	if os.Getenv("DIREKTIV_API_KEY") != "" {
		req.Header.Set("Direktiv-Token", os.Getenv("DIREKTIV_API_KEY"))
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return
	}
}

func (te *TargetEventPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	ctx := r.Context()

	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if len(te.Namespaces) == 0 {
		te.Namespaces = []string{currentNS}
	}
	for _, ns := range te.Namespaces {
		if ns != currentNS && currentNS != core.SystemNamespace {
			gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")

			return nil, nil
		}
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not create request")

		return nil, nil
	}

	for _, ns := range te.Namespaces {
		te.sendToNamespace(ctx, ns, r.Header, payload)
	}

	return w, r
}

func (te *TargetEventPlugin) Type() string {
	return TargetEventPluginName
}

//nolint:gochecknoinits
func init() {
	gateway.RegisterPlugin(&TargetEventPlugin{})
}
