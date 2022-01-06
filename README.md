# go-kubetest: Kubernetes integration tests

![go-kubetest logo](/assets/images/logo.png)
<br>

## Abstract

Go-kubetest is a tool to run integrations tests on kubernetes clusters, by defining simple custom resources. Go-kubetest can run in 2 modes:

As a controller and in-cluster solution, running tests periodically and providing metrics about the tests results.

As a oneshot process to run tests against a given cluster.

Go-kubetest comes with 3 CRDs: TestDefinition, TestResource and TestResult.

A user could run `kubectl get testresults` and quickly see how many tests have failed or successed, or run `kubectl get tests` to see which tests have been defined and deployed into a given namespace/cluster.

Go-kubetest is intended to be used to run integration tests or behaviour testing on Kubernetes only.


### Test Definition

As mentioned above TestDefinition is a CRD. It has 3 required "fields": setup, teardown, assert. Although, the fields are required, they can be empty. This is what I have defined as "soft test". Let me explain with an example.

The following is a TestDefinition (soft test) that **only** ensures that a set of resources are present within a Kubernetes cluster:

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
As you can see in the above spec **setup** and **teardown** are defined as empty dicts, this means that if we run the go-kubetest controller, it won't create or delete any resource from the cluster, it will just ensure that the ones defined in the **assert** field exist.

On the other hand, sometimes we might want to run more complex tests that create, delete, and ensure that things are running in our clusters, this is what I have defined as **hard tests**. Let's take in consideration an example of hard test:

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

We will explain what a TestResource is in the next paragraph, but for now think of it as plain Kubernetes manifests that can be applied or deleted.

In the above example, the TestDefinition will tell the controller to create the resources defined in "namespaces" a TestResource.
The controller will then wait for some resources (setup.WaitFor) to be created, run some assertions and ensure that the namespaces have been created. Finally, it will delete the resources defined in teardown (teardown.WaitFor).

Obviously, this mechanic can be applied to **any** resource in a given Kubernetes cluster, and not only namespaces.


**Assertions Types**:

At the moment there are only 2 types of assertion: expectedResources and expectedErrors.

(TODO)



### Test Resource:

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

(TODO)

### Test Result:

(TODO)

## Tutorial:
(TODO)

### Additional Options:

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
