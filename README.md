<img src="https://avatars0.githubusercontent.com/u/70762365" align="right" />

# KubePreset
> **Streamline Application Connectivity with Backing Services**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release Charts](https://github.com/kubepreset/helm-charts/workflows/Release%20Charts/badge.svg)](https://github.com/kubepreset/helm-charts/actions)
[![Build](https://github.com/kubepreset/kubepreset/workflows/Build/badge.svg?branch=main)](https://github.com/kubepreset/kubepreset/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubepreset/kubepreset)](https://goreportcard.com/report/github.com/kubepreset/kubepreset)
[![codecov](https://codecov.io/gh/kubepreset/kubepreset/branch/main/graph/badge.svg)](https://codecov.io/gh/kubepreset/kubepreset)
[![go.dev Reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/mod/github.com/kubepreset/kubepreset)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/kubepreset)](https://artifacthub.io/packages/search?repo=kubepreset)
[![Docker Repository on Quay](https://quay.io/repository/kubepreset/kubepreset/status "Docker Repository on Quay")](https://quay.io/repository/kubepreset/kubepreset)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/kubepreset/kubepreset/blob/main/CONTRIBUTING.md)

KubePreset is a [Kubernetes operator][operator] that aims to streamline
application connectivity with backing services.  KubePreset implements the
[Service Binding Specification for Kubernetes][spec].

From the [spec][spec] introduction:

> Today in Kubernetes, the exposure of secrets for connecting applications to external services such as REST APIs, databases, event buses, and many more is manual and bespoke.  Each service provider suggests a different way to access their secrets, and each application developer consumes those secrets in a custom way to their applications.  While there is a good deal of value to this flexibility level, large development teams lose overall velocity dealing with each unique solution.  To combat this, we already see teams adopting internal patterns for how to achieve this application-to-service linkage.
>
> This specification aims to create a Kubernetes-wide specification for communicating service secrets to applications in an automated way.  It aims to create a widely applicable mechanism but _without_ excluding other strategies for systems that it does not fit easily.  The benefit of Kubernetes-wide specification is that all of the actors in an ecosystem can work towards a clearly defined abstraction at the edge of their expertise and depend on other parties to complete the chain.
>
> * Application Developers expect their secrets to be exposed consistently and predictably.
> * Service Providers expect their secrets to be collected and exposed to users consistently and predictably.
> * Platforms expect to retrieve secrets from Service Providers and expose them to Application Developers consistently and predictably.

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

## Community

We have a [mailing list (kubepreset@googlegroups.com)][group] for community
support and discussion.  You are welcome to ask any questions about KubePreset.

To report any issues, use our [primary GitHub issue tracker][tracker].  You can
make feature requests and report bugs.  For reporting any security issues, see
the [security policy page][security-policy].

You are welcome to contribute code and documentation to this project. See the
[contribution guidelines][contribution] for more details.

If you are a backing service creator and want to make your service accessible
through KubePreset to the application developer, [see our
documentation][backing-service].

If you are an application developer or a Kubernetes cluster administrator, [read
the documentation][application-developer] to understand how to connect your
application to a service using KubePreset.

---
This project is maintained by [KubePreset Team (kubepreset@googlegroups.com)][group]

[operator]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[spec]: https://github.com/k8s-service-bindings/spec
[group]: https://groups.google.com/g/kubepreset
[chart]: https://artifacthub.io/packages/helm/kubepreset/kubepreset
[helm]: https://helm.sh
[helm-docs]: https://helm.sh/docs/
[tracker]: https://github.com/kubepreset/kubepreset/issues
[security-policy]: https://github.com/kubepreset/kubepreset/blob/main/SECURITY.md
[contribution]: https://github.com/kubepreset/kubepreset/blob/main/CONTRIBUTING.md
[backing-service]: https://kubepreset.dev
[application-developer]: https://kubepreset.dev
