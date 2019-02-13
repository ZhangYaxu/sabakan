# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.2.0] - 2019-02-13

### Added
- `Machine.Info` brings NIC configuration information (#136).  
    This new information is also exposed in GraphQL and REST API.
- `ipam.json` adds new mandatory field `node-gateway-offset` (#136).  
    Existing installations continue to work thanks to automatic data conversion.

### Changed
- GraphQL data type `BMCInfoIPv4` is renamed to `NICConfig`.

### Removed
- `dhcp.json` obsoletes `gateway-offset` field (#136).  
    The field is moved to `ipam.json` as `node-gateway-offset`.

## [1.1.0] - 2019-01-29

### Added
- [ignition] `json` template function to render objects in JSON (#134).

## [1.0.1] - 2019-01-28

### Changed
- Fix a regression in ignition template introduced in #131 (#133).

## [1.0.0] - 2019-01-28

### Breaking changes
- `ipam.json` adds new mandatory field `bmc-ipv4-gateway-offset` (#132).
- Ignition template renderer sets `.` as `Machine` instead of `MachineSpec` (#132).

### Added
- `Machine` has additional information field for BMC NIC configuration (#132).

## Ancient changes

See [CHANGELOG-0](./CHANGELOG-0.md).

[Unreleased]: https://github.com/cybozu-go/sabakan/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/cybozu-go/sabakan/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/cybozu-go/sabakan/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/cybozu-go/sabakan/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/cybozu-go/sabakan/compare/v0.31...v1.0.0
