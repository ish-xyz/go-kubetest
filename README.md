# KUBETEST: Kubernetes integration tests

A tool to run integrations tests on kubernetes, by defining simple YAML files.

## Intro

With go-kubetest tests can be defined simply by creating one or more yaml files, with the following syntax:

```
- name: namespace_creation
  manifest: ./manifests/test-1.yaml
  setup:
    waitFor:
    - resource: v1/Namespace/namespace-1
      timeout: 15s
  assert:
  - type: expectedResources
    apiVersion: v1
    kind: Namespace
    count: 1
    selectors:
      metadata.name: namespace-1
      status.phase: Active
      metadata.labels.myCustomLabel: myCustomValue
  - type: expectedErrors
    errors: []
  teardown:
    waitFor:
    - resource: v1/Namespace/namespace-1
      timeout: 10s
```

In the above example the controller will try to create the resources within `./manifest/test-1.yaml` and wait for the resources specified in `setup.waitFor` for a maximum amount of time of `15s`.

At the moment the controller uses only 2 assertions.

The first one, "expectedResources", will try to retrieve one or more Kubernetes resources with the specified selectors (fieldSelectors and labelSelectors). Right after, will compare the number taken from the field `count` (expected number of resources) with the number of actual objects retrieved from the cluster, if it doesn't match the assertion has failed.

The second assertion, "expectedErrors", will count the errors during the creation of the resources and match the number of error and the regex defined in the test definition with the errors found during the setup. An empty array is considered as 0 errors expected.

## Metrics

Kubetest run as a controller and exposes 4 simple metrics about the integration tests statuses.

| Metric Name                   | Type  | Description |
| :---                          | :---  | :---  |
| kubetest_test_status          | Gauge | The status of each integration test, 0 if it has failed and 1 if it has passed |
| kubetest_total_tests          | Gauge | The total number of tests that the controller ran in the last execution |
| kubetest_total_tests_passed   | Gauge | The total number of **passed** tests that the controller ran in the last execution |
| kubetest_total_tests_failed   | Gauge | The total number of **failed** tests that the controller ran in the last execution        |


For some other practical examples see the `examples` folder.

## Step by step tutorial

1. Run the following commands:
```
mkdir -p testsdir/manifests
touch testsdir/test-1.yaml
```

2. Edit the file `testsdir/test-1.yaml` with:
```
- name: namespace_creation
  manifest: ./manifests/test-1.yaml
  setup:
    waitFor:
    - resource: v1/Namespace/namespace-1
      timeout: 15s
  assert:
  - type: expectedResources
    apiVersion: v1
    kind: Namespace
    count: 1
    selectors:
      metadata.name: namespace-1
      status.phase: Active
      metadata.labels.myCustomLabel: myCustomValue
  - type: expectedErrors
    errors: []
  teardown:
    waitFor:
    - resource: v1/Namespace/namespace-1
      timeout: 10s
```

3. Download go-kubetest and test it locally against your Kubernetes cluster:

```
./go-kubetest --kubeconfig ~/.kube/config --testsdir examples --interval 60 --debug
```
