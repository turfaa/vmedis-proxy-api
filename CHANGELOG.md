# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.16.4] - 2024-02-29
### :bug: Bug Fixes
- [`be6cc7b`](https://github.com/turfaa/vmedis-proxy-api/commit/be6cc7b690a1f7d7bfd3991880f3de66f3aab009) - Log email *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.16.3] - 2024-02-29
### :bug: Bug Fixes
- [`79dbc0b`](https://github.com/turfaa/vmedis-proxy-api/commit/79dbc0b81a4301c967019f469ce1fe6b3ac0b0b6) - Log email *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.16.2] - 2024-02-29
### :bug: Bug Fixes
- [`8063ccf`](https://github.com/turfaa/vmedis-proxy-api/commit/8063ccf7bc4a9f15fd065163faafc3664e551eea) - Log emailer *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.16.1] - 2024-02-29
### :bug: Bug Fixes
- [`973da86`](https://github.com/turfaa/vmedis-proxy-api/commit/973da8659904fd24788e97994b171e3512f01951) - Add env key replacer, replacing '.' with '_' *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.16.0] - 2024-02-29
### :sparkles: New Features
- [`0891058`](https://github.com/turfaa/vmedis-proxy-api/commit/08910583ed2a0242f6ac37203a5deb37dc4946f6) - Support sending procurement report to IQVIA *(commit by [@turfaa](https://github.com/turfaa))*
- [`71c5ff0`](https://github.com/turfaa/vmedis-proxy-api/commit/71c5ff0f22539e8f3ddebe5fbf043450d3b6f7e5) - Support sending sales report to IQVIA *(commit by [@turfaa](https://github.com/turfaa))*

### :recycle: Refactors
- [`f435f04`](https://github.com/turfaa/vmedis-proxy-api/commit/f435f04d5b17e973e5f79224764b9cda3aa6f513) - Create pkg2 package *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.15.0] - 2024-02-24
### :sparkles: New Features
- [`6846fd1`](https://github.com/turfaa/vmedis-proxy-api/commit/6846fd14ceb4ae26b2bd980e0a8a7483ccc4a9e6) - Add stock opname summaries API *(commit by [@turfaa](https://github.com/turfaa))*

### :recycle: Refactors
- [`3e91588`](https://github.com/turfaa/vmedis-proxy-api/commit/3e91588d2129335350aeb69bc353e13c9e557f64) - Move stock opname codes to a new package *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.14.0] - 2024-02-21
### :sparkles: New Features
- [`f82e38c`](https://github.com/turfaa/vmedis-proxy-api/commit/f82e38c4afff8fabc7c889ee2844d96d8437215d) - Divide token provider and refresher, create a command to refresh tokens *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.6] - 2024-02-21
### :bug: Bug Fixes
- [`23ea1b1`](https://github.com/turfaa/vmedis-proxy-api/commit/23ea1b16d7666c3b7a194c9af350ccb4012770e5) - remove procurement-related jobs from dumper *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.5] - 2024-02-21
### :bug: Bug Fixes
- [`8711fd1`](https://github.com/turfaa/vmedis-proxy-api/commit/8711fd119ac54b420a21ffd81916c27fdb3f8a12) - dump procurements until the end of today *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.4] - 2024-02-19
### :bug: Bug Fixes
- [`8bf20a5`](https://github.com/turfaa/vmedis-proxy-api/commit/8bf20a51edb31b511d9ac454814e652e727d6d9e) - start stock opname from 2024-02-19 *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.3] - 2024-02-12
### :bug: Bug Fixes
- [`7a8b378`](https://github.com/turfaa/vmedis-proxy-api/commit/7a8b3783385ac31ff2d0b826ec2ff0b5c2f3f172) - Exit when drug consumer returns error *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.2] - 2024-02-08
### :bug: Bug Fixes
- [`5da4e63`](https://github.com/turfaa/vmedis-proxy-api/commit/5da4e635a534b5c636f8f39e02d1d90b940daf67) - Update drug dumper schedule, avoid 00.00 (vmedis is always down at 00.00) *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.1] - 2024-02-08
### :bug: Bug Fixes
- [`16d1fce`](https://github.com/turfaa/vmedis-proxy-api/commit/16d1fce51262c7298e99908723d5e4bfd5b4ab45) - move limiter.Wait() to the top of getWithSessionId *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.13.0] - 2024-02-08
### :sparkles: New Features
- [`5b37563`](https://github.com/turfaa/vmedis-proxy-api/commit/5b37563328cdb5f7abbfe053f1d0f33102933beb) - Move vmedis session ids to database *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.12.2] - 2024-02-08
### :bug: Bug Fixes
- [`f38b345`](https://github.com/turfaa/vmedis-proxy-api/commit/f38b345bf3402e8b41dd0ed46c3d8e28de92ed1f) - Fix procurement cron schedule *(commit by [@turfaa](https://github.com/turfaa))*

### :wrench: Chores
- [`3d46f1d`](https://github.com/turfaa/vmedis-proxy-api/commit/3d46f1d4ac50d18a8fcba52c707a4c925ee45c84) - Remove unused proxy/schema *(commit by [@turfaa](https://github.com/turfaa))*
- [`12f9722`](https://github.com/turfaa/vmedis-proxy-api/commit/12f9722196efed0a37723d65a1f80377e2c6b173) - Remove unused proxy/filter_units.go *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.12.1] - 2024-02-08
### :bug: Bug Fixes
- [`39821c3`](https://github.com/turfaa/vmedis-proxy-api/commit/39821c39b0ac772d672bf359d68dad95f1229cbf) - Inline DrugStock in procurement recommendations using the experimental encoding/json/v2 candidate *(commit by [@turfaa](https://github.com/turfaa))*

### :wrench: Chores
- [`33e0da7`](https://github.com/turfaa/vmedis-proxy-api/commit/33e0da7318e4edb741721a6edb4c86461c2157cd) - Update all deps *(commit by [@turfaa](https://github.com/turfaa))*
- [`8a7c704`](https://github.com/turfaa/vmedis-proxy-api/commit/8a7c704e8c972213b4098c7f8307444c0bf90e1d) - use protojson instead of jsonpb *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.12.0] - 2024-02-08
### :sparkles: New Features
- [`5878296`](https://github.com/turfaa/vmedis-proxy-api/commit/58782964a7a73b91d91b8e2cc0603d4fe58aa1e7) - Add procurements dumper *(commit by [@turfaa](https://github.com/turfaa))*
- [`42f3e1e`](https://github.com/turfaa/vmedis-proxy-api/commit/42f3e1e479e71de4906e3140ac4f2157312a47d9) - Support dumping procurements via API *(commit by [@turfaa](https://github.com/turfaa))*

### :recycle: Refactors
- [`2db79d0`](https://github.com/turfaa/vmedis-proxy-api/commit/2db79d0141a2d1a6a91c52d8c096212b2cf68c49) - move procurement recommendations to the procurement package *(commit by [@turfaa](https://github.com/turfaa))*
- [`07d237f`](https://github.com/turfaa/vmedis-proxy-api/commit/07d237f9768192791756e44a6ca551ff4fe0a738) - move invoice calculators to the procurement package *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.11.5] - 2024-01-30
### :bug: Bug Fixes
- [`7c7f608`](https://github.com/turfaa/vmedis-proxy-api/commit/7c7f60816d4780e127d21741942ee0b8ef0535e8) - Put stock opname id as stock opname request key *(commit by [@turfaa](https://github.com/turfaa))*

### :wrench: Chores
- [`623a6a8`](https://github.com/turfaa/vmedis-proxy-api/commit/623a6a8898d7dfdf498897416e9e5d8a3adc1576) - do not use cache when building *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.11.4] - 2024-01-30
### :bug: Bug Fixes
- [`8134cbd`](https://github.com/turfaa/vmedis-proxy-api/commit/8134cbd2a2eed3fb3cabeab575c23ccc715df61b) - Allow zero drug stock *(commit by [@turfaa](https://github.com/turfaa))*


## [v0.11.3] - 2024-01-28
### :bug: Bug Fixes
- [`37df300`](https://github.com/turfaa/vmedis-proxy-api/commit/37df300b6a69d076d7cda05e68cc3c1cf5370164) - Add logs in drug consumer *(commit by [@turfaa](https://github.com/turfaa))*
- [`7a6f12b`](https://github.com/turfaa/vmedis-proxy-api/commit/7a6f12b8163f589ca0d43a359f06f28c9f69f186) - Add docker cache for building *(commit by [@turfaa](https://github.com/turfaa))*


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
[v0.11.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.2...v0.11.3
[v0.11.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.3...v0.11.4
[v0.11.5]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.4...v0.11.5
[v0.12.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.11.5...v0.12.0
[v0.12.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.12.0...v0.12.1
[v0.12.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.12.1...v0.12.2
[v0.13.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.12.2...v0.13.0
[v0.13.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.0...v0.13.1
[v0.13.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.1...v0.13.2
[v0.13.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.2...v0.13.3
[v0.13.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.3...v0.13.4
[v0.13.5]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.4...v0.13.5
[v0.13.6]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.5...v0.13.6
[v0.14.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.13.6...v0.14.0
[v0.15.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.14.0...v0.15.0
[v0.16.0]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.15.0...v0.16.0
[v0.16.1]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.16.0...v0.16.1
[v0.16.2]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.16.1...v0.16.2
[v0.16.3]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.16.2...v0.16.3
[v0.16.4]: https://github.com/turfaa/vmedis-proxy-api/compare/v0.16.3...v0.16.4