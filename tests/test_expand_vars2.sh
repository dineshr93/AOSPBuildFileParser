#!/bin/bash
# Test AOSP variable expansion with Make - proper approach

cd ~/aosp

# Source the build environment
. build/envsetup.sh 2>/dev/null

echo "=== Testing Makefile with proper my-dir ==="

# Create test directory with Android.mk
mkdir -p /tmp/aosp_test
cd /tmp/aosp_test

# Create Android.mk
cat > Android.mk <<'EOF'
LOCAL_PATH := $(call my-dir)

# Test wildcard
TEST_WILDCARD := $(wildcard $(LOCAL_PATH)/*.c)

all:
	@echo "LOCAL_PATH=$(LOCAL_PATH)"
	@echo "TEST_WILDCARD=$(TEST_WILDCARD)"

# Create a test file
$(LOCAL_PATH)/test.c:
	@touch $(LOCAL_PATH)/test.c
EOF

# Create a .c file
touch /tmp/aosp_test/test.c

# Run Make
echo "Running Make to test wildcard expansion:"
make 2>&1

echo ""
echo "=== Now test with AOSP wildcard pattern ==="

# Test the exact pattern from AOSP
cat > Android2.mk <<'EOF'
LOCAL_PATH := $(call my-dir)
TARGET_BOARD_PLATFORM := hikey960

# Test the pattern from AOSP
EXTRA_FILE := $(if $(wildcard $(LOCAL_PATH)/Android.$(TARGET_BOARD_PLATFORM).mk), found, not-found)

all:
	@echo "LOCAL_PATH=$(LOCAL_PATH)"
	@echo "TARGET_BOARD_PLATFORM=$(TARGET_BOARD_PLATFORM)"
	@echo "EXTRA_FILE=$(EXTRA_FILE)"
EOF

echo "Running Make with AOSP pattern:"
make -f Android2.mk 2>&1

echo ""
echo "=== Check if file exists in AOSP tree ==="
cd ~/aosp/device/linaro/hikey/gralloc960
TARGET_BOARD_PLATFORM=hikey960
pattern="$(LOCAL_PATH)/Android.$(TARGET_BOARD_PLATFORM).mk"
echo "Looking for: $(pwd)/Android.$(TARGET_BOARD_PLATFORM).mk"
if [ -f "$(pwd)/Android.$(TARGET_BOARD_PLATFORM).mk" ]; then
    echo "FOUND: $(pwd)/Android.$(TARGET_BOARD_PLATFORM).mk"
else
    echo "NOT FOUND"
fi
