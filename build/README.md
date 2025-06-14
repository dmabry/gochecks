




# Build Configuration

This directory contains configuration files for building gochecks packages.

## Package Configuration

The `package/nfpm.yaml` file defines the package metadata and contents for creating RPM, DEB, and APK packages using [nfpm](https://github.com/goreleaser/nfpm).

### Package Contents

The following binaries are included in the packages:

- `/usr/lib/nagios/plugins/check_interfaces`
- `/usr/lib/nagios/plugins/check_interface_usage`
- `/usr/lib/nagios/plugins/check_sysdescr`

## Building Packages

To build packages, use the `scripts/build-packages.sh` script.

```bash
# Build packages with a specific tag (e.g., "v1.0.0")
./scripts/build-packages.sh v1.0.0

# Build packages with the default test tag
./scripts/build-packages.sh
```

The resulting packages will be available in the `build/package/release` directory.



