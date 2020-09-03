#!/usr/bin/env bash

# tell shellcheck about help
declare -A help; export help

help[generate-protobuf]="Generate Go code for all protobuf files."
generate-protobuf() {
    find . -name '*.proto' | grep -v vendor | xargs --max-args=1 ./do generate-protobuf-file
    echo Generated all protobuf Go files
}

help[generate-protobuf]="Generate Go code for a single protobuf file.

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
    local source_dir; source_dir="$(git rev-parse --show-toplevel)"
 
    local gogo_proto_path
    gogo_proto_path=$(go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)
    local gogo_proto_mod_path
    # shellcheck disable=SC2001
    gogo_proto_mod_path=$(sed -e 's,\(.*\)github.com.*,\1,' <<< "${gogo_proto_path}")
    local gogo_proto_imp_replace
    gogo_proto_imp_replace="Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types"
    gogo_proto_imp_replace="${gogo_proto_imp_replace},Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types"
    gogo_proto_imp_replace="${gogo_proto_imp_replace},Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types"
    gogo_proto_imp_replace="${gogo_proto_imp_replace},Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api"
    gogo_proto_imp_replace="${gogo_proto_imp_replace},Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types"
 
    local proto_go_path=${filename%%.proto}.pb.go
    local proto_go_bin_path=${filename%%.proto}.pb.binary.go
    
    local go_proto_out="paths=source_relative"
    if [ -z "${EXCLUDE_GRPC-}" ]; then
        go_proto_out="${go_proto_out},plugins=grpc"
    fi
 
    if [ -z "${DISABLE_IMPORT_REPLACE-}" ]; then
        go_proto_out="${go_proto_out},${gogo_proto_imp_replace}"
    fi
 
    if test -n "${go_proto_out}"; then
        go_proto_out="${go_proto_out}:"
    fi
 
    _echo_green "Generating ${filename} into ${proto_go_path} and ${proto_go_bin_path}"
    protoc \
        -I="${gogo_proto_path}/protobuf" \
        -I="${gogo_proto_path}" \
        -I="${gogo_proto_mod_path}" \
        -I="${source_dir}" \
        --gofast_out="${go_proto_out}${source_dir}" \
        --go-binary_out="${source_dir}" \
        "${filename}"
 
    local build_tags
    build_tags=$(sed -e '/^[[:space:]]*$/,$d' < "${filename}" | grep '// +build' || true)
    if test -n "${build_tags}";  then
        echo -e "${build_tags}\n" >> "${proto_go_path}.new"
        cat "${proto_go_path}" >> "${proto_go_path}.new"
        mv "${proto_go_path}.new" "${proto_go_path}"
        
        echo -e "${build_tags}\n" >> "${proto_go_bin_path}.new"
        cat "${proto_go_bin_path}" >> "${proto_go_bin_path}.new"
        mv "${proto_go_bin_path}.new" "${proto_go_bin_path}"
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
