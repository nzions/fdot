# FDOT Project Version

## Current Version
**v1.2.0** - credmgr OO API migration and test optimization

## Version History

### v1.2.0 (2025-01-17)
- **MINOR**: credmgr package improvements and test optimization
  - credmgr package v3.0.0: Migrated to object-oriented (OO) API
    - Changed from package-level functions to CredManager interface
    - Added `New(path)` and `Default()` factory functions
    - Platform-specific implementations: windowsCredManager, linuxCredManager, otherCredManager
    - Maintains backward compatibility via wrapper functions (deprecated)
  - Updated cmd/credmgr CLI to use OO API
  - Updated fuser package to use CredManager instance
  - Test suite rationalization:
    - Removed 201 lines of duplicate and deprecated test code
    - Changed tests to use isolated temp directories (os.MkdirTemp)
    - No longer uses default paths (~/.fdot/, ~/.local/credmgr/)
    - Removed credmgr_oo_test.go (162 lines of duplicates)
    - Removed TestDeprecatedFunctions (39 lines)
  - Updated credmgr_api examples to v3.0.0 OO API
  - All 25 tests passing with improved isolation

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