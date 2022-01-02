# go-kubetest: Kubernetes integration tests

![go-kubetest logo](/assets/images/logo.png)

A tool to run integrations tests on kubernetes, by defining simple YAML files.<br>
Kubetest runs in cluster as a controller, executes integration tests periodically and exposes Prometheus metrics about itself and the tests results.
<br>
<br>

## Intro

With go-kubetest, integration tests can be created with one or multiple YAML file. Let's take in consideration a test that tells us if we can or cannot create namespaces into a given Kubernetes cluster.

```
test-1.yaml:

- name: namespace_creation
  resources:
  - ./manifests/test-1/namespace.yaml
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

In the above example the controller will try to create the resources within `./manifest/test-1/namespace.yaml` and wait for the resources specified in `setup.waitFor` for maximum `15s`.

At the moment the controller uses only 2 type of assertions.

The first one, "expectedResources", will try to retrieve one or more Kubernetes resources with the specified selectors (fieldSelectors and labelSelectors) and compare the number of resources fetched from the Kubernetes API, with the number of expected resources specified in `count`.

The second assertion, "expectedErrors", will count the number of errors during the creation of the resources specified in the `manifest`, and compare it with the length of the array `errors`, if the number is the same it will check if the error/s are match the regexes defined  in the array of errors.

For example if you are expecting an error that says:
```
{something something} ... you need to specify SecurityContext ... {something something}
```
when creating a pod without securityContext, your expectedErrors could look like:
```
- type: "expectedErrors"
  errors:
  - '.*SecurityContext.*'

```
So the controller will match 1 error with array length 1, and the regex in the element one of the array `errors` with the content of the error.

## Metrics

Kubetest run as a controller and exposes 4 simple metrics about the integration tests statuses.<br/>


| Metric Name                   | Type  | Description |
| :---                          | :---  | :---  |
| kubetest_test_status          | Gauge | The status of each integration test, 0 if it has failed and 1 if it has passed |
| kubetest_total_tests          | Gauge | The total number of tests that the controller ran in the last execution |
| kubetest_total_tests_passed   | Gauge | The total number of **passed** tests that the controller ran in the last execution |
| kubetest_total_tests_failed   | Gauge | The total number of **failed** tests that the controller ran in the last execution        |


For some other practical examples see the `examples` folder.<br/>

## Step by step tutorial

1. Run the following commands:<br/>
```
mkdir -p testsdir/manifests
touch testsdir/test-1.yaml
```

2. Edit the file `testsdir/test-1.yaml` with:<br/>
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

3. Download go-kubetest and test it locally against your Kubernetes cluster:<br/>

```
./go-kubetest --kubeconfig ~/.kube/config --testsdir examples --interval 60 --debug
```
