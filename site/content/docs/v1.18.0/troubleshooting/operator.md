# Operator Troubleshooting

[Sesame Operator][1] runs in a Kubernetes cluster and is managed by a
[Deployment][2] resource. The following steps can be used to verify and
troubleshoot the operator:

Verify the operator deployment is available:
```bash
$ kubectl get deploy/sesame-operator -n sesame-operator
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
sesame-operator   1/1     1            1           39m
```

Check the logs of the operator:
```bash
$ kubectl logs deploy/sesame-operator -n sesame-operator -c sesame-operator -f
2020-12-01T00:10:14.245Z	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": "127.0.0.1:8080"}
2020-12-01T00:10:14.341Z	INFO	setup	starting sesame-operator
I1201 00:10:14.343439       1 leaderelection.go:243] attempting to acquire leader lease  sesame-operator/0d879e31.projectsesame.io...
2020-12-01T00:10:14.345Z	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
I1201 00:10:14.442755       1 leaderelection.go:253] successfully acquired lease sesame-operator/0d879e31.projectsesame.io
2020-12-01T00:10:14.447Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:14.540Z	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"ConfigMap","namespace":"Sesame-operator","name":"0d879e31.projectsesame.io","uid":"40ebcf47-a105-4efc-b7e9-993df2070a07","apiVersion":"v1","resourceVersion":"33899"}, "reason": "LeaderElection", "message": "Sesame-operator-55d6c7b4b5-hkd78_e960056c-f959-45c9-8686-9b85e09e21d8 became leader"}
2020-12-01T00:10:14.842Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:15.046Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:15.340Z	INFO	controller	Starting Controller	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame"}
2020-12-01T00:10:15.341Z	INFO	controller	Starting workers	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "worker count": 1}
...
```

When a `Sesame` is created, the operator should successfully reconcile the object:
```bash
...
2020-12-01T00:10:15.341Z	INFO	controllers.Sesame	reconciling	{"request": "default/Sesame-sample"}
...
2020-12-01T00:10:15.442Z	DEBUG	controller	Successfully Reconciled	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "name": "Sesame-sample", "namespace": "default"}
```

Check the status of your `Sesame` resource:
```bash
$ kubectl get sesame/sesame-sample -o yaml
apiVersion: operator.projectsesame.io/v1alpha1
kind: Sesame
metadata:
  name: sesame-sample
  namespace: default
...
status:
  availableSesames: 2
  availableEnvoys: 1
  conditions:
  - lastTransitionTime: "2020-12-01T00:55:38Z"
    message: Sesame has minimum availability.
    reason: SesameAvailable
    status: "True"
    type: Available
```

If the `Sesame` does not become available, check the status of operands.
```bash
$ kubectl get po -n projectsesame
NAME                       READY   STATUS      RESTARTS   AGE
sesame-7649c6f6cc-9hxn9   1/1     Running     0          6m18s
sesame-7649c6f6cc-sb4nn   1/1     Running     0          6m18s
sesame-certgen-vcsqd      0/1     Completed   0          6m18s
envoy-qshxf                2/2     Running     0          6m18s
```

Check the logs of the operands. The following example checks the logs of the
Sesame deployment operand:
```bash
$ kubectl logs deploy/sesame -n projectsesame -c sesame -f
2020-12-01T00:10:14.245Z	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": "127.0.0.1:8080"}
2020-12-01T00:10:14.341Z	INFO	setup	starting sesame-operator
I1201 00:10:14.343439       1 leaderelection.go:243] attempting to acquire leader lease  sesame-operator/0d879e31.projectsesame.io...
2020-12-01T00:10:14.345Z	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
I1201 00:10:14.442755       1 leaderelection.go:253] successfully acquired lease sesame-operator/0d879e31.projectsesame.io
2020-12-01T00:10:14.447Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:14.540Z	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"ConfigMap","namespace":"Sesame-operator","name":"0d879e31.projectsesame.io","uid":"40ebcf47-a105-4efc-b7e9-993df2070a07","apiVersion":"v1","resourceVersion":"33899"}, "reason": "LeaderElection", "message": "Sesame-operator-55d6c7b4b5-hkd78_e960056c-f959-45c9-8686-9b85e09e21d8 became leader"}
2020-12-01T00:10:14.842Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:15.046Z	INFO	controller	Starting EventSource	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "source": "kind source: /, Kind="}
2020-12-01T00:10:15.340Z	INFO	controller	Starting Controller	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame"}
2020-12-01T00:10:15.341Z	INFO	controller	Starting workers	{"reconcilerGroup": "operator.projectsesame.io", "reconcilerKind": "Sesame", "controller": "Sesame", "worker count": 1}
...
```

__Note:__ `projectsesame` is the default namespace used by operands of a `Sesame`
when `spec.namespace.name` is unspecified.
 
[1]: https://github.com/projectsesame/sesame-operator
[2]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
