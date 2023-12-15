# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.8.5] - 2023-12-15
### :bug: Bug Fixes
- [`c9e040a`](https://github.com/turfaa/vmedis-proxy-api/commit/c9e040aff74c3ff3df963f1d0deaa28d6902c9b8) - 10 november *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.8.4] - 2023-12-14
### :bug: Bug Fixes
- [`9d5568f`](https://github.com/turfaa/vmedis-proxy-api/commit/9d5568fa0a6925e1948ae89e32cc29a9bf210db2) - remove unused variable *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.8.2] - 2023-12-11
### :bug: Bug Fixes
- [`f4f6035`](https://github.com/turfaa/vmedis-proxy-api/commit/f4f6035364d7264db8d7fbbcda537dd7193ccea1) - use 2 months for drugs to stock opname *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.8.1] - 2023-11-09
### :bug: Bug Fixes
- [`19c5e4f`](https://github.com/turfaa/vmedis-proxy-api/commit/19c5e4ff917a226f86fad4ea9654530a05e69af1) - **sales-statistics-dumper**: Add 30 seconds delay to fetch sales statistics *(commit by [@turfaa](https://github.com/turfaa))*
- [`59bc266`](https://github.com/turfaa/vmedis-proxy-api/commit/59bc266552040ad75c3f77d561c057304bc323a0) - **drugs-to-stock-opname-api**: Change threshold to 1 month *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.8.0] - 2023-11-03
### :sparkles: New Features
- [`e71b67c`](https://github.com/turfaa/vmedis-proxy-api/commit/e71b67c882dd33e14ba1506bdb954359a5927eeb) - Add invoice calculators API *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.7.0] - 2023-10-31
### :sparkles: New Features
- [`38ddcd9`](https://github.com/turfaa/vmedis-proxy-api/commit/38ddcd9c9efb34b296dcfdeb728cbb97d32db72a) - **drugs-to-stock-opname**: Support conservative mode *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.6.0] - 2023-10-25
### :sparkles: New Features
- [`70040b0`](https://github.com/turfaa/vmedis-proxy-api/commit/70040b02d98ede50920336d00db4b9011cd39a84) - **auth**: Create a user when login() doesn't find the user *(commit by [@turfaa](https://github.com/turfaa))*
- [`7235ee5`](https://github.com/turfaa/vmedis-proxy-api/commit/7235ee5a6220d464951744ee67a6e9d6f4b69403) - **sales-statistics**: Support getting sales statistics for other days *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.5.3] - 2023-10-24
### :bug: Bug Fixes
- [`ecc235e`](https://github.com/turfaa/vmedis-proxy-api/commit/ecc235ec371139c475c073d4a1efd37d642518ce) - **stock-opname-summaries**: Fix bug where some stock opnames data are skipped *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.5.2] - 2023-10-22
### :bug: Bug Fixes
- [`963275e`](https://github.com/turfaa/vmedis-proxy-api/commit/963275e3918ae59c98e41795cdb149b4ca8290e8) - **stock-opname-summary**: Change 'batch' to 'batchCode' to make it consistent *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.5.1] - 2023-10-22
### :bug: Bug Fixes
- [`ca26834`](https://github.com/turfaa/vmedis-proxy-api/commit/ca268340e65591fbace90b29b9157a53e6736f9e) - Return empty list when empty *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.5.0] - 2023-10-22
### :sparkles: New Features
- [`d88daec`](https://github.com/turfaa/vmedis-proxy-api/commit/d88daeceaceb07a792a15841fa0317208701106a) - Support multiple days queries *(commit by [@turfaa](https://github.com/turfaa))*
- [`a50eadb`](https://github.com/turfaa/vmedis-proxy-api/commit/a50eadb8892371a668c177ead83603ce427c56c4) - Remove Sun Oct 22 05:11:48 WIB 2023 in response schemas because it's not relevant anymore with the introduction of multiple days queries *(commit by [@turfaa](https://github.com/turfaa))*
- [`8c4ba61`](https://github.com/turfaa/vmedis-proxy-api/commit/8c4ba616e42aadfb300a645fe2af6c5fb50f763d) - Support stock opname summary *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.4.4] - 2023-10-20
### :bug: Bug Fixes
- [`8013f55`](https://github.com/turfaa/vmedis-proxy-api/commit/8013f5581da862fda1ff48531480c281f74840d4) - Support getSalesBetween(until, from) *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.4.3] - 2023-10-20
### :bug: Bug Fixes
- [`35b1a04`](https://github.com/turfaa/vmedis-proxy-api/commit/35b1a04c295973574611c3000fe7016addb5a1e5) - Use last month sales data for stock opname *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.4.2] - 2023-10-10
### :bug: Bug Fixes
- [`bc0a8d1`](https://github.com/turfaa/vmedis-proxy-api/commit/bc0a8d16e4b8071e6365928d01f67b2b83518f66) - Order stock opnames by vmedis_id *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.4.1] - 2023-10-08
### :bug: Bug Fixes
- [`2272265`](https://github.com/turfaa/vmedis-proxy-api/commit/227226528d105c4d448c95c3302dda9ea9912175) - Use strings.ToLower() when finding disallowed units *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.4.0] - 2023-10-08
### :sparkles: New Features
- [`423e1e6`](https://github.com/turfaa/vmedis-proxy-api/commit/423e1e6d95cde5350d53fddf87cfed75a01a8ed0) - Filter disallowed units *(commit by [@turfaa](https://github.com/turfaa))*

### :bug: Bug Fixes
- [`9a9be34`](https://github.com/turfaa/vmedis-proxy-api/commit/9a9be3489699bc76798fc603e236075370312b90) - set default last page as 1 *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.3.1] - 2023-10-07
### :bug: Bug Fixes
- [`6dfc00f`](https://github.com/turfaa/vmedis-proxy-api/commit/6dfc00f0e94a37a682c1495901605614388285dc) - **workflow**: don't generate latest tag *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.2.1] - 2023-10-03
### :bug: Bug Fixes
- [`0100cd6`](https://github.com/turfaa/vmedis-proxy-api/commit/0100cd6bed6eb722c050807374061dacb3f98870) - **api**: Remove unneeded caching *(commit by [@turfaa](https://github.com/turfaa))*


[v0.2.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.2.0...v0.2.1
[v0.3.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.3.0...v0.3.1
[v0.4.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.3.1...v0.4.0
[v0.4.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.4.0...v0.4.1
[v0.4.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.4.1...v0.4.2
[v0.4.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.4.2...v0.4.3
[v0.4.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.4.3...v0.4.4
[v0.5.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.4.4...v0.5.0
[v0.5.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.5.0...v0.5.1
[v0.5.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.5.1...v0.5.2
[v0.5.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.5.2...v0.5.3
[v0.6.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.5.3...v0.6.0
[v0.7.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.6.0...v0.7.0
[v0.8.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.7.0...v0.8.0
[v0.8.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.8.0...v0.8.1
[v0.8.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.8.1...v0.8.2
[v0.8.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.8.3...v0.8.4
[v0.8.5]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.8.4...v0.8.5