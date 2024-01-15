package kilt

import (
	"fmt"

	"github.com/go-akka/configuration"
)

func extractRuntime(config *configuration.Config) (*Runtime, error) {
	r := new(Runtime)

	if config.IsArray("runtime.upload") {
		uploads := config.GetValue("runtime.upload").GetArray()

		for k, u := range uploads {
			if u.IsObject() {
				var err error
				upload := u.GetObject()

				newUpload := new(RuntimeUpload)

				newUpload.Payload, err = retrievePayload(upload)
				if err != nil {
					return nil, fmt.Errorf("could not extract payload for entry %d: %w", k, err)
				}

				newUpload.Destination = upload.GetKey("as").GetString()

				if newUpload.Destination == "" {
					return nil, fmt.Errorf("could not extract destination for entry %d: 'as' cannot be empty", k)
				}

				newUpload.Uid = getWithDefaultUint16(upload, "uid", DefaultUserID)
				newUpload.Gid = getWithDefaultUint16(upload, "gid", DefaultGroupID)
				newUpload.Permissions = getWithDefaultUint32(upload, "permissions", DefaultPermissions)

				r.Uploads = append(r.Uploads, *newUpload)
			}
		}
	}

	if config.IsArray("runtime.exec") {
		for k, e := range config.GetValue("runtime.exec").GetArray() {
			if e.IsObject() {
				exec := e.GetObject()

				execParams := exec.GetKey("run").GetStringList()

				if len(execParams) == 0 {
					return nil, fmt.Errorf("could not add exec at entry %d: run cannot have 0 arguments", k)
				}

				r.Executables = append(r.Executables, RuntimeExecutable{Run: execParams})
			}
		}
	}

	return r, nil
}
