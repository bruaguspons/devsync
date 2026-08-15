#!/bin/bash
# Plain-bash assertion script (RED test, threat-matrix item) for install.sh's
# asset-filename validation. Run with: bash tools/devsync/install_test.sh
#
# It sources install.sh with DEVSYNC_SOURCE_ONLY=1 set, which must
# make the script define its functions (including validate_asset_filename)
# and return without executing the download/install flow. This lets us
# unit-test the filename guard without any network access.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_SH="${SCRIPT_DIR}/install.sh"

if [[ ! -f "${INSTALL_SH}" ]]; then
  echo "FAIL: ${INSTALL_SH} does not exist yet" >&2
  exit 1
fi

# shellcheck source=/dev/null
DEVSYNC_SOURCE_ONLY=1 source "${INSTALL_SH}"

fail=0

assert_rejects() {
  local filename="$1"
  if validate_asset_filename "${filename}"; then
    echo "FAIL: expected rejection of malformed asset filename '${filename}'" >&2
    fail=1
  else
    echo "PASS: rejected malformed asset filename '${filename}'"
  fi
}

assert_accepts() {
  local filename="$1"
  if ! validate_asset_filename "${filename}"; then
    echo "FAIL: expected acceptance of well-formed asset filename '${filename}'" >&2
    fail=1
  else
    echo "PASS: accepted well-formed asset filename '${filename}'"
  fi
}

# Malformed / unexpected asset names must be rejected before extraction.
assert_rejects "evil.sh"
assert_rejects "devsync_1.0.0_linux_amd64.zip"
assert_rejects "../../etc/passwd"
assert_rejects "devsync_1.0.0_windows_amd64.tar.gz"
assert_rejects ""

# Well-formed release asset names for supported platforms must be accepted.
assert_accepts "devsync_1.0.0_linux_amd64.tar.gz"
assert_accepts "devsync_1.0.0_linux_arm64.tar.gz"

# The skills.snapshot swap must be a remove-then-move (near-atomic),
# never a remove-then-copy: an interrupted `cp -R` can leave the
# installed directory half-copied, while an interrupted `mv` fails as
# "old removed, new absent" but never "half-copied".
assert_no_pattern_for_swap() {
  local pattern="$1"
  if grep -Eq "${pattern}" "${INSTALL_SH}"; then
    echo "FAIL: install.sh still contains the disallowed swap pattern '${pattern}'" >&2
    fail=1
  else
    echo "PASS: install.sh does not contain the disallowed swap pattern '${pattern}'"
  fi
}

assert_has_pattern() {
  local pattern="$1"
  if grep -Eq "${pattern}" "${INSTALL_SH}"; then
    echo "PASS: install.sh contains the required pattern '${pattern}'"
  else
    echo "FAIL: install.sh is missing the required pattern '${pattern}'" >&2
    fail=1
  fi
}

assert_no_pattern_for_swap 'cp -R.*skills\.snapshot'
assert_has_pattern 'mv.*skills\.snapshot'

if [[ "${fail}" -ne 0 ]]; then
  echo "install_test.sh: FAILURES DETECTED" >&2
  exit 1
fi

echo "install_test.sh: all assertions passed"
