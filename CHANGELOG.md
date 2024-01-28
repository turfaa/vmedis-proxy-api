# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.11.2] - 2024-01-28
### :bug: Bug Fixes
- [`a83b455`](https://github.com/turfaa/vmedis-proxy-api/commit/a83b4553d3c06790bff6fea827e597d4c88c8398) - Add drug.Stock MarshalText *(commit by [@turfaa](https://github.com/turfaa))*
- [`9f882aa`](https://github.com/turfaa/vmedis-proxy-api/commit/9f882aaaf2cc298ced89cffc2963ced68560b044) - Fix dumper schedule *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.11.1] - 2024-01-28
### :bug: Bug Fixes
- [`2cafb37`](https://github.com/turfaa/vmedis-proxy-api/commit/2cafb37bdc084ae2197842ae070d9ccd34b0f5fd) - **Dockerfile**: Copy all repository contents *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.11.0] - 2024-01-28
### :sparkles: New Features
- [`00eb975`](https://github.com/turfaa/vmedis-proxy-api/commit/00eb975e025d45dacb612538d9e6d3351543cf22) - Produce dumped drugs to Kafka *(commit by [@turfaa](https://github.com/turfaa))*
- [`e827963`](https://github.com/turfaa/vmedis-proxy-api/commit/e827963d748e5eaba4246c730d3500cdf78bde22) - Produce sold drugs and stock opnamed drugs to Kafka *(commit by [@turfaa](https://github.com/turfaa))*
- [`6d66d1d`](https://github.com/turfaa/vmedis-proxy-api/commit/6d66d1df23d57ac1c0111f166dd51c75b9e8f15b) - Use consumer to dump drug details *(commit by [@turfaa](https://github.com/turfaa))*

### :recycle: Refactors
- [`34e7703`](https://github.com/turfaa/vmedis-proxy-api/commit/34e7703928553227cad3d1e11ed2bcfd31f25a1c) - flatten project structure *(commit by [@turfaa](https://github.com/turfaa))*
- [`b6b5718`](https://github.com/turfaa/vmedis-proxy-api/commit/b6b57182ab8b28e93da0dcc9c743ae0ffd96b508) - move drugs-related code to domain-based package *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.10.1] - 2024-01-23
### :bug: Bug Fixes
- [`206a28c`](https://github.com/turfaa/vmedis-proxy-api/commit/206a28c0860bf35630fe2041b21a03a6eef4232f) - Do not omit empty on drug units and stocks *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.10.0] - 2024-01-23
### :sparkles: New Features
- [`11576b8`](https://github.com/turfaa/vmedis-proxy-api/commit/11576b828872d9d59bc62759b578d06c06ab9f54) - Support parsing stocks in drug details page *(commit by [@turfaa](https://github.com/turfaa))*
- [`eb04263`](https://github.com/turfaa/vmedis-proxy-api/commit/eb04263673c13158bed5d23df47cf1f861bcca59) - Support storing drug stocks *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.9] - 2024-01-08
### :bug: Bug Fixes
- [`5f25df7`](https://github.com/turfaa/vmedis-proxy-api/commit/5f25df7c926bb34ae9868d36359b6c737b210e88) - start stock opname from 2023-01-06 *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.8] - 2024-01-08
### :bug: Bug Fixes
- [`2d637ee`](https://github.com/turfaa/vmedis-proxy-api/commit/2d637eeea7cd46dab2e36734f2ee4f9ce28f63d9) - start stock opname from 2023-01-09 *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.7] - 2024-01-06
### :bug: Bug Fixes
- [`0140ef9`](https://github.com/turfaa/vmedis-proxy-api/commit/0140ef906844f27439162b3a190398d65fa0a4e0) - initiate summaries even if there's nothing so that the json will *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.6] - 2024-01-04
### :bug: Bug Fixes
- [`3e719df`](https://github.com/turfaa/vmedis-proxy-api/commit/3e719df710b706c90c10e7299be6492d1ad0bc3d) - update stock opname date *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.5] - 2023-12-22
### :bug: Bug Fixes
- [`4ea7534`](https://github.com/turfaa/vmedis-proxy-api/commit/4ea7534a6b8562f8030e71612584813b09751287) - return to today *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.4] - 2023-12-22
### :bug: Bug Fixes
- [`b16106b`](https://github.com/turfaa/vmedis-proxy-api/commit/b16106b9e76a3f44de243a915e8ef3d8a15b976c) - add empty sale unit check *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.3] - 2023-12-22
### :bug: Bug Fixes
- [`343d532`](https://github.com/turfaa/vmedis-proxy-api/commit/343d5328c967a406f2b90afe1e65d7af01018ae8) - fix url *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.2] - 2023-12-22
### :bug: Bug Fixes
- [`4169552`](https://github.com/turfaa/vmedis-proxy-api/commit/4169552fe7671c699ea415f7a0ae391238e03649) - use data from 21 dec to backfill data *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.1] - 2023-12-19
### :bug: Bug Fixes
- [`e508e13`](https://github.com/turfaa/vmedis-proxy-api/commit/e508e1350e3cf4e21f2298b0abdc9478b7963414) - change sales statistics dumper to run at *:59:30 *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.9.0] - 2023-12-18
### :sparkles: New Features
- [`5859c91`](https://github.com/turfaa/vmedis-proxy-api/commit/5859c914b787d9ec72e5585b0051c8eff6e4429e) - Add daily history *(commit by [@turfaa](https://github.com/turfaa))*


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
[v0.9.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.8.5...v0.9.0
[v0.9.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.0...v0.9.1
[v0.9.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.1...v0.9.2
[v0.9.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.2...v0.9.3
[v0.9.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.3...v0.9.4
[v0.9.5]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.4...v0.9.5
[v0.9.6]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.5...v0.9.6
[v0.9.7]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.6...v0.9.7
[v0.9.8]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.7...v0.9.8
[v0.9.9]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.8...v0.9.9
[v0.10.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.9.9...v0.10.0
[v0.10.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.10.0...v0.10.1
[v0.11.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.10.1...v0.11.0
[v0.11.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.0...v0.11.1
[v0.11.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.1...v0.11.2