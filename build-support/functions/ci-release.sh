function release_precheck {
   # Arguments: (yeah there are lots)
   #   $1 - Path to the top level Consul source
   #   $2 - boolean whether to tag the release yet
   #   $3 - boolean whether to build the binaries
   #   $4 - boolean whether to generate the sha256 sums
   #   $5 - version to set within version.go and the changelog
   #   $6 - release date to set within the changelog
   #   $7 - release version to set
   #   $8 - alternative gpg key to use for signing operations (optional)
   #
   # Returns:
   #   0 - success
   #   * - error

    echo "in release_precheck"
   debug "Source Dir:    $1"
   debug "Tag Release:   $2"
   debug "Build Release: $3"
   debug "Sign Release:  $4"
   debug "Version:       $5"
   debug "Release Date:  $6"
   debug "Release Vers:  $7"
   debug "GPG Key:       $8"

   if ! test -d "$1"
   then
      err "ERROR: '$1' is not a directory. build_release must be called with the path to the top level source as the first argument'"
      return 1
   fi

   if test -z "$2" -o -z "$3" -o -z "$4"
   then
      err "ERROR: build_release requires 4 arguments to be specified: <path to consul source> <tag release bool?> <build binaries bool?> <shasum 256 bool?>"
      return 1
   fi

   local sdir="$1"
   local do_tag="$2"
   local do_build="$3"
   local do_sha256="$4"
   local gpg_key="$8"

   if test -z "${gpg_key}"
   then
      gpg_key=${HASHICORP_GPG_KEY}
   fi

   if ! is_set "${RELEASE_UNSIGNED}"
   then
      if ! have_gpg_key "${gpg_key}"
      then
         err "ERROR: Aborting build because no useable GPG key is present. Set RELEASE_UNSIGNED=1 to bypass this check"
         return 1
      fi
   fi

   if ! is_git_clean "${sdir}" true && ! is_set "${ALLOW_DIRTY_GIT}"
   then
      err "ERROR: Refusing to build because Git is dirty. Set ALLOW_DIRTY_GIT=1 in the environment to proceed anyways"
      return 1
   fi

   local set_vers="$5"
   local set_date="$6"
   local set_release="$7"

   if test -z "${set_vers}"
   then
      set_vers=$(get_version "${sdir}" false false)
      set_release=$(parse_version "${sdir}" true false true)
   fi

   if is_set "${do_tag}" && ! set_release_mode "${sdir}" "${set_vers}" "${set_date}" "${set_release}"
   then
      err "ERROR: Failed to put source into release mode"
      return 1
   fi

   local vers="$(get_version ${sdir} true false)"
   if test $? -ne 0
   then
      err "Please specify a version (couldn't find one based on build tags)."
      return 1
   fi

   # Make sure we arent in dev mode
   unset CONSUL_DEV
}

function ui_smoke {
    # parse the version
    version=$(parse_version ".")

    # Check the version is baked in correctly
    status_stage "===> Checking version baked into UI"
    local ui_vers=$(ui_version "./ui-v2/dist/index.html")
    if test "${version}" != "${ui_vers}"
    then
        err "ERROR: UI version mismatch. Expecting: '${version}' found '${ui_vers}'"
        exit 1
    fi
    # Check the logo is baked in correctly
    status_stage "===> Checking logo baked into UI"
    local ui_logo_type=$(ui_logo_type "./ui-v2/dist/index.html")
    if test "${logo_type}" != "${ui_logo_type}"
    then
        err "ERROR: UI logo type mismatch. Expecting: '${logo_type}' found '${ui_logo_type}'"
        exit 1
    fi
}
function ci_release {
   # Arguments: (yeah there are lots)
   #   $1 - Path to the top level Consul source
   #   $2 - boolean whether to tag the release yet
   #   $3 - boolean whether to build the binaries
   #   $4 - boolean whether to generate the sha256 sums
   #   $5 - version to set within version.go and the changelog
   #   $6 - release date to set within the changelog
   #   $7 - release version to set
   #   $8 - alternative gpg key to use for signing operations (optional)
   #
   # Returns:
   #   0 - success
   #   * - error

    echo "in release_precheck"
   debug "Source Dir:    $1"
   debug "Tag Release:   $2"
   debug "Build Release: $3"
   debug "Sign Release:  $4"
   debug "Version:       $5"
   debug "Release Date:  $6"
   debug "Release Vers:  $7"
   debug "GPG Key:       $8"

   if ! test -d "$1"
   then
      err "ERROR: '$1' is not a directory. build_release must be called with the path to the top level source as the first argument'"
      return 1
   fi

   if test -z "$2" -o -z "$3" -o -z "$4"
   then
      err "ERROR: build_release requires 4 arguments to be specified: <path to consul source> <tag release bool?> <build binaries bool?> <shasum 256 bool?>"
      return 1
   fi

   local sdir="$1"
   local do_tag="$2"
   local do_build="$3"
   local do_sha256="$4"
   local gpg_key="$8"

   if test -z "${gpg_key}"
   then
      gpg_key=${HASHICORP_GPG_KEY}
   fi

   if ! is_set "${RELEASE_UNSIGNED}"
   then
      if ! have_gpg_key "${gpg_key}"
      then
         err "ERROR: Aborting build because no useable GPG key is present. Set RELEASE_UNSIGNED=1 to bypass this check"
         return 1
      fi
   fi

   if ! is_git_clean "${sdir}" true && ! is_set "${ALLOW_DIRTY_GIT}"
   then
      err "ERROR: Refusing to build because Git is dirty. Set ALLOW_DIRTY_GIT=1 in the environment to proceed anyways"
      return 1
   fi

   local set_vers="$5"
   local set_date="$6"
   local set_release="$7"

   if test -z "${set_vers}"
   then
      set_vers=$(get_version "${sdir}" false false)
      set_release=$(parse_version "${sdir}" true false true)
   fi

   if is_set "${do_tag}" && ! set_release_mode "${sdir}" "${set_vers}" "${set_date}" "${set_release}"
   then
      err "ERROR: Failed to put source into release mode"
      return 1
   fi

   local vers="$(get_version ${sdir} true false)"
   if test $? -ne 0
   then
      err "Please specify a version (couldn't find one based on build tags)."
      return 1
   fi

   # Make sure we arent in dev mode
   unset CONSUL_DEV
}