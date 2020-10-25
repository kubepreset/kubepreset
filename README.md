<img src="https://avatars0.githubusercontent.com/u/70762365" align="right" />

# KubePreset
> Streamline Application Connectivity with Backing Services

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

**Disclaimer**: [KubePreset](https://kubepreset.dev) project is [my side project and not endorsed by my employer](https://www.redhat.com/en/about/open-source/participation-guidelines).  If you need any further clarity about it, please reach out to me directly.  -- [Baiju Muthukadan](https://twitter.com/baijum) (Creator of this project)

This project is in the **Alpha** stage right now.  The recommended approach for installation is through [Helm charts][chart]

---
This project is maintained by [KubePreset Team (kubepreset@googlegroups.com)][group]

[operator]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[spec]: https://github.com/k8s-service-bindings/spec
[group]: https://groups.google.com/g/kubepreset
[chart]: https://artifacthub.io/packages/helm/kubepreset/kubepreset
