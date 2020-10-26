<img src="https://avatars0.githubusercontent.com/u/70762365" align="right" />

# KubePreset
> **Streamline Application Connectivity with Backing Services**

KubePreset is a [Kubernetes operator][operator] that aims to streamline
application connectivity with backing services.  KubePreset implements the
[Service Binding Specification for Kubernetes][spec].

**Disclaimer**: [KubePreset](https://kubepreset.dev) project is [my side project and not endorsed by my employer](https://www.redhat.com/en/about/open-source/participation-guidelines).  If you need any further clarity about it, please reach out to me directly.  -- [Baiju Muthukadan](https://twitter.com/baijum) (Creator of this project)

## Installation

This project is in the **Alpha** stage right now.  The recommended approach for installation is through [Helm charts][chart]

[Helm][helm] must be installed to use the charts.  Please refer to Helm's
[documentation][helm-docs] to get started.

Once Helm has been set up correctly, add the repo as follows:

```
helm repo add kubepreset https://kubepreset.github.io/helm-charts
```

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
kubepreset` to see the charts.

To install the `kubepreset` chart:

```
helm install my-kubepreset kubepreset/kubepreset
```

To uninstall the chart:

```
helm delete my-kubepreset
```


[operator]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[spec]: https://github.com/k8s-service-bindings/spec
[chart]: https://artifacthub.io/packages/helm/kubepreset/kubepreset
[helm]: https://helm.sh
[helm-docs]: https://helm.sh/docs/
