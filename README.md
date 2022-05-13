(The controller is not production ready)
# go-kubetest: Kubernetes integration tests

![go-kubetest logo](/assets/images/logo.png)
<br>

## ABSTRACT

Go-kubetest is a tool to run integrations tests on Kubernetes clusters by defining simple custom resources. 

Go-kubetest can run in 2 modes:

* As a controller, and in-cluster solution, running tests periodically and exposing metrics.

* As a oneshot process to run tests against a given cluster.

Go-kubetest comes with 3 CRDs: TestDefinition, TestResource and TestResult.

A user could run `kubectl get testresults` and quickly see how many tests have failed or passed, or run `kubectl get tests` to see which tests have been defined and deployed into a given namespace/cluster.

Go-kubetest is intended to be used to run integration tests or behaviour testing on Kubernetes only.
<br>
<br>

TODO: Better docs coming soon
