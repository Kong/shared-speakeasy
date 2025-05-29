package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
)

// BeforeRequest defaults mesh related fields:
// for mesh control plane - the "features" field to include "MeshCreation" and "HostnameGeneratorCreation" with "enabled" set to false
// for mesh - the "skipCreatingInitialPolicies" field to include "*".
func BeforeRequest(apiPrefix string, features bool, initialPolicies bool) func(req *http.Request) (*http.Request, error) {
	return func(req *http.Request) (*http.Request, error) {
		if features && req.Method == http.MethodPost && req.URL.Path == apiPrefix {
			var bodyMap map[string]interface{}

			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				defer req.Body.Close()

				if len(bodyBytes) > 0 {
					if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
						return nil, err
					}
				}
			}

			if bodyMap == nil {
				bodyMap = make(map[string]interface{})
			}

			// Define the default "features" value
			defaultFeatures := []map[string]interface{}{
				{
					"type":         "MeshCreation",
					"meshCreation": map[string]bool{"enabled": false},
				},
				{
					"type":                      "HostnameGeneratorCreation",
					"hostnameGeneratorCreation": map[string]bool{"enabled": false},
				},
			}

			if _, exists := bodyMap["features"]; !exists {
				bodyMap["features"] = defaultFeatures
			}

			newBody, err := json.Marshal(bodyMap)
			if err != nil {
				return nil, err
			}

			req.Body = io.NopCloser(bytes.NewReader(newBody))
			req.ContentLength = int64(len(newBody))
			req.Header.Set("Content-Type", "application/json")
		}

		if initialPolicies {
			// Check for requests matching POST: apiPrefix/*/api/meshes/*
			match, err := regexp.MatchString(`^`+apiPrefix+`/[^/]+/api/meshes/[^/]+$`, req.URL.Path)
			if err != nil {
				return nil, err
			}

			if req.Method == http.MethodPut && match {
				var bodyMap map[string]interface{}

				if req.Body != nil {
					bodyBytes, err := io.ReadAll(req.Body)
					if err != nil {
						return nil, err
					}
					defer req.Body.Close()

					if len(bodyBytes) > 0 {
						if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
							return nil, err
						}
					}
				}

				if bodyMap == nil {
					bodyMap = make(map[string]interface{})
				}

				// Define the default "skipCreatingInitialPolicies" value
				if _, exists := bodyMap["skipCreatingInitialPolicies"]; !exists {
					bodyMap["skipCreatingInitialPolicies"] = []string{"*"}
				}

				newBody, err := json.Marshal(bodyMap)
				if err != nil {
					return nil, err
				}

				req.Body = io.NopCloser(bytes.NewReader(newBody))
				req.ContentLength = int64(len(newBody))
				req.Header.Set("Content-Type", "application/json")
			}

		}
		return req, nil
	}
}
