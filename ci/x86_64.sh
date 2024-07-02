#!/usr/bin/env bash

go build -o drm .
if [ $? -ne 0 ]; then
    echo "Go build failed"
fi

# testfiles=`find ./ci/test -type f -name '*.dor'`
# for test in $testfiles; do
#     echo $test
# done

while IFS= read -r line; do
    funcinf=$(echo $line | tr ":" "\n")
    rc=-1
    for inf in $funcinf; do
        if [ "$rc" -ne -1 ]; then
            if [ $rc -ne $inf ]; then
                echo "Test Failed - expected $inf, got $rc"
                exit 1
            else
                echo "Test Succeeded"
            fi
            break
        fi
        echo "$inf"
        ./drm -a x86_64 ci/test/$inf.dor
        ./out/aarch64/$inf
        rc=$?
    done
done < ./ci/test/metadata.tests
