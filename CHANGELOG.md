# Changelog

## [1.0.2](https://github.com/grasse-oss/cron-set-controller/compare/cron-set-controller-v1.0.1...cron-set-controller-v1.0.2) (2023-09-01)


### Bug Fixes

* **deployment:** just use release full name to be named deployment ([b3aa5a2](https://github.com/grasse-oss/cron-set-controller/commit/b3aa5a2051a76a0c2da502f4267f4f3a1ce42759))
* **deployment:** use chat app version instead of tag value ([e3249ff](https://github.com/grasse-oss/cron-set-controller/commit/e3249ff0a46989cf37e2a2a321905ab0c576bad3))
* **image-tag:** use image tag for special case ([a7c7cac](https://github.com/grasse-oss/cron-set-controller/commit/a7c7cac9523af4b8983b37a3b2b29605a675a43d))

## [1.0.1](https://github.com/grasse-oss/cron-set-controller/compare/cron-set-controller-v1.0.0...cron-set-controller-v1.0.1) (2023-09-01)


### Bug Fixes

* **crd:** remove the label propety of CRD ([1ef29c6](https://github.com/grasse-oss/cron-set-controller/commit/1ef29c629169fe12ad3deb288fcf0a724ba900c0))
* **ctrl:** add some more logs ([6b14ec0](https://github.com/grasse-oss/cron-set-controller/commit/6b14ec00bdcd5ab18e59cfd3b68cdcc827f4ddec))
* **ctrl:** add the logs to controller ([84742aa](https://github.com/grasse-oss/cron-set-controller/commit/84742aab2e0a69901627329ee15a791f016ce34d))
* **ctrl:** Modify the logs to match the logger format. ([9ba7503](https://github.com/grasse-oss/cron-set-controller/commit/9ba7503e907ad01ad7a321c0957e62517c3a01d4))

## 1.0.0 (2023-08-31)


### Features

* create update and delete cronjobs ([2a8843d](https://github.com/grasse-oss/cron-set-controller/commit/2a8843d71699aaaf1ec6bfcc07a9e23b82ce4620))
* **cronset-api:** scaffold cronset api ([5df3d9a](https://github.com/grasse-oss/cron-set-controller/commit/5df3d9a437eae702de97af146a135b80ee5150cc))
* **cronset-spec:** add cronset spec ([d51b3df](https://github.com/grasse-oss/cron-set-controller/commit/d51b3df521be7b64b393ae156eaaa420515a23c6))
* **cronset-spec:** add cronset status ([1e9f261](https://github.com/grasse-oss/cron-set-controller/commit/1e9f2613d1a0d5033d514bc4288d7e9d25fd6da6))
* **cronset-spec:** add json/protobuf tag to CronSetStatus ([042b904](https://github.com/grasse-oss/cron-set-controller/commit/042b9042778d7cc70fda8cc80a70ed975706dffd))
* generate deep copy for cronset CRD ([3e76953](https://github.com/grasse-oss/cron-set-controller/commit/3e7695368bd08649923589d03ade7a364e79ffec))
* get the list of nodes ([12d4c90](https://github.com/grasse-oss/cron-set-controller/commit/12d4c90c18b45c981434ec8c08f8f56a796d012b))
* **helm:** generate a helm chart ([0e59139](https://github.com/grasse-oss/cron-set-controller/commit/0e59139e820581033abd171433493f886827a1c5))
* **informer:** watch dependent object Cronjob ([6c594f5](https://github.com/grasse-oss/cron-set-controller/commit/6c594f5491e8853c031dad637855f8a70a3f69c5))
* init project using operator-sdk ([fe40974](https://github.com/grasse-oss/cron-set-controller/commit/fe40974ecea9e7eacfe6af4ba35c391bb81bcffe))
* **node-event:** append node watcher to cronset controller setting ([1a59c1f](https://github.com/grasse-oss/cron-set-controller/commit/1a59c1f52b1def48c5b47aa5be14ef94f388c2b1))
* **node-event:** filter CronSet request using nodeSelector every node event ([6f7b16b](https://github.com/grasse-oss/cron-set-controller/commit/6f7b16bf29883f70905f48ebbdfdc0e91f06d233))
* run the make manifests command ([f6447b8](https://github.com/grasse-oss/cron-set-controller/commit/f6447b89b1516d11a060c2b55714a960c4753e81))
* support watchinf for node events ([29cf77c](https://github.com/grasse-oss/cron-set-controller/commit/29cf77c1fa07302e9f59e47017ea2bbd3e33eb10))


### Bug Fixes

* add node rbac to controller ([c4dc33e](https://github.com/grasse-oss/cron-set-controller/commit/c4dc33eca2a0a488257aa71b8c2b38a237b917e0))
* add node rbac to controller ([4a13c65](https://github.com/grasse-oss/cron-set-controller/commit/4a13c6518c8550a8bf78ac30e7e9871e6a953e03))
* **helm:** modify version to 1.0.0 ([1bcfd21](https://github.com/grasse-oss/cron-set-controller/commit/1bcfd2129dcf8932a3fbe83df28d5c121fb9cecf))
* **helm:** refine the helm chart ([f61205d](https://github.com/grasse-oss/cron-set-controller/commit/f61205d429f2ccc86a152ffce04cc912a2803b34))
* **helm:** use crds folder ([377d33f](https://github.com/grasse-oss/cron-set-controller/commit/377d33fb3bc61ab46b0a616b61e58e6cb7546c3d))
* leverage struct embedding ([8f327ef](https://github.com/grasse-oss/cron-set-controller/commit/8f327ef769a6ea1b3ba0fc76f5fc27de2dd74b9f))
* **log:** add key in reconcile log to resolve panic error ([08a8d2d](https://github.com/grasse-oss/cron-set-controller/commit/08a8d2d65c9f0797e53ee1d6e721d0751d7d0df9))
* modify the fields in the CronSetSpec ([2a645bb](https://github.com/grasse-oss/cron-set-controller/commit/2a645bb30e575ac47b42efb3669257fc9a33fdfa))
* **rbac:** add cronjobs permission to watch ([41b6ccc](https://github.com/grasse-oss/cron-set-controller/commit/41b6cccb650f58a01c49b4d592a2025da20689bd))
* Rename CronJobTemplate's json to cronJobTemplate ([e25fef5](https://github.com/grasse-oss/cron-set-controller/commit/e25fef597c71942718e365b9bd3fe7bd4f33bb01))
* **unit-test:** fix schema ([69b4787](https://github.com/grasse-oss/cron-set-controller/commit/69b478740104c8518684a5f45e243c3378bc6d8a))
