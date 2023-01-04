#!/bin/bash

my_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$my_dir/.."

coverdir="$(mktemp -d)"
stdout="$(mktemp)"

die () {
    echo fatal: "$@" >&2
    exit 1
}

sed-i () {
    # what if we're running on a BSD sed (like on macOS)?
    local is_gnu_sed=false
    sed --version 2>&1 | grep 'GNU sed' > /dev/null && is_gnu_sed=true
    if "$is_gnu_sed"; then
        sed -i "$@"
    else
        # assume everything else is BSD-like
        sed -i '' "$@"
    fi
}

GOCOVERDIR="$coverdir" ./gosmopolitan \
    -escapehatches '(github.com/xen0n/gosmopolitan/testdata/pkgFoo).escapeHatch,(github.com/xen0n/gosmopolitan/testdata/pkgFoo).pri18ntln,(github.com/xen0n/gosmopolitan/testdata/pkgFoo).i18nMessage' \
    ./testdata/pkgFoo > "$stdout" 2>&1 && die "return code should be non-zero"

sed-i 's@^.*/testdata/pkgFoo/@ROOT/@' "$stdout"
diff -u ./testdata/pkgFoo/expected.txt "$stdout" || die "unexpected linter output"
rm "$stdout"

go tool covdata textfmt -i="$coverdir" -o ./coverage.txt
rm -rf "$coverdir"

# report to Codecov if running in CI
if [[ -n $CI ]]; then
    "$my_dir"/codecov.sh
fi
