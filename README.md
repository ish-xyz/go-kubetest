# go-kubetest: Kubernetes integration tests

![go-kubetest logo](/assets/images/logo.png)

A tool to run integrations tests on kubernetes, by defining simple YAML files.<br>
Kubetest runs in cluster as a controller, executes integration tests periodically and exposes Prometheus metrics about itself and the tests results.
<br>
<br>

## Intro

With go-kubetest, integration tests can be created with one or multiple Kubernetes Resources. Let's take in consideration a test that tells us if we can or cannot create namespaces into a given Kubernetes cluster.

<br>

**Test Resource:**

```
apiVersion: go-kubetest.io/v1
kind: TestResource
metadata:
  name: namespaces
spec:
  data: |
    apiVersion: v1
    kind: Namespace
    metadata:
      labels:
        myCustomLabel: myCustomValue
      name: namespace-1
    spec: {}
    ---
    apiVersion: v1
    kind: Namespace
    metadata:
      labels:
        myCustomLabel: myCustomValue
      name: namespace-2
    spec: {}
```
<br>

**Test Definition:**

```
apiVersion: go-kubetest.io/v1
kind: TestDefinition
metadata:
  name: basic-tests
spec:
  - name: namespace_creation
    resources:
    - namespaces
    setup:
      waitFor:
      - resource: v1:Namespace:namespace-1
        timeout: 30s
      - resource: v1:Namespace:namespace-2
        timeout: 30s
    assert:
    - name: count_namespaces 
      type: expectedResources
      resource: v1:Namespace
      timeout: 120s
      count: 2
      selectors:
        status.phase: Active
        metadata.labels.myCustomLabel: myCustomValue
    teardown:
      waitFor:
      - resource: v1:Namespace:namespace-1
        timeout: 30s
      - resource: v1:Namespace:namespace-2
        timeout: 30s
```
<br>


In the above example the controller will try to create the resources within TestResource `namespaces` defined above, and wait for the resources specified in `setup.waitFor` for maximum `30s`.

At the moment the controller uses only 2 type of assertions.

The first one, "expectedResources", will try to retrieve one or more Kubernetes resources with the specified selectors (fieldSelectors and labelSelectors) and compare the number of resources fetched from the Kubernetes API, with the number of expected resources specified in `count`.

The second assertion, "expectedErrors", will count the number of errors during the creation of the resources specified in the `resources`, and compare it with the length of the array `errors`, if the number is the same it will check if the error/s match the regexes defined in the array of errors.

For example if you are expecting an error that says:
```
{something something} ... you need to specify SecurityContext ... {something something}
```
when creating a pod without securityContext, your "expectedErrors" assertion could look like:
```
- type: "expectedErrors"
  errors:
  - '.*SecurityContext.*'

```
So the controller will match 1 error with array length 1, and the regex in the element one of the array `errors` with the content of the error.

## Metrics

Kubetest run as a controller and exposes 4 simple metrics about the integration tests statuses.<br/>


| Metric Name                   | Type  | Description |
| :---                          | :---  | :---        |
| kubetest_test_status          | Gauge | The status of each integration test, 0 if it has failed and 1 if it has passed      |
| kubetest_total_tests          | Gauge | The total number of tests that the controller ran in the last execution             |
| kubetest_total_tests_passed   | Gauge | The total number of **passed** tests that the controller ran in the last execution  |
| kubetest_total_tests_failed   | Gauge | The total number of **failed** tests that the controller ran in the last execution  |
| kubetest_assertion_status     | Gauge | The status of each assertion, 0 if it has failed and 1 if it has passed             |


For some other practical examples see the `examples` folder.<br/>

## Step by step tutorial (TODO)

...