# go-kubetest: Kubernetes integration tests

![go-kubetest logo](/assets/images/logo.png)
<br>

## ABSTRACT

Go-kubetest is a tool to run integrations tests on Kubernetes clusters by defining simple custom resources. 

Go-kubetest can run in 2 modes:

* As a controller, and in-cluster solution, running tests periodically and exposing metrics.

* As a oneshot process to run tests against a given cluster.

Go-kubetest comes with 3 CRDs: TestDefinition, TestResource and TestResult.

A user could run `kubectl get testresults` and quickly see how many tests have failed or successed, or run `kubectl get tests` to see which tests have been defined and deployed into a given namespace/cluster.

Go-kubetest is intended to be used to run integration tests or behaviour testing on Kubernetes only.
<br>
<br>

### TEST DEFINITION

As mentioned above TestDefinition is a CRD. It has 3 required fields: setup, teardown, assert.<br>
Although, the fields are required, they can be empty. This is what I have defined as "soft test".<br>
A "soft test" is a type of integration tests that only verifies that certain resources are present in the cluster and it does not create or delete any Kubernetes resource.

Let's see the following example.s

The manifest below shows a TestDefinition that **only** ensures that a set of resources are present within a Kubernetes cluster:

```
apiVersion: go-kubetest.io/v1
kind: TestDefinition
metadata:
  name: soft-tests
  labels:
    type: soft
spec:
  - name: ensure-resources-exist
    resources: []
    setup: {}
    teardown: {}
    assert:
    - name: namespace
      type: expectedResources
      resource: v1:Namespace
      timeout: 10s
      count: 1
      selectors:
        status.phase: Active
        metadata.name: myCustomNamespace
    - name: deployment
      type: expectedResources
      resource: apps/v1:Deployment:nginx-controller
      timeout: 10s
      count: 1
      selectors:
        metadata.name: nginx-controller
    - name: daemonset
      type: expectedResources
      resource: apps/v1:DaemonSet:kube-system
      timeout: 10s
      count: 1
      selectors:
        metadata.name: kube-proxy
```
As you can see in the above spec **setup** and **teardown** are defined as empty dicts, this means that if we run the go-kubetest controller, it won't create or delete any resources from the cluster, it will just run the assertions defined under `assert:`.
<br>
<br>

On the other hand, sometimes we might want to run more complex tests that create, delete, and ensure that things are running in our clusters.
<br>

Let's say that we have a policy in our clusters that only allows namespaces with a label called `metadata.labels.myCustomLabel` to be created.
<br>

We want to test that the policy always work.

To test that, we can define a test that will tell the go-kubetest controller to create 2 namespaces, wait for creation, try to fetch them with a label selector `metadata.labels.myCustomLabel: myCustomValue`, verify that the number of expected namespaces match the number of fetched namespaces and teardown (delete) the namespaces. 
<br>

The following definition does just what I have described:

```
apiVersion: go-kubetest.io/v1
kind: TestDefinition
metadata:
  name: basic-tests
  labels:
    type: hard
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

We will explain what a TestResource is in the next paragraph, but for now think of it as plain Kubernetes manifests that can be applied or deleted.

Obviously, this mechanism can be applied to **any** resource in a given Kubernetes cluster, and not only namespaces.
<br><br>

**ASSERTION TYPES**:
<br>

At the moment I have implemented only 2 types of assertion: *expectedResources* and *expectedErrors*.

**expectedResources** will use the `spec.assert.[*].selectors` and the `spec.assert.[*].resource` defined to try to fetch the resources from the Kubernetes API. 
If the `spec.assert.[*].count` value is equal to the number of resources fetched from the Kubernetes API it will mark the assertion as passed.
It will try to fetch the resources every 2 seconds until the `spec.assert.[*].timeout` defined has expired.

The resource path can be used as follow, with `:` as delimiter.

For namespaced resources -> ($group/$version):$kind:$namespace (e.g.: apps/v1:Deployment/default)
<br>
For cluster-wide resources ->  ($group/$version):$kind (e.g.: v1:Namespace)
<br>
<br>

**expectedErrors** only works during setup time. The controller will store every error it encounters while provisioning the resources, count the number of errors, and check if the lenght of `spec.assert.[*].errors` (defined in the spec, see below) matches the number of errors stored by the controller.
However, only counting the errors is not enough. The controller will then take the regexes defined in `spec.assert.[*].errors` and match those with the errors returned by the Kubernetes API. Following an example:

```
apiVersion: go-kubetest.io/v1
kind: TestDefinition
metadata:
  name: basic-tests
  labels:
    type: hard
spec:
  - name: pods_creation
    resources:
    - pod
    setup:
      waitFor:
      - resource: v1:Pod:default:my-pod
        timeout: 30s

    assert:
    - name: pods-errors
      type: expectedErrors
      errors:
      - '.*SecurityContext.*'

    - name: pods-resources
      type: expectedResources
      resource: v1:Pod:default
      timeout: 30s
      count: 0
      selectors:
        metadata.name: mypod

    teardown:
      waitFor:
      - resource: v1:Pod:default:my-pod
        timeout: 30s
```

Consider the above example. 
Let's say we have an OPA policy that doesn't permit to create pods without securityContext. If we try, the kubernetes API would return an error similar to:

```
{something something} ... you need to specify SecurityContext ... {something something}
```

So we say to go-kubetest to try to create the resource `pod`, which does NOT have a securityContext, and to expect an error, only 1, that matches the regex `'.*SecurityContext.*'`. 
As second assertion, we say that go-kubetest should try to fetch the resource pod and it should expect 0 resources. This way we ensure that the OPA policy is working and it's denying our pod creation.
<br>
<br>


### TEST RESOURCE:

A TestResource it's a CRD similar to ConfigMap, the only supported field is `data:` and within data you can specify one or more manifests.

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
    status: {}
    ---
    apiVersion: v1
    kind: Namespace
    metadata:
      labels:
        myCustomLabel: myCustomValue
      name: namespace-2
    spec: {}
    status: {}
```

The above example shows a TestResource that, if specified in a test, will be used to create the 2 namespaces: namespace-1 and namespace-2.
<br>
<br>

### TEST RESULT:

TestResult is a resource that will show you the results of the tests. See below:

```
apiVersion: go-kubetest.io/v1
kind: TestResult
metadata:
  name: name-of-the-test
spec:
  result: false
  assertion-1: true
  assertion-2: true
  assertion-3: false
```
<br>
<br>

## DEMO:

**ONE SHOT MODE**:

![terminal](./assets/images/terminal.gif)

Video link: https://asciinema.org/a/Pwy2eq8j5rGlbAoMrETOJIScG

**NOTE**: The *controller mode* it runs as the *oneshot mode* with the only difference that exposes metrics and it runs forever with intervals.
<br>
<br>

## METRICS

Kubetest run as a controller and exposes 4 simple metrics about the integration tests statuses.<br/>

| Metric Name                   | Type  | Description |
| :---                          | :---  | :---        |
| kubetest_test_status          | Gauge | The status of each integration test, 0 if it has failed and 1 if it has passed      |
| kubetest_total_tests          | Gauge | The total number of tests that the controller ran in the last execution             |
| kubetest_total_tests_passed   | Gauge | The total number of **passed** tests that the controller ran in the last execution  |
| kubetest_total_tests_failed   | Gauge | The total number of **failed** tests that the controller ran in the last execution  |
| kubetest_assertion_status     | Gauge | The status of each assertion, 0 if it has failed and 1 if it has passed             |
<br>
<br>

## OTHER REFERENCES:

[Examples](./examples)
<br>

[CLI Reference](./docs/cli-reference.md)
<br>

[Design Document](./docs/design.md)
<br>
