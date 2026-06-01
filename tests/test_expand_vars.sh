#!/bin/bash
# Test AOSP variable expansion with Make

cd ~/aosp

# Source the build environment
. build/envsetup.sh 2>/dev/null

echo "=== Testing Makefile variable expansion ==="

# Create a test Makefile with wildcard
cat > /tmp/test.mk <<'EOF'
LOCAL_PATH := $(call my-dir)
TEST_FILES := $(wildcard $(LOCAL_PATH)/*.txt)
TEST_FILES2 := $(wildcard $(LOCAL_PATH)/*.md)
all:
	@echo "LOCAL_PATH=$(LOCAL_PATH)"
	@echo "TEST_FILES=$(TEST_FILES)"
	@echo "TEST_FILES2=$(TEST_FILES2)"
EOF

# Run Make to expand variables
echo ""
echo "Testing wildcard expansion:"
cd /tmp && make -f test.mk 2>&1

echo ""
echo "=== Testing with actual AOSP file ==="

# Test with the actual Android.mk that has wildcard
cd ~/aosp/device/linaro/hikey/gralloc960

# Create a minimal test that just evaluates the wildcard
cat > /tmp/test_aosp.mk <<'EOF'
LOCAL_PATH := $(call my-dir)
TARGET_BOARD_PLATFORM := hikey960
EXTRA_FILE := $(if $(wildcard $(LOCAL_PATH)/Android.$(TARGET_BOARD_PLATFORM).mk), found, not-found)
all:
	@echo "LOCAL_PATH=$(LOCAL_PATH)"
	@echo "TARGET_BOARD_PLATFORM=$(TARGET_BOARD_PLATFORM)"
	@echo "EXTRA_FILE=$(EXTRA_FILE)"
EOF

cd ~/aosp/device/linaro/hikey/gralloc960 && make -f /tmp/test_aosp.mk 2>&1

echo ""
echo "=== Checking if wildcard pattern resolves ==="
pattern="$(pwd)/Android.${TARGET_BOARD_PLATFORM}.mk"
echo "Pattern: $pattern"
if [ -f "$pattern" ]; then
    echo "File exists: $pattern"
else
    echo "File does not exist: $pattern"
fi
