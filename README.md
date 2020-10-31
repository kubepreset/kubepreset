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

## Roadmap

### 0.2.0

- [ ] Update `ServiceBinding` resource's `.status.binding.name` with the secret name that is used binding
- [ ] Always create a [new secret for binding](https://github.com/k8s-service-bindings/spec/issues/101#issuecomment-717679053).
- [ ] Support `ServiceBinding` resource's `.spec.name` to override `.metadata.name` for directory name
- [ ] Support `ServiceBinding` resource's `.spec.type` to override value from the `ProvisionedService` secret
- [ ] Support `ServiceBinding` resource's `.spec.provider` to override value from the `ProvisionedService` secret

### 0.3.0

- [ ] Add support for label selectors for application
- [ ] Add support for specifying containers to inject (only name-based, and don't support index)
- [ ] Add support for environment variables

### 0.4.0

- [ ] Add support for Custom Projection extension
- [ ] Add support for Direct Secret Reference extension
- [ ] User manual
- [ ] Demo video

### pre-1.0

- [ ] Add support for mappings
- [ ] Add support for full spec except "Binding Secret Generation Strategies" extension

## Contributing to KubePreset

:+1::tada: First off, thanks for taking the time to contribute!
:tada::+1:

You can look at the issues [with help wanted label][help-wanted] for items that
you can work on.

If you need help, please feel free to [reach out to our discussion
group!][group]

When contributing to this repository, please first discuss the change you wish
to make via issue, email, or any other method with the owners of this repository
before making a change.  Small pull requests are easy to review and merge.  So,
please send small pull requests.

Please note we have a [code of conduct][conduct], please follow it in all your
interactions with the project.

Contributions to this project should conform to the [Developer Certificate of
Origin][dco].

Remember, when you send pull requests:

1. Write tests.
2. Write a [good commit message][commit-message].

See the [contribution guidelines][contribution] for more details.  The [KubePreset Wiki][wiki]
has additional information for contributors.

### Development

We recommend using GNU/Linux systems for the development of KubePreset. This
project requires the [Go version 1.14][go] or above installed in your
system. You also should have [make][make] and [GCC][gcc] installed in your
system.

To build the project:

    make

To run the tests:

    make test


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
[help-wanted]: https://github.com/kubepreset/kubepreset/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22
[conduct]: https://github.com/kubepreset/kubepreset/blob/main/CODE_OF_CONDUCT.md
[dco]: http://developercertificate.org
[commit-message]: https://chris.beams.io/posts/git-commit/
[wiki]: https://github.com/kubepreset/kubepreset/wiki
[go]: https://golang.org
[make]: https://en.wikipedia.org/wiki/Make_(software)
[gcc]: https://gcc.gnu.org
