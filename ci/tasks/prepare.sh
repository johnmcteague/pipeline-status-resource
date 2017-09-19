#!/bin/bash -e
mkdir -p prepare-output/assets
cp source/Dockerfile prepare-output/.
cp releases/check prepare-output/assets/check
cp releases/in prepare-output/assets/in
cp releases/out prepare-output/assets/out
