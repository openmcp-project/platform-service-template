 #!/usr/bin/env bash
  # Usage: ./init.sh github.com/myorg/my-platform-service MyService
  set -euo pipefail

  MODULE=${1:?usage: init.sh <module> <Kind>}
  KIND=${2:?usage: init.sh <module> <Kind>}
  KIND_LOWER=$(echo "$KIND" | tr '[:upper:]' '[:lower:]')

  find . -type f \( -name "*.go" -o -name "*.yaml" -o -name "go.mod" \) \
    | grep -v ".git" \
    | xargs sed -i.bak \
        -e "s/FooService/${KIND}/g" \
        -e "s/fooservice/${KIND_LOWER}/g" \
        -e "s|github.com/openmcp-project/platform-service-template|${MODULE}|g"

  find . -name "*.bak" -delete

  git mv internal/controller/fooservice_controller.go \
         internal/controller/${KIND_LOWER}_controller.go