package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/api"
)

// GenSpec writes the application's OpenAPI 3.1 document as JSON to out
// (or stdout when out is empty / "-"). It does not need a database or
// HTTP listener — the spec is constructed purely from handler
// declarations in internal/api/handlers.
//
// The Taskfile's `gen` target invokes this and pipes the output to
// orval, which turns it into a typed TypeScript client under
// web/src/generated/.
func GenSpec(out string) error {
	spec := api.Spec()
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if out == "" || out == "-" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(out, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	return nil
}
