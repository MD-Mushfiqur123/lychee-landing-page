#!/bin/sh

set -eu

DOCKER_ORG=${DOCKER_ORG:-"lychee"}
FINAL_IMAGE_REPO=${FINAL_IMAGE_REPO:-"${DOCKER_ORG}/lychee"}

echo "Updating ${FINAL_IMAGE_REPO}:latest -> ${FINAL_IMAGE_REPO}:${VERSION}"
docker buildx imagetools create -t ${FINAL_IMAGE_REPO}:latest ${FINAL_IMAGE_REPO}:${VERSION}
echo "Updating ${FINAL_IMAGE_REPO}:rocm -> ${FINAL_IMAGE_REPO}:${VERSION}-rocm"
docker buildx imagetools create -t ${FINAL_IMAGE_REPO}:rocm ${FINAL_IMAGE_REPO}:${VERSION}-rocm

# Update GHCR latest tags if GITHUB_REPOSITORY_OWNER is set
if [ -n "${GITHUB_REPOSITORY_OWNER:-}" ]; then
    GHCR_REPO="ghcr.io/$(echo "${GITHUB_REPOSITORY_OWNER}" | tr '[:upper:]' '[:lower:]')/lychee"
    echo "Updating ${GHCR_REPO}:latest -> ${GHCR_REPO}:${VERSION}"
    docker buildx imagetools create -t ${GHCR_REPO}:latest ${GHCR_REPO}:${VERSION}
    echo "Updating ${GHCR_REPO}:rocm -> ${GHCR_REPO}:${VERSION}-rocm"
    docker buildx imagetools create -t ${GHCR_REPO}:rocm ${GHCR_REPO}:${VERSION}-rocm
fi
