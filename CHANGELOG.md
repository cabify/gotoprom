# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Support for labels using unsigned integers.

## [1.0.0] - 2020-01-10
### Added
- Support for empty buckets tag, which will generate nil buckets for the prometheus Histogram and use default prometheus buckets.
- Support for empty objectives tag, which will generate nil objectives for the prometheus Summary and use an empty objectives map after all.

### Changed
- **Breaking**: `prometheus.Histogram` is now used to build histograms, instead of `prometheus.Observer`, which means that previous code building `prometheus.Observer` won't compile anymore.

### Removed
- **Breaking**: default buckets on histograms. All histogram should explicitly specify their buckets now or they will fail to build.
- **Breaking**: default objectives on summaries. All summaries should explicitly specify their objectives now or they will fail to build.

### Fixed
- Summary building was not failing with malformed objectives.

### Migration instructions
If you're migrating from a `v0.x.y`, you will need to:
- Replace `Metric func() prometheus.Observer` by `Metric func() prometheus.Histogram`
- On `prometheus.Histogram` metrics, add `buckets:""` which will use the `prometheus.DefBuckets` bucekts, or even better, define yours.
- On `prometheus.Summary` metrics, add `objectives:".5,.95,.99"` to keep using the same objectives, or define yours.

## [0.3.0] - 2019-10-10
### Added
- Add objectives to summaries through struct tag and set default values when none specified
### Changed
- Upgraded client_golang to v1

## [0.2.1] - 2019-06-05
### Changed
- Reduced the number of default buckets from 12 to 7 between 0.05s and 10s

## [0.2.0] - 2019-05-20
### Fixed
- Included the implementation builder for summaries [#14](https://github.com/cabify/gotoprom/pull/14)

## [0.1.1] - 2019-05-08
### Fixed
- Not failing when embedded labels are wrong [#13](https://github.com/cabify/gotoprom/pull/13) 

## [0.1.0] - 2019-05-07
### Added
- All the code for the initial open source release

[Unreleased]: https://github.com/cabify/gotoprom/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/cabify/gotoprom/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/cabify/gotoprom/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/cabify/gotoprom/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/cabify/gotoprom/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/cabify/gotoprom/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/cabify/gotoprom/compare/v0.1.0

