#!/bin/bash
# Test AOSP variable expansion with 'm' command

cd ~/aosp

# Source the build environment
. build/envsetup.sh 2>/dev/null

echo "=== Testing m command from AOSP tree ==="

# Try to build an existing module
cd ~/aosp/device/linaro/hikey/gralloc960
echo "Building from: $(pwd)"

# Set TOP
export TOP=~/aosp

# Try to build with m
echo "Building gralloc.hikey960:"
cd ~/aosp
m gralloc.hikey960 2>&1 | head -50
