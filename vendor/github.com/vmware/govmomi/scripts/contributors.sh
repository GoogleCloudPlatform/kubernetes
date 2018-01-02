#!/bin/bash -e

file="$(git rev-parse --show-toplevel)/CONTRIBUTORS"

cat <<EOF > "$file"
# People who can (and typically have) contributed to this repository.
#
# This script is generated by $(basename "$0")
#

EOF

git log --format='%aN <%aE>' | sort -uf >> "$file"
