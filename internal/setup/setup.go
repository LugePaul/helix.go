package setup

import (
	"github.com/mountayaapp/helix.go/internal/cloudprovider"
	"github.com/mountayaapp/helix.go/internal/cloudprovider/kubernetes"
	"github.com/mountayaapp/helix.go/internal/cloudprovider/nomad"
	"github.com/mountayaapp/helix.go/internal/cloudprovider/render"
	"github.com/mountayaapp/helix.go/internal/cloudprovider/unknown"
)

/*
init ensures helix.go global environment is properly setup: cloud provider is
mandatory for logger and tracer, which are required for a service to work as
expected.
*/
func init() {
	if cloudprovider.Detected == nil {
		cloudproviders := []cloudprovider.CloudProvider{
			kubernetes.Get(),
			nomad.Get(),
			render.Get(),
			unknown.Get(),
		}

		for _, orch := range cloudproviders {
			if orch != nil {
				cloudprovider.Detected = orch
				break
			}
		}
	}
}
