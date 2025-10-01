package valkey

import (
	"fmt"

	"github.com/mountayaapp/helix.go/telemetry/trace"
)

/*
setKeyAttributes sets key attributes to a trace span.
*/
func setKeyAttributes(span *trace.Span, key string) {
	span.SetStringAttribute(fmt.Sprintf("%s.key", identifier), key)
}
