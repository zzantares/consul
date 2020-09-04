#!/usr/bin/env bash

# tell shellcheck about help
declare -A help; export help

help[generate-protobuf]="Generate Go code for all protobuf files."
generate-protobuf() {
    find . -name '*.proto' | grep -v vendor | xargs --max-args=1 ./do generate-protobuf-file
    echo Generated all protobuf Go files
}

help[generate-protobuf-file]="Generate Go code for a single protobuf file.

Args:
  filename - the file path to the .proto file

Environment Variables:
  EXCLUDE_GRPC           - if set, do not include the grpc plugin
  DISABLE_IMPORT_REPLACE - if set, do not replace protobuf imports with gogoproto imports

"
generate-protobuf-file() {
    local filename="${1:?proto file name must be specified as positional arg}"
    # Remove ./ prefix because protobuf requires paths to have the same format
    filename=${filename#"./"}

    local out="paths=source_relative"
    if [ -z "${DISABLE_IMPORT_REPLACE-}" ]; then
        out="$out,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types"
        out="$out,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types"
        out="$out,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types"
        out="$out,Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api"
        out="$out,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types"
    fi

    if [ -z "${EXCLUDE_GRPC-}" ]; then
        out="$out,plugins=grpc"
    fi
 
    local pb_go_path=${filename%%.proto}.pb.go
    local pb_bin_path=${filename%%.proto}.pb.binary.go
    _echo_green "Generating $filename into $pb_go_path and $pb_bin_path"
    local source_dir; source_dir="$(git rev-parse --show-toplevel)"
    local gogo; gogo=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)
    protoc \
        -I="$gogo/protobuf" \
        -I="$gogo" \
        -I="$(go env GOPATH)/pkg/mod" \
        -I="$source_dir" \
        --gofast_out="$out:$source_dir" \
        --go-binary_out="$source_dir" \
        "$filename"

    _copy_build_tags "$filename" "$pb_go_path"
    _copy_build_tags "$filename" "$pb_bin_path"
}

_copy_build_tags() {
    local from="$1"
    local to="$2"
    local tags; tags=$(grep '// +build' "$from" || true)
    if [ -n "$tags" ];  then
        (echo "$tags"; echo; cat "$to") > "$to.new"
        mv "$to.new" "$to"
    fi
}

_echo_green() {
  if test ! -t 1; then
    echo "$@"
    return
  fi

  echo -ne "\e[32m"
  echo "$@"
  echo -ne "\e[39m"
}
