package compose

import _ "embed"

//go:embed templates/docker-compose.yaml.tmpl
var composeTemplate string
