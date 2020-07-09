# grafana-tracing-demo

This is the code that I used to run a quick presentation for [London DevOps Exchange](https://www.meetup.com/DevOps-Exchange-London/events/271450322/) Meetup group on July 9th 2020

Good fun :smile:

Slides can be found [here](https://docs.google.com/presentation/d/1R9DWCmyc9TPBa6PQhOTFNPMl61gFQAnjZFQrkpug0y8/edit?usp=sharing)

## Folders

### app

Golang application that uses Open Telemetry for tracing and Prometheus instrumetation for metrics

This application will randomly return `500`, `200` or call itself to generate a "deeper" trace

The is available on [Dockerhub](https://hub.docker.com/repository/docker/carlosjgp/observability-demo-app)

## k8s-deployments

"values" files for Helm releases and a plain YAML Kubernetes deployment manifest
to provision all the required to run

## How to run this code

You will need to instal:
- [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://helm.sh/docs/intro/install/)

### Crate a k8s cluster

```shell
$ minikube start
  minikube v1.11.0 on Ubuntu 20.04
✨  Using the docker driver based on existing profile
👍  Starting control plane node minikube in cluster minikube
🔄  Restarting existing docker container for "minikube" ...
🐳  Preparing Kubernetes v1.18.3 on Docker 19.03.2 ...
    ▪ kubeadm.pod-network-cidr=10.244.0.0/16
🔎  Verifying Kubernetes components...
🌟  Enabled addons: default-storageclass, storage-provisioner
🏄  Done! kubectl is now configured to use "minikube"
```

### Add Helm repositories

- [Stable repository](https://github.com/helm/charts#how-do-i-enable-the-stable-repository-for-helm-3)
- [Grafana repository](https://github.com/grafana/loki/tree/master/production/helm#loki-helm-chart)

```shell
$ helm repo add stable https://kubernetes-charts.storage.googleapis.com
$ helm repo add loki https://grafana.github.io/loki/charts
$ helm repo up
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "loki" chart repository
...Successfully got an update from the "stable" chart repository
Update Complete. ⎈ Happy Helming!⎈ 
```

### Install the stack

```shell
$ ./install.sh
```

Now chill... Minikube is going to take some time to download all the images

Eventually you will get all the PODs up and running
```shell
NAMESPACE       NAME                                            READY   STATUS    RESTARTS   AGE
default         observability-demo-app-78bdc5f866-nxhk5         2/2     Running   0          7m39s
kube-system     coredns-66bff467f8-9k4h6                        1/1     Running   1          3h28m
kube-system     coredns-66bff467f8-hkc69                        1/1     Running   1          3h28m
kube-system     etcd-minikube                                   1/1     Running   0          8m12s
kube-system     kube-apiserver-minikube                         1/1     Running   0          8m12s
kube-system     kube-controller-manager-minikube                1/1     Running   1          3h29m
kube-system     kube-proxy-n6jz7                                1/1     Running   1          3h28m
kube-system     kube-scheduler-minikube                         1/1     Running   1          3h29m
kube-system     storage-provisioner                             1/1     Running   3          3h29m
observability   grafana-64bb799587-bztzz                        1/1     Running   1          3h27m
observability   grafana-image-renderer-79745fcc94-bthvl         1/1     Running   0          7m39s
observability   jaeger-operator-7db7cf477c-8dddc                1/1     Running   0          7m42s
observability   loki-0                                          1/1     Running   0          7m42s
observability   prometheus-kube-state-metrics-c65b87574-bll4t   1/1     Running   0          7m42s
observability   prometheus-server-849bff647d-lnbh6              2/2     Running   0          7m42s
observability   promtail-7d5lf                                  1/1     Running   0          7m42s
observability   simplest-77b8c6fc95-dvjsf                       1/1     Running   0          6m23s
```

You can now let the `observability-demo-app` run for a little bit...
the POD has a sidecar that sends a request every second.
That will generate some traffic, metrics and traces.

### Take a look at Grafana 7

```shell
$ ./port-forward.sh 
Prometheus:   http://localhost:9090
Jaeger:       http://localhost:16686
Loki:         http://localhost:3100
Grafana:      http://localhost:3000
Run 'pkill kubectl' to clean up
```

Open your browser on any of the addresses above to see the UIs.
Well... Loki does not have a UI but you could use it's cool CLI tool
[LogCLI](https://github.com/grafana/loki/blob/master/docs/getting-started/logcli.md) and learn about [LogQL](https://github.com/grafana/loki/blob/master/docs/logql.md) language

```shell
$ export LOKI_ADDR=http://localhost:3100
$ logcli query '{pod=~"observability-.*"} |= "level=error"' --tail
```

### Clean up

:fire::fire::fire:

```shell
$ minikube delete
```