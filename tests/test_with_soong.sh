#!/bin/bash
# Test AOSP variable expansion with Soong

cd ~/aosp

echo "=== Testing Soong build system ==="

# Try to use Soong directly
export TOP=~/aosp
export OUT_DIR=~/aosp/out

echo "Checking if soong_ui.bash exists:"
ls -la ~/aosp/build/soong/soong_ui.bash 2>&1

echo ""
echo "Trying to run soong_ui.bash --make-mode:"
cd ~/aosp
build/soong/soong_ui.bash --make-mode 2>&1 | head -50
