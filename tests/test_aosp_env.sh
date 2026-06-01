#!/bin/bash
# Test AOSP build environment variable expansion

cd ~/aosp

echo "=== Sourcing AOSP build environment ==="
. build/envsetup.sh
echo "Environment sourced successfully"

echo ""
echo "=== Available lunch targets ==="
lunch --help 2>&1 | head -30

echo ""
echo "=== Checking if target exists ==="
# Try to find a valid target
for target in aosp_cf_x86_64_only_phone-aosp_current-userdebug aosp_arm64-aosp_current-userdebug; do
    echo "Testing: $target"
    lunch $target 2>&1 | tail -5
    if [ $? -eq 0 ]; then
        echo "Success! Target: $target"
        break
    fi
done

echo ""
echo "=== Test wildcard expansion with Make ==="
cd ~/aosp/device/linaro/hikey/gralloc960
echo "Testing Makefile with wildcard pattern:"
cat ~/aosp/device/linaro/hikey/gralloc960/Android.mk | grep -E "wildcard|include.*Android"
echo ""
echo "Running make to expand variables..."
make -f ~/aosp/device/linaro/hikey/gralloc960/Android.mk 2>&1 | head -20
