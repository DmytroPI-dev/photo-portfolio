# Third-Party Notices

This file records third-party software included in deployable project artifacts.
It does not change the licence of this repository's own source code.

## Image Worker

`services/gallery-api/cmd/image-worker/Dockerfile` builds a custom ARM64
container image for AWS Lambda. The image includes:

| Component | Purpose | Licence | Source |
| --- | --- | --- | --- |
| libvips | Image decoding, resizing, EXIF orientation, and WebP encoding | LGPL-2.1-or-later | [libvips source and licence](https://github.com/libvips/libvips) |
| Alpine Linux packages | Base operating-system libraries and the `vips`/`vipsheader` command-line tools | Package-specific | [Alpine package repository](https://pkgs.alpinelinux.org/packages) |

The Go worker invokes the `vips` command-line program; it does not embed or
link libvips into the Go executable. The worker container tested locally used
Alpine `vips` and `vips-tools` version `8.15.3-r5`.

## Release Checklist

Before publishing or distributing a worker image outside the private deployment
environment:

1. Preserve this notice with the image source and include the full LGPL-2.1-or-
   later licence text in the distributed image or accompanying release notices.
2. Record the exact Alpine image digest and installed package versions for the
   release, preferably as an SBOM attached to the immutable ECR image digest.
3. Preserve the upstream source location and licence notices for libvips and
   any other bundled package whose licence requires attribution or source
   availability.
4. Recheck notices when changing the image base, libvips version, or image
   processing implementation.

This is an engineering compliance reminder, not legal advice.
