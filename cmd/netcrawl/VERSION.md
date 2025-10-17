# NetCrawl Version Management

## Current Version
**v1.1.0** - Refactored eventstream package architecture

## Version History

### v1.1.0 (2025-10-17)
- Refactored eventstream package to use instance-based architecture
- Moved package prefix detection from global init() to per-Handler detectPackagePrefix() method
- Removed global variables for cleaner, more maintainable design
- Updated examples to use new Handler-based API
- Comprehensive test suite rebuild
- All existing functionality preserved with improved architecture

### v1.0.0 (2025-10-17)
- Initial release
- HP ProCurve/Aruba switch support
- SSH connectivity with network device optimizations
- Command execution: show version, show running-config, show lldp neighbors detail
- Parser for HP/Aruba command outputs
- Interface IP addresses, VRF assignments, VLAN tagging
- LLDP neighbor discovery
- JSON database storage using dsjdb
- Command-line flags: -device, -port, -timeout
- Raw output saved to text files

## Semantic Versioning

NetCrawl follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR version** (X.0.0): Incompatible API changes or breaking changes
- **MINOR version** (0.X.0): Backward-compatible functionality additions
- **PATCH version** (0.0.X): Backward-compatible bug fixes

### When to Increment

#### MAJOR (Breaking Changes)
- Change command-line flag names or remove flags
- Change JSON output structure in incompatible ways
- Remove or rename exported functions/types in libraries
- Change default behavior that breaks existing workflows

#### MINOR (New Features)
- Add new command-line flags (with defaults)
- Add support for new device vendors
- Add new parsing capabilities
- Add new fields to JSON output (backward-compatible)
- Add new commands to execute

#### PATCH (Bug Fixes)
- Fix parsing bugs
- Fix connection issues
- Security patches
- Documentation-only changes
- Performance improvements without API changes

## Updating the Version

To update the version, edit the `Version` constant in `cmd/netcrawl/netcrawl.go`:

```go
const Version = "vX.Y.Z"
```

Then rebuild and install:

```bash
make install-netcrawl
```

## Checking the Version

```bash
# Show version only
netcrawl -version

# Version is also displayed when running normally
netcrawl -device 192.168.1.1
```
