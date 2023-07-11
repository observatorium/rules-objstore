// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  name: error 'must provide name',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide image',
  imagePullPolicy: 'IfNotPresent',
  replicas: error 'must provide replicas',
  logLevel: 'info',
  logFormat: 'logfmt',
  ports: {
    public: 8080,
    internal: 8081,
  },
  resources: {},
  serviceMonitor: false,
  objectStorageConfig: error 'must provide objectStorageConfig',

  commonLabels:: {
    'app.kubernetes.io/name': 'rules-objstore',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'rules-storage',
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local ro = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,
  // Safety checks for combined config of defaults and params
  assert std.isNumber(ro.config.replicas) && ro.config.replicas >= 0 : 'rules-objstore replicas has to be number >= 0',
  assert std.isObject(ro.config.resources),
  assert std.isBoolean(ro.config.serviceMonitor),
  assert std.isObject(ro.config.objectStorageConfig),

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: ro.config.name,
      namespace: ro.config.namespace,
      labels: ro.config.commonLabels,
    },
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: ro.config.name,
      namespace: ro.config.namespace,
      labels: ro.config.commonLabels,
    },
    spec: {
      selector: ro.config.podLabelSelector,
      ports: [
        {
          name: name,
          port: ro.config.ports[name],
          targetPort: ro.config.ports[name],
        }
        for name in std.objectFields(ro.config.ports)
      ],
    },
  },

  deployment: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name: ro.config.name,
      namespace: ro.config.namespace,
      labels: ro.config.commonLabels,
    },
    spec: {
      replicas: ro.config.replicas,
      selector: { matchLabels: ro.config.podLabelSelector },
      strategy: {
        rollingUpdate: {
          maxSurge: 0,
          maxUnavailable: 1,
        },
      },
      template: {
        metadata: { labels: ro.config.commonLabels },
        spec: {
            affinity: {
              podAntiAffinity: {
                preferredDuringSchedulingIgnoredDuringExecution: [
                  {
                    weight: 100,
                    podAffinityTerm: {
                      labelSelector: {
                        matchExpressions: [
                          {
                            key: 'app.kubernetes.io/name',
                            operator: 'In',
                            values: [
                              'rules-objstore',
                            ],
                          },
                        ],
                      },
                      topologyKey: 'kubernetes.io/hostname',
                    },
                  },
                ],
              },
            },
          serviceAccountName: ro.serviceAccount.metadata.name,
          containers: [
            {
              name: 'rules-objstore',
              image: ro.config.image,
              imagePullPolicy: ro.config.imagePullPolicy,
              args: [
                '--debug.name=%s' % ro.config.name,
                '--web.listen=0.0.0.0:%d' % ro.config.ports.public,
                '--web.internal.listen=0.0.0.0:%d' % ro.config.ports.internal,
                '--web.healthchecks.url=http://localhost:%d' % ro.config.ports.public,
                '--log.level=%s' % ro.config.logLevel,
                '--log.format=%s' % ro.config.logFormat,
                '--objstore.config-file=/etc/rules-objstore/%s' % ro.config.objectStorageConfig.key,
              ],
              ports: [
                { name: name, containerPort: ro.config.ports[name] }
                for name in std.objectFields(ro.config.ports)
              ],
              resources: if ro.config.resources != {} then ro.config.resources else {},
              livenessProbe: {
                failureThreshold: 10,
                periodSeconds: 30,
                httpGet: {
                  path: '/live',
                  port: ro.config.ports.internal,
                  scheme: 'HTTP',
                },
              },
              readinessProbe: {
                failureThreshold: 12,
                periodSeconds: 5,
                httpGet: {
                  path: '/ready',
                  port: ro.config.ports.internal,
                  scheme: 'HTTP',
                },
              },
              volumeMounts: [{
                name: 'objstore',
                mountPath: '/etc/rules-objstore/%s' % ro.config.objectStorageConfig.key,
                subPath: ro.config.objectStorageConfig.key,
                readOnly: true,
              }],
            },
          ],
          volumes: [{
            secret: {
              secretName: ro.config.objectStorageConfig.name
            },
            name: 'objstore',
          }],
        },
      },
    },
  },

  serviceMonitor: if ro.config.serviceMonitor == true then {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata+: {
      name: ro.config.name,
      namespace: ro.config.namespace,
    },
    spec: {
      selector: {
        matchLabels: ro.config.commonLabels,
      },
      endpoints: [{ port: 'internal' }],
    },
  },
}
