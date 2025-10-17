# FDOT Project Version

## Current Version
**v1.1.0** - Refactored eventstream package architecture

## Version History

### v1.1.0 (2025-10-17)
- **MINOR**: Refactored eventstream package architecture
  - Moved from global state to instance-based design
  - Removed `init()` function and global variables
  - Added per-Handler package prefix detection
  - Updated examples and rebuilt test suite
  - Backward compatible Handler API maintained
  - Improved code maintainability and testability

### v1.0.0 (2025-10-17)
- **MAJOR**: Initial release
  - NetCrawl tool with HP ProCurve/Aruba switch support
  - Eventstream package for event logging and streaming
  - Credential manager with cross-platform support
  - Network device abstraction and SSH connectivity

## Semantic Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes