



# Build Scripts

This directory contains build scripts for gochecks.

## Available Scripts

- `build-multiplatform.sh`: Build binaries for multiple platforms and architectures
- `build-packages.sh`: Build packages (RPM, DEB, APK) for distribution

## Usage

### Building Multi-platform Binaries

```bash
# Build binaries with a specific tag (e.g., "v1.0.0")
./scripts/build-multiplatform.sh v1.0.0

# Build binaries with the default test tag
./scripts/build-multiplatform.sh
```

### Building Packages

```bash
# Build packages with a specific tag (e.g., "v1.0.0")
./scripts/build-packages.sh v1.0.0

# Build packages with the default test tag
./scripts/build-packages.sh
```


