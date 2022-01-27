
local ro = (import '../lib/rules-objstore.libsonnet')({
  name: 'example-rules-objstore',
  namespace: 'observatroium',
  version: 'main-2022-01-19-8650540',
  image: 'quay.io/observatorium/rules-objstore',
  replicas: 3,
  objectStorageConfig: {
    name: 'rules-objectstorage',
    key: 'objstore.yaml',
  },
  serviceMonitor: true,
});

{ [name]: ro[name] for name in std.objectFields(ro) }
